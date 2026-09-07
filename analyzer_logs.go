// analyzer_logs.go
package main

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// logFileNames lists the files to scan for SSSD log messages
var logFileNames = []string{"sssd.txt", "messages", "messages.txt"}

// analyzeSSSDConfigAndLogs orchestrates the full analysis of SSSD configuration and logs.
// It loads sssd.conf, scans log files for known error patterns, builds a timeline,
// and checks for common configuration misconfigurations.
func analyzeSSSDConfigAndLogs(dirPath string, report *ReportData) {
	sssdConfContent := readFileSafe(dirPath, "sssd.conf")
	if sssdConfContent == "" {
		sssdConfContent = extractSection(dirPath, "sssd.txt", "# /etc/sssd/sssd.conf")
	}

	// Quick checks for known issues that map to report.Problems/Warnings directly
	if anyFileContains(dirPath, logFileNames, "User account has expired") || anyFileContains(dirPath, logFileNames, "Clients credentials have been revoked") {
		report.Problems = append(report.Problems, "[AUTHENTICATION] Logs indicate an Active Directory user account is expired, locked, or credentials have been revoked.")
	}
	if anyFileContains(dirPath, logFileNames, "terminated by own WATCHDOG") {
		report.Warnings = append(report.Warnings, "[TUNING] Since a WATCHDOG termination was found, consider setting 'ignore_group_members = true' in sssd.conf to speed up ssh/sudo initial lookups.")
	}
	if anyFileContains(dirPath, logFileNames, "service key not available") || anyFileContains(dirPath, logFileNames, "TGT failed verification") || anyFileContains(dirPath, logFileNames, "KDC has no support for encryption type") {
		report.Warnings = append(report.Warnings, "[AD CRYPTO BUG] Crypto mismatch or 'service key not available' detected. Microsoft AD forces deprecated RC4 encryption if the 'operatingSystemVersion' attribute in AD starts with a number less than 6 (e.g., '5.14.21'). If your Linux crypto-policy disables RC4, authentication will fail. Fix: Prepend the AD attribute with 'Linux ' (e.g., 'Linux 5.14'), OR re-enable RC4 on this host using 'update-crypto-policies --set DEFAULT:AD-SUPPORT' and reboot.")
	}

	errorPatterns := buildErrorPatterns(report.MACType)
	detectedLogErrors, timelineEvents := scanAndCollectErrors(dirPath, errorPatterns)

	report.Timeline = timelineEvents
	report.SSSDLogErrors = buildSortedLogErrors(detectedLogErrors)

	if sssdConfContent != "" {
		report.SssdConfigFound = true
		report.SSSDConfigSnippet = sssdConfContent
		// Phase A engine (same as analyzeData], instead of the removed legacy
		// substring-based analyzeSSSDConfig: typed parser + whole-config duplicate
		// detection + AD-specific option validation.

		cfg := parseSssdConfig(sssdConfContent)
		validateDuplicateKeys(cfg, report)
		validateADConfig(cfg, report)
	} else {
		report.Problems = append(report.Problems, "sssd.conf or SSSD configuration block not found in the supportconfig.")
	}

	if report.SssdService != "Running" {
		report.Problems = append(report.Problems, "sssd.service is not actively running.")
	}
}

// buildErrorPatterns returns the comprehensive SSSD error pattern dictionary,
// filtering out SELinux-specific noise unless the system uses SELinux.
func buildErrorPatterns(macType string) map[string]string {
	errorPatterns := map[string]string{
		"KRB5_KDC_ERR_C_PRINCIPAL_UNKNOWN":       "Kerberos: Client principal unknown (Machine account missing or deleted in AD)",
		"Preauthentication failed":               "Kerberos: Preauthentication failed (Bad keytab or password mismatch)",
		"Constraint violation":                   "LDAP: Constraint violation (AD Policy restriction)",
		"Clock skew too great":                   "Kerberos: Clock skew too great (Time synchronization failure with DC)",
		"Could not start TLS encryption":         "LDAP: Could not start TLS encryption (Certificate or protocol issue)",
		"Invalid credentials":                    "LDAP: Invalid credentials (Bad bind DN or password)",
		"KDC has no support for encryption type": "Kerberos: KDC has no support for encryption type (Crypto policy mismatch)",
		"Server not found in Kerberos database":  "Kerberos: Server not found in database (SPN missing)",
		"terminated by own WATCHDOG":             "SSSD Watchdog: Process blocked for too long (Slow DNS, slow AD, or heavy load)",
		"sdap_async_sys_connect request failed":  "LDAP Network Error: Connection to AD/LDAP timed out",
		"krb5_child_timeout":                     "Kerberos Timeout: krb5_child_timeout reached (KDC slow, distant, or firewalled)",
		"tsig verify failure":                    "DNS Update Failed: TSIG verify failure (Dynamic DNS update rejected)",
		"SSSD is offline":                        "SSSD Offline: Provider forced offline (Network loss, DNS failure, or firewall)",
		"ldap_extended_operation failed":         "LDAP/IPA Error: Extended operation failed (Directory server busy or unreachable)",
		"ldap_install_tls failed":                "TLS Handshake Failed: ldap_install_tls failed (Certificate mismatch or bad CA)",
		"ldap_sasl_interactive_bind_s failed":    "LDAP SASL Bind Failed: Could not negotiate SASL bind (Crypto mismatch or unreachable KDC)",
		"database disk image is malformed":       "SSSD Cache Corrupt: LDB database image is malformed (Recommend running: rm -f /var/lib/sss/db/* && systemctl restart sssd)",
		"Machine account password expired":       "Kerberos: Machine account password expired (AD dropped trust, needs rejoin)",
		"Message stream modified":                "Kerberos/LDAP: Message stream modified (Often indicates an expired machine password or bad keytab)",
		"service key not available":              "Kerberos: Service key not available (KDC forced RC4 due to AD OS version bug)",
		"TGT failed verification using key":      "Kerberos: TGT validation failed (Crypto mismatch between AD and keytab)",

		"Section [domain/] is not allowed":                           "Config Error: Empty or invalid domain section in sssd.conf. Check for typos.",
		"ConfDB initialization has failed [22]":                      "Config Error: SSSD couldn't load the configuration database. Check sssd.conf for syntax errors or typos.",
		"ConfDB initialization has failed [Operation not permitted]": "Permission Denied: SSSD cannot access its configuration or state files.",
		"[13][Permission denied]":                                    "Permission Denied: SSSD cannot access a required file or log directory.",
		"cannot open shared object file":                             "Library Error: A required shared library (like libwbclient or libldb) is missing.",
		"Accessing a corrupted shared library":                       "Library Error: SSSD is trying to access a corrupted shared library.",
		"Could not restart critical service [nss]":                   "Service Crash: SSSD could not restart the NSS service (Often caused if the 'sssd' system user is missing).",
		"tdb_rec_read bad magic":                                     "Cache Corrupt: Local SSSD cache files (LDB/TDB) have bad magic numbers (Recommend running: rm -f /var/lib/sss/db/*).",
		"File must be owned by gid [0]":                              "System Error: The primary group of the root user has been modified, or SSSD pipe permissions are broken.",
		"Could not open the sysdb cache [17]":                        "Cache Corrupt: Could not open the sysdb cache due to 'File exists' lock (Recommend running: rm -f /var/lib/sss/db/*).",
		"Cannot connect to database for":                             "Cache Corrupt: Cannot connect to sysdb database due to 'File exists' lock.",
		"File ownership and permissions check failed":                "Config Error: /etc/sssd/sssd.conf has incorrect ownership or permissions.",
		"[28][No space left on device]":                              "Disk Full: No space left on device. SSSD cannot write cache or logs.",
		"Invalid base DN":                                            "Config Error: Invalid LDAP base DN syntax in sssd.conf. Run 'sssctl config-check'.",
		"error while loading shared libraries":                       "Library Error: SSSD failed to start due to missing shared libraries.",

		// --- New Responder Patterns ---
		"Reached maximum number of shells":                "Shells Limit Exceeded: /etc/shells is too large or corrupted. Users may be denied access.",
		"socket path defined in systemd unit":             "Socket Mismatch: The socket path in the systemd unit and sssd.conf do not match. Check systemd configs.",
		"is above process hard limit":                     "File Descriptor Limit: Requested fd_limit in sssd.conf exceeds the system's hard limit (ulimit). SSSD may drop connections.",
		"Refusing to read overlarge packet":               "Packet Overload: A client sent a request exceeding SSSD's maximum allowed packet size (often caused by massive Active Directory groups).",
		"Access denied for uid":                           "Permission Denied: A local process was blocked from querying SSSD because its UID is not in the 'allowed_uids' list.",
		"SELINUX_getpeercon failed":                       "SELinux Warning: SSSD attempted to read SELinux socket contexts but failed (usually means SELinux is disabled on the host).",
		"Activated socket is not a UNIX listening socket": "Socket Error: systemd passed an invalid socket type to SSSD during socket activation.",

		// --- Data Provider (DP) & PAM Patterns ---
		"Data Provider Error: 1": "Data Provider Offline (DP_ERR_OFFLINE): SSSD cannot reach the backend authentication server (AD/LDAP).",
		"Data Provider Error: 2": "Data Provider Timeout (DP_ERR_TIMEOUT): The backend server is too slow or network latency is too high.",
		"Data Provider Error: 3": "Data Provider Fatal (DP_ERR_FATAL): The backend process crashed or encountered an unrecoverable state.",
		"PAM_AUTHINFO_UNAVAIL":   "PAM Offline Trigger: The provider is offline; the PAM responder was instructed to fall back to cached authentication.",

		// --- Active Directory Provider Patterns ---
		"is detected as IP address, this can cause GSSAPI": "Kerberos SPN Risk: 'ad_server' is configured as an IP address instead of a hostname. Kerberos (GSSAPI) requires hostnames to request tickets.",
		"Group Policy Container with DN":                   "GPO Permission Denied: The Linux machine account cannot read the GPO in Active Directory. Check AD security filtering permissions.",
		"PAM service":                                      "GPO Configuration Error: A PAM service is mapped to multiple GPO rules. Check 'ad_gpo_map_*' settings in sssd.conf.",
		"would have been denied GPO-based logon":           "[DIAGNOSTIC HINT] GPO Permissive Mode: A user logged in, but would have been blocked if 'ad_gpo_access_control' was set to enforcing.",
		"no netlogon information available":                "CLDAP Ping Failed: SSSD could not reach the Domain Controller over UDP port 389 to get Site information. Check firewalls.",
		"DNS updates requested but nsupdate not available": "Missing Dependency: Dynamic DNS updates are enabled, but the 'nsupdate' utility (bind-utils) is not installed.",
		"AD domain can not be set as case-sensitive":       "Configuration Error: 'case_sensitive = true' is set, but Active Directory domains are strictly case-insensitive.",

		// --- Simple Access Provider Patterns ---
		"does not exist. Possible typo in simple_":           "Simple Access Typo: A user or group listed in your simple_allow or simple_deny configuration does not exist in the directory.",
		"No rules supplied for simple access provider":       "[DIAGNOSTIC HINT] Open Access: The simple access provider is enabled, but no rules are defined. All users are being allowed access.",
		"found in deny list, access denied":                  "Access Denied: The user or their group was explicitly blocked by a 'simple_deny_users' or 'simple_deny_groups' configuration.",
		"POSIX group without GID":                            "Corrupt Group Data: SSSD found a POSIX group in the cache, but it lacks a GID number. This breaks access control lookups.",
		"Failed to refresh filter lists, denying all access": "Fail Closed: SSSD failed to read its internal filter lists and defaulted to a hard deny for all users to protect the system.",

		// --- LDAP & Authentication Provider Patterns ---
		"Shadow password policy is selected but ldap_chpass_update_last_change is not set": "Password Reset Loop: Shadow policy is enforcing expiration, but SSSD is not configured to update the 'shadowLastChange' attribute. Users will be stuck resetting their passwords forever.",
		"No encryption detected on LDAP connection":                                        "Insecure Auth Blocked: SSSD aborted the login because it refuses to send passwords over an unencrypted LDAP connection. Enable TLS/SSL or use GSSAPI.",
		"not found in keytab": "Keytab Principal Missing: The ldap_child process could not find the required Kerberos Principal in the system keytab. Check /etc/krb5.keytab.",
		"Mapping ID":          "ID Mapping Failure: SSSD failed to algorithmically map a Windows SID to a Linux UID/GID. The SID might be from an untrusted domain or out of range.",
		"The last password change time is in the future":                       "Time Skew Bug: The LDAP server's clock and the local client's clock are out of sync, causing the password age calculation to break.",
		"Cannot determine the Kerberos realm, aborting":                        "Missing Kerberos Realm: SSSD requires a Kerberos realm for GSSAPI auth, but none was defined in sssd.conf or /etc/krb5.conf.",
		"LDAP sizelimit was exceeded, returning incomplete data":               "LDAP Size Limit Hit: The LDAP server refused to send all requested data because the query exceeded the server's hard size limit. Check LDAP paging settings.",
		"seems slow, took more than 80% of timeout":                           "[DIAGNOSTIC HINT] LDAP Latency: An LDAP operation nearly timed out. The Domain Controller is overloaded, network latency is high, or an unindexed search was performed.",
		"The user account is disabled on the AD server":                        "AD Account Disabled: Active Directory's 'userAccountControl' attribute indicates this account is explicitly disabled.",
		"The user account is expired on the AD server":                         "AD Account Expired: Active Directory's 'accountExpires' timestamp is in the past. The user must be renewed in AD.",
		"Conflicting values for options":                                       "Configuration Conflict: Your sssd.conf has conflicting cache expiration settings (e.g., keeping offline credentials forever while strictly expiring the account cache).",
		"Missing authorized services. Access denied":                           "Service Access Denied: 'ldap_access_order = service' is configured, but the user lacks the 'authorizedService' attribute in LDAP.",
		"is already used by SSSD, please choose a different":                   "Attribute Mapping Collision: You are trying to map a custom LDAP attribute in 'ldap_user_extra_attrs' to an internal name that SSSD already uses.",
		"Member [.*] was not found in cache. Is it out of scope?":              "Ghost Member Issue: A group contains a member (often from a trusted AD domain) that falls outside SSSD's configured search bases. The member will be tracked as an unresolved 'ghost'.",
		"is outside nesting limit":                                             "Nesting Limit Reached: A user is missing group memberships because the AD/LDAP group hierarchy is deeper than the configured 'ldap_group_nesting_level' (default 2).",
		"Consider enabling sssd-ldap option ldap_ignore_unreadable_references": "Unreadable Group Member: A group contains an object SSSD cannot read (lack of permissions or unknown type). This aborted the group lookup. Set 'ldap_ignore_unreadable_references = True' to bypass.",
		"doesn't have subid range":                                             "Sub-ID Missing: SSSD was asked to find a subordinate UID/GID range for a user (often required for rootless containers like Podman), but none exists in LDAP.",
		"Found more than one netgroup with the name":                           "Netgroup Conflict: Multiple nisNetgroup objects share the same name in LDAP. SSSD requires unique names. This usually breaks sudo rules or NFS mounts relying on this netgroup.",

		// --- LDAP Async Parsing & Logic Patterns ---
		"data 775,":                                        "AD Account Locked: Active Directory returned 'data 775' during bind, which specifically indicates the user account is locked out.",
		"filtered out! (uid out of range)":                 "ID Out of Range: The user's UID is outside the 'min_id' or 'max_id' configured for this domain. SSSD is intentionally ignoring them.",
		"filtered out! (primary gid out of range)":         "ID Out of Range: The user's primary GID is outside the 'min_id' or 'max_id' configured for this domain. SSSD is intentionally ignoring them.",
		"The LDAP scheme is ldapi://, cannot proceed":      "DynDNS Configuration Error: Dynamic DNS updates cannot be performed when the LDAP URI is set to a local socket (ldapi://).",
		"Unable to retrieve host information, host filter": "Sudo Host Filter Disabled: SSSD could not resolve the local machine's hostname or IP. Sudo rules restricted by host might not work correctly.",
		"Password expired, grace logins exhausted":         "Grace Logins Exhausted: The user's password expired and they have used up all of their allowed grace logins.",

		// --- LDAP Connection, ID Mapping & Range Retrieval Patterns ---
		"Range size does not divide evenly":                            "ID Mapping Warning: Your 'ldap_min_id', 'ldap_max_id', and 'ldap_idmap_rangesize' do not divide evenly. SSSD will ignore the uppermost IDs, which wastes ID space.",
		"LDAP server claims to support deref, but deref search failed": "Performance Degradation: The LDAP server advertised Dereference support but failed to execute it. SSSD is falling back to slower, individual user lookups.",
		"too many communication failures, giving up":                   "Failover Exhausted: SSSD experienced repeated network/communication errors and has exhausted all retry attempts across all failover servers. The domain is now offline.",
		"Both ldap_min_id and ldap_max_id either must be 0":            "Configuration Error: 'ldap_min_id' and 'ldap_max_id' must both be positive integers, or both left unset (0). You cannot set one without the other.",

		// --- LDAP Sudo Refresh & Cache Maintenance Patterns ---
		"Periodical smart refresh will be disabled":                    "Sudo Configuration Error: 'ldap_sudo_smart_refresh_interval' is set higher than the full refresh interval. Smart refreshes have been disabled, which may degrade performance.",
		"Skipping smart refresh because there is ongoing full refresh": "[DIAGNOSTIC HINT] Sudo Refresh Collision: SSSD skipped a quick 'smart' sudo update because a heavy 'full' download of the sudoers tree was currently running.",
		"Server reinitialization detected. Cleaning cache":             "USN Rollback Detected: The LDAP server's Update Sequence Number (USN) is lower than SSSD's cached value. This usually means the Domain Controller was restored from an old snapshot. SSSD is wiping its cache to prevent corruption.",
		"Illegal deref option":                                         "Configuration Error: The 'ldap_deref' option in sssd.conf is invalid. It must be set to 'never', 'searching', 'finding', or 'always'.",

		// --- Kerberos (krb5) Provider Patterns ---
		"Cannot find KDC for requested realm":                    "KDC Unreachable: SSSD cannot locate the Kerberos server for the realm. Check DNS SRV records (_kerberos._tcp) or the 'krb5_server' configuration.",
		"Cannot resolve network address for KDC":                 "DNS Resolution Failed: SSSD found the KDC hostname but cannot resolve its IP address. Check your DNS resolver (/etc/resolv.conf).",
		"Client not found in Kerberos database":                  "Principal Missing: The user's Kerberos Principal (UPN) does not exist in the AD/IPA database, or the realm is incorrect.",
		"Authentication failed and offline login is not allowed": "Offline Blocked: The authentication server is unreachable, and 'cache_credentials = False' prevents the user from logging in with a cached password.",

		// --- Kerberos Ticket Renewal & Keytab Patterns ---
		"TGT is not renewable":      "Non-Renewable Ticket: SSSD attempted to renew a Kerberos ticket in the background, but the KDC issued the ticket without the 'renewable' flag. Check KDC/AD ticket policies.",
		"Failed to read keytab":     "Keytab I/O Error: SSSD cannot read the Kerberos keytab file off the disk. This is usually caused by incorrect file permissions on /etc/krb5.keytab.",
		"TGT renewal failed":        "Ticket Renewal Failed: The background task to extend the user's Kerberos ticket (TGT) failed. The user may lose Single Sign-On (SSO) access to network resources.",
		"Failed to initialize FAST": "FAST Tunnel Failure: Kerberos FAST (secure tunneling) was requested, but SSSD failed to initialize it. This usually means the machine's own AD/IPA account is expired or broken.",

		// --- Data Provider (DP) Core Patterns ---
		"Cannot load module":                "Missing Provider Module: SSSD is trying to load a provider module (e.g., libsss_ad.so), but the file is missing from the system. Check your installed packages.",
		"dlopen failed:":                    "Shared Library Error: SSSD failed to dynamically load a required library. This usually indicates a broken installation or corrupted package.",
		"is not configured":                 "Unconfigured Feature: A subsystem (like sudo or autofs) requested data, but its specific provider (e.g., 'sudo_provider') is not defined in sssd.conf.",
		"BUG: The DP request was completed": "Internal DP Bug: The Data Provider finished an operation but failed to send a reply back to the responder. This is an internal SSSD bug.",
		"Unable to load target":             "Target Load Failure: SSSD failed to initialize a specific target (like sudo, autofs, or hostid). This usually means a missing library or memory allocation failure.",

		// --- Backend (BE) & Failover Core Patterns ---
		"Unable to subscribe to netlink monitor":      "Network Monitor Failure: SSSD could not subscribe to kernel netlink events. It will not be able to automatically detect when the network goes offline or comes back online.",
		"Requested type 'Number' for option":          "Configuration Typo: A setting in sssd.conf requires a number, but text was provided. Please check your configuration file for type mismatches.",
		"Requested type 'Boolean' for option":         "Configuration Typo: A setting in sssd.conf requires a boolean (true/false), but a different type was provided.",
		"journald logging might not work as expected": "[DIAGNOSTIC HINT] Journald Tagging Failed: SSSD could not set the domain environment variable. Logs will still write, but journalctl domain filtering may be degraded.",
		"Could not initialize backend":                "Backend Crash: The Data Provider process completely failed to initialize. Check the logs immediately preceding this error for missing keytabs, unreachable DNS, or bad configuration.",

		// --- Proxy Provider Patterns ---
		"proxy_pam_target is not set": "Proxy Configuration Error: 'proxy_pam_target' is missing in sssd.conf. The proxy auth provider requires a PAM service name to forward requests to.",
		"proxy_lib_name is not set":   "Proxy Configuration Error: 'proxy_lib_name' is missing in sssd.conf. The proxy ID provider must know which NSS library (e.g., 'files', 'nis') to load.",
		"Proxy lib":                   "Proxy Library Missing: SSSD tried to dynamically load the requested NSS proxy library, but it was not found on the system. Check your proxy_lib_name.",
		"pam_authenticate failed":     "Proxy PAM Failure: The underlying PAM module rejected the authentication request. Check the host's PAM logs (e.g., /var/log/secure) for the target PAM service.",
		"The proxy provider is unable to return multiple entries": "Proxy Limitation: The proxy provider was asked to enumerate or perform a wildcard search, which is fundamentally unsupported by the proxy architecture.",

		// --- IdP (OIDC/OAuth2) Provider Patterns ---
		"Missing required option 'idp_auth_scope'":           "IdP Configuration Error: The 'idp_auth_scope' option is missing in sssd.conf. SSSD needs to know what OAuth2 scopes to request.",
		"Missing required option 'idp_device_code_endpoint'": "IdP Configuration Error: The 'idp_device_auth_endpoint' option is missing in sssd.conf. SSSD requires the URL for the OAuth2 Device Authorization flow.",
		"Missing required option 'idp_userinfo_endpoint'":    "IdP Configuration Error: The 'idp_userinfo_endpoint' option is missing in sssd.conf. SSSD needs the URL to fetch user attributes.",
		"Failed to store JSON":                               "IdP Data Error: SSSD failed to parse or store the JSON payload returned by the external Identity Provider.",
		"has no UUID attribute":                              "IdP Security Block: The user authenticated successfully, but SSSD rejected the login because the cached user lacks a UUID attribute to securely bind the token.",

		// --- IPA Provider Patterns ---
		"Server doesn't support Desktop Profile": "Legacy Server Warning: The client requested Desktop Profile (FleetCommander) data, but the FreeIPA server is too old to support this feature.",
		"No rules apply to this host":            "[DIAGNOSTIC HINT] Desktop Profile/HBAC: Rules were fetched successfully from FreeIPA, but none of them target this specific machine's hostname or hostgroups.",
		"trying Kerberos authentication again":   "[DIAGNOSTIC HINT] IPA Auth Fallback: SSSD is retrying authentication via Kerberos. This is frequently seen during transparent Password Migrations.",
		"Dynamic DNS update failed":              "IPA DNS Error: SSSD failed to update the machine's IP address in the FreeIPA DNS zone. Check the machine's host Kerberos ticket and IPA DNS zone permissions.",

		// --- IPA HBAC & ID Mapping Patterns ---
		"does not map to either a user or group. Maybe it is an object which is currently not in the cache": "HBAC Dangling Reference: A FreeIPA HBAC rule references a user or group that SSSD cannot find in its cache or LDAP. The rule might not apply correctly.",
		"Skipping malformed entry":        "HBAC Corruption: An HBAC rule contains a malformed LDAP DN. SSSD expects a specific hierarchy for services and hosts. The rule will be partially skipped.",
		"Could not initialize the ID map": "IPA ID Mapping Error: SSSD failed to initialize the ID map for FreeIPA/AD trusts, likely due to bad range configurations or cache corruption. You may need to clear the SSSD cache.",
		"Added external source host":      "[DIAGNOSTIC HINT] HBAC External Host: SSSD is successfully applying an HBAC rule to an unmanaged, external host.",

		// --- IPA SELinux Mapping Patterns ---
		"Cannot determine the host name":                      "Missing Hostname for SELinux: SSSD cannot determine the machine's hostname, which is strictly required to evaluate host-specific SELinux user maps in FreeIPA.",
		"SELinux maps referenced an HBAC rule":                "[DIAGNOSTIC HINT] SELinux/HBAC Linkage: A FreeIPA SELinux user map is linked to an HBAC rule. SSSD is fetching the HBAC rule to evaluate the SELinux context.",
		"SELinux maps were recently updated -> force offline": "[DIAGNOSTIC HINT] Cached SELinux Evaluation: SELinux maps were refreshed recently. SSSD is evaluating the SELinux context using the local cache to save network traffic.",
		"Failed to evaluate ordered SELinux users array":      "SELinux Map Evaluation Failed: SSSD could not map the logging-in user to a valid SELinux user context based on the FreeIPA rules and priority order.",

		// --- IPA Subdomains (AD Trust) & Sudo Patterns ---
		"lookups of subdomain users will likely fail": "Cross-Forest Trust Warning: 'full_name_format' was altered in sssd.conf. SSSD strictly relies on the default format to route Active Directory trust users. AD logins will likely break.",
		"Could not reinitialize subdomains":           "Trust Topology Error: SSSD failed to refresh the list of trusted Active Directory domains from the FreeIPA server. Check the IPA-to-AD trust status.",
		"Unable to choose sudo schema":                "Sudo Initialization Error: SSSD could not determine whether to use the native IPA sudo schema or the legacy LDAP sudo schema. Check your configuration.",
		"is not a valid DN":                           "[DIAGNOSTIC HINT] Corrupt Directory Object: SSSD encountered a malformed Distinguished Name (DN) while parsing directory objects, causing it to skip the entry.",

		// --- IPA ID Views & SELinux Child Patterns ---
		"is defined in policy, cannot be deleted": "SELinux Policy Conflict: SSSD tried to apply a FreeIPA SELinux mapping, but a local admin manually defined a conflicting rule using the 'semanage' command on this host.",
		"Broken IPA anchor":                       "ID View Corruption: SSSD encountered a malformed ID Override anchor. The cache for overridden Active Directory users might be corrupted.",
		"setresuid() failed":                      "Privilege Escalation Failed: The selinux_child helper process failed to change its UID/GID. Check if AppArmor, SELinux, or container capabilities are blocking CAP_SETUID.",

		// --- Responder (Frontend) & Cache Patterns ---
		"is vetoed. Using fallback":                     "Shell Override: The user's LDAP shell is explicitly vetoed in sssd.conf. SSSD is overriding it with a safe fallback shell.",
		"is allowed but does not exist. Using fallback": "Shell Missing: The user's LDAP shell does not exist on this Linux machine (not in /etc/shells). SSSD is assigning a fallback shell to prevent a login crash.",
		"Data Provider Error:":                          "Backend Unreachable: The SSSD frontend responder requested data, but the backend Data Provider returned a fatal error or timed out.",
		"Failed to store permanent uid filter for root": "Negative Cache Error: SSSD failed to add 'root' to the negative cache. SSSD might cause high network load by querying the network for local system accounts.",

		// --- Responder Cache Request Patterns ---
		"Cannot bypass cache and dp at the same time!":                   "Internal Request Error: SSSD attempted to bypass both the local cache and the network provider simultaneously. The request was dropped.",
		"No requested domains found, please check configuration options": "Domain Resolution Error: A requested domain was not found. Please check 'domain_resolution_order' or 'domains' in sssd.conf for typos.",
		"Mismatch between input domain name":                             "Domain Parsing Mismatch: The requested domain name does not match the internally parsed domain. Check your 'full_name_format' setting.",
		"Enumeration requested but not enabled":                          "Enumeration Blocked: A bulk data request (like 'getent passwd') was attempted, but 'enumerate = True' is not configured for this domain.",

		// --- Responder Cache Request Common Patterns ---
		"does not contain the root user":            "Security Block: SSSD prevented a remote domain from returning a user with UID 0 (root). The root user is strictly restricted to the local files domain to prevent privilege escalation.",
		"The entry is outside the allowed ID range": "ID Out of Bounds: SSSD successfully retrieved the entry, but its UID/GID falls outside the 'min_id' or 'max_id' boundaries defined for this domain. The entry has been dropped.",
		"Unable to parse name":                      "Input Parsing Error: SSSD could not parse the requested name into a valid shortname and domain. This usually indicates a malformed request or an issue with 'full_name_format' regex settings.",

		// --- Responder Cache Request Plugins Patterns ---
		"is outside of ID range":             "ID Out of Range (Frontend): A requested UID or GID falls outside the 'min_id' or 'max_id' configured for the domain. The SSSD frontend rejected the lookup before even querying the backend.",
		"Subdomain is not set, UPN":          "UPN Routing Failure: SSSD received an initgroups request using a User Principal Name (UPN), but it could not determine the correct active subdomain to route the request to. Check AD trust status.",
		"filtered out! (groupname override)": "Group Name Override: SSSD ignored a group lookup because the group's name has been locally overridden (e.g., via FreeIPA ID Views or local overrides). The original name is no longer valid for lookups.",

		// --- Responder Cache Request Plugins (Specialized) Patterns ---
		"Multiple users matched the certificate": "Smartcard Auth Collision: SSSD searched the directory for the provided Smartcard/X.509 certificate and found multiple users mapped to it. Authentication is blocked for security.",
		"Cannot parse certificate":               "Smartcard Driver Error: SSSD received a certificate payload from the PAM responder, but it was corrupted or malformed. Check your pcscd or PAM PKCS#11 configuration.",
		"does not belong to any known domain":    "SID Resolution Failure: SSSD was asked to resolve a Windows SID, but the SID prefix does not match the local domain or any known trusted Active Directory domains.",
		"No SSH host keys found for":             "[DIAGNOSTIC HINT] SSH Key Missing: SSSD found the requested host object in the directory, but it lacks the required 'sshPublicKey' attribute.",

		// --- Sudo Responder Patterns ---
		"Unable to set up sudo search base":     "Sudo Config Error: SSSD could not initialize the sudo search base. Sudo rules in LDAP/AD will not be accessible.",
		"Failed to initialize SUDO ncache":      "Sudo Cache Error: The Sudo negative cache failed to initialize. This may degrade performance or cause repeated failed lookups.",
		"Failed to get user info for sudo user": "User Resolution Failure: SSSD could not resolve the identity of the user requesting sudo. The permission check cannot proceed.",
		"Data provider is not available for":    "Sudo Backend Error: The Sudo responder cannot reach the Data Provider for this domain. Rule refreshes will fail.",
		"Sudo rules refresh failed":             "Sudo Refresh Failure: SSSD encountered an error while trying to update sudo rules from the backend network provider.",
		"Failed to create sudo query":           "Sudo Query Error: SSSD failed to construct the LDAP search query for sudo rules. Check for malformed filters in sssd.conf.",

		// --- SSH Responder Patterns ---
		"Failed to initialize SSH ncache":      "SSH Cache Error: The SSH negative cache failed to initialize. SSSD-managed SSH key lookups will be disabled.",
		"Failed to get user info for SSH user": "User Resolution Failure: SSSD could not resolve the user identity prior to fetching SSH keys. Key lookups will fail.",
		"Failed to parse certificate":          "SSH Certificate Error: SSSD could not parse the X.509 certificate from the directory into a valid SSH key. The certificate may be malformed.",
		"Not an SSH certificate":               "SSH Certificate Error: The certificate provided in LDAP/AD is not a valid SSH certificate and cannot be used for authentication.",
		"Failed to create SSH request":         "Internal SSH Error: SSSD failed to initialize the SSH request context. This may indicate memory pressure or an internal protocol error.",

		// --- PAM Responder Patterns ---
		"Passkey data for user .* not found":    "Passkey Provisioning Error: The user attempted to log in with a Passkey, but their LDAP/AD object is missing the required Passkey registration data.",
		"Malformed PAM prompting configuration": "Prompting Config Error: The [pam] section of sssd.conf has invalid prompting rules. Check the syntax of your custom MFA messages.",
		"Data provider is not available":        "Backend Disconnected: The PAM responder cannot reach the Data Provider process. This usually means the backend (AD/IPA/LDAP) has crashed or is hung.",

		// --- PAC and NSS Responder Patterns ---
		"Failed to create PAC responder context":           "PAC Initialization Error: The PAC responder (used for AD forest trusts) failed to start. Group memberships for trusted domain users may be missing.",
		"Search for root user in domain .* is not allowed": "Security Block: SSSD blocked a request to find 'root' in a remote domain. This is a safety feature to prevent remote directories from hijacking UID 0.",
		"Illegal ID .* for search":                         "Invalid ID Request: The OS or an application requested a negative or overflowed UID/GID. SSSD rejected the request as invalid.",
		"PAC check failed":                                 "Kerberos PAC Error: SSSD could not validate the Privilege Attribute Certificate (PAC) in the user's ticket. Group memberships might not be updated.",

		// --- NSS and IFP Responder Patterns ---
		"Not enough space in the .* mmap cache":       "Performance Warning: The fast-path memory cache is full. SSSD is falling back to slower disk-based lookups. Consider increasing the 'memcache_size' for this domain.",
		"Failed to initialize NSS ncache":             "NSS Init Error: The NSS negative cache failed to start. User and group resolution might be completely disabled or severely degraded.",
		"Failed to format string SID from binary SID": "SID Corruption: SSSD encountered a binary SID in AD/LDAP that it cannot translate into a string. The directory object may be malformed or corrupted.",
		"Could not initialize InfoPipe responder":     "InfoPipe Error: The SSSD D-Bus interface (ifp) failed to initialize. D-Bus based identity queries will fail.",
		"has no name attribute!?!":                    "Cache Consistency Error: SSSD found a directory object without a mandatory name attribute. This usually indicates cache corruption or interrupted database writes.",

		// --- InfoPipe (IFP) Responder Patterns ---
		"Permission denied for UID":           "IFP UID Blocked: The process attempting to query SSSD via D-Bus is not authorized. Add the process's UID to 'allowed_uids' in the [ifp] section of sssd.conf.",
		"Attribute .* not allowed for user":   "IFP Attribute Filter: An application requested an LDAP/AD attribute that is not whitelisted. Add it to 'user_attributes' in the [ifp] section.",
		"Failed to setup IFP D-Bus interface": "D-Bus Connection Error: InfoPipe could not register its D-Bus service. Check if dbus-daemon is running and if the org.freedesktop.sssd.infopipe policy file is correct.",
		"Failed to initialize IFP domains":    "IFP Initialization Failure: InfoPipe failed to load the domain list. This usually points to an overall configuration error in sssd.conf.",
		"Failed to initialize IFP ncache":     "IFP Cache Error: The InfoPipe negative cache failed to start. Performance for failed lookups will be significantly degraded.",

		// --- Monitor (Supervisor) Patterns ---
		"Monitor reached maximum number of restarts": "Service Flapping: SSSD tried to restart a crashing service (like nss or pam) too many times. SSSD has stopped trying. Check for segfaults or configuration errors in that specific service.",
		"is not responding to heartbeat":             "Service Deadlock: A child process stopped responding to SSSD's internal health checks. SSSD is killing and restarting the hung process to restore service.",
		"Unable to initialize ConfDB":                "Config Database Error: SSSD could not initialize its internal configuration database. This usually means sssd.conf has a syntax error or the disk is full.",
		"SSSD is already running":                    "Instance Conflict: SSSD detected another instance is already running (or an old PID file exists). SSSD refused to start to prevent data corruption.",
		"No domains configured, exiting":             "Configuration Error: You have not defined any [domain/NAME] sections in sssd.conf. SSSD has nothing to do and is exiting.",

		// --- SysDB (Local Cache) Patterns ---
		"Could not open ldb at":                       "Database Access Error: SSSD cannot open its local cache file. Check file permissions on /var/lib/sss/db/ or check for disk space/corruption.",
		"The database was created by a newer version": "Version Mismatch: You are trying to run an older version of SSSD with a database created by a newer version. You may need to clear the cache.",
		"Failed to upgrade the database":              "Upgrade Failure: SSSD tried to update the internal database schema to a new version and failed. The cache may be corrupted.",
		"sysdb_transaction_start failed":              "Database Locking Error: SSSD could not start a write transaction. This often indicates the disk is full or the database file is read-only.",
		"Failed to commit transaction!":               "Database Write Error: SSSD could not finalize a data save. This is a common symptom of a full disk or failing hardware.",
		"Failed to store ID mapping":                  "ID Map Persistence Error: SSSD could not save a generated UID/GID. This will break identity consistency for Active Directory users.",

		// --- Final SysDB & Cache Upgrade Patterns ---
		"Failed to store ID range":    "ID Range Conflict: SSSD could not save an ID range for a domain. This is likely due to overlapping ranges in sssd.conf or a conflict with a trusted domain.",
		"Failed to store ID override": "Override Error: SSSD failed to save an ID View override from FreeIPA. Users may log in with original, un-overridden attributes.",
		"Internal search error":       "LDB Corruption: SSSD encountered a fatal error querying its own database. This is a strong indicator of file corruption or a broken LDB index.",

		// --- ConfDB, Kerberos Plugins, and ID Mapping Patterns ---
		"Multiple domains with the same name": "Configuration Error: You have defined the same domain name multiple times in sssd.conf. SSSD cannot start with ambiguous domain definitions.",
		"ConfDB initialization failed":        "Fatal Config Error: SSSD failed to build its internal configuration database. Check for sssd.conf syntax errors or disk space issues.",
		"Failed to locate KDC via SSSD":       "Kerberos Locator Error: The Kerberos plugin tried to ask SSSD for a Domain Controller location, but SSSD was unreachable or had no KDC data.",
		"Failed to convert SID to UID/GID":    "ID Mapping Error: SSSD could not translate a Windows SID to a Linux ID. This will cause file permission issues on CIFS/Samba shares.",
		"Invalid SID format":                  "Directory Data Error: SSSD encountered a malformed Windows SID in the directory. The object cannot be mapped to a Unix identity.",
		"ID mapping out of range":             "Range Conflict: An AD user's calculated ID falls outside the 'min_id' or 'max_id' limits. The user will be unable to log in or own files.",
	}

	if macType != "SELinux" {
		for key, desc := range errorPatterns {
			if strings.Contains(desc, "SELinux") && !strings.Contains(desc, "AppArmor") {
				delete(errorPatterns, key)
			}
		}
	}

	return errorPatterns
}

// normalizeTimestamp converts a timestamp string from any supported format
// into a canonical YYYY-MM-DD HH:MM:SS representation for reliable sorting.
// Returns an empty string if the timestamp cannot be parsed.
func normalizeTimestamp(ts string) string {
	// Already in SSSD native format
	if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}

	// ISO 8601 format (with or without fractional seconds and timezone)
	isoFormats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.999999999-07:00",
		"2006-01-02T15:04:05.999-07:00",
		"2006-01-02T15:04:05+07:00",
		"2006-01-02T15:04:05.999999999+07:00",
		"2006-01-02T15:04:05.999+07:00",
	}
	for _, isoFmt := range isoFormats {
		if t, err := time.Parse(isoFmt, ts); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	// Syslog format: "Jan  2 15:04:05" or "Jan  2 15:04"
	syslogFormats := []string{
		"Jan  2 15:04:05",
		"Jan 2 15:04:05",
		"Jan  2 15:04",
		"Jan 2 15:04",
	}
	for _, syslogFmt := range syslogFormats {
		if t, err := time.Parse(syslogFmt, ts); err == nil {
			// Use current year for syslog timestamps (they lack year info)
			currentYear := time.Now().Year()
			t = t.AddDate(currentYear-t.Year(), 0, 0)
			return t.Format("2006-01-02 15:04:05")
		}
	}

	// Syslog format with year: "Jan  2 15:04:05 2024"
	syslogFormatsWithYear := []string{
		"Jan  2 15:04:05 2006",
		"Jan 2 15:04:05 2006",
	}
	for _, syslogFmt := range syslogFormatsWithYear {
		if t, err := time.Parse(syslogFmt, ts); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
	}

	return ""
}

// scanAndCollectErrors scans the log files for known SSSD error patterns,
// deduplicates matches (max 3 examples per pattern), and builds a timeline of events.
func scanAndCollectErrors(dirPath string, errorPatterns map[string]string) (map[string][]string, []TimelineEvent) {
	detectedLogErrors := make(map[string][]string)
	var timelineEvents []TimelineEvent

	// Build the fast O(1) lookup and the compiled regex for scanning
	var errorKeys []string
	fastLookup := make(map[string]string)
	for pattern, desc := range errorPatterns {
		errorKeys = append(errorKeys, regexp.QuoteMeta(pattern))
		fastLookup[strings.ToLower(pattern)] = desc
	}
	// Use global regex cache to avoid recompiling this large pattern on every analysis
	errorRegex := globalRegexCache.Get("(?i)(" + strings.Join(errorKeys, "|") + ")")

	// Regex to match standard SSSD log timestamps, standard syslog timestamps, and ISO 8601 timestamps
	timeRegex := globalRegexCache.Get(`(?:\((\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\)|([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})|(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?))`)

	scanFiles(dirPath, logFileNames, func(line string) {
		lineTrimmed := strings.TrimSpace(line)

		// Ignore logs from winbindd to prevent false positives in /var/log/messages
		if strings.Contains(strings.ToLower(lineTrimmed), "winbindd") {
			return
		}

		matches := errorRegex.FindAllString(lineTrimmed, -1)
		for _, match := range matches {
			// O(1) lookup instead of O(N) nested loop
			description, ok := fastLookup[strings.ToLower(match)]
			if !ok {
				continue
			}

			examples := detectedLogErrors[description]
			if len(examples) < 3 {
				isDupe := false
				for _, ex := range examples {
					if ex == lineTrimmed {
						isDupe = true
						break
					}
				}
				if !isDupe {
					detectedLogErrors[description] = append(detectedLogErrors[description], lineTrimmed)

					tsMatch := timeRegex.FindStringSubmatch(lineTrimmed)
					ts := "Unknown Time"
					if len(tsMatch) > 1 && tsMatch[1] != "" {
						ts = tsMatch[1] // SSSD native format
					} else if len(tsMatch) > 2 && tsMatch[2] != "" {
						ts = tsMatch[2] // Syslog format
					} else if len(tsMatch) > 3 && tsMatch[3] != "" {
						ts = tsMatch[3] // ISO 8601 format
					}

					timelineEvents = append(timelineEvents, TimelineEvent{
						Timestamp: ts,
						Message:   description,
						RawLog:    lineTrimmed,
					})
				}
			}
		}
	})

	// Sort the timeline chronologically using normalized timestamps
	sort.SliceStable(timelineEvents, func(i, j int) bool {
		ti := normalizeTimestamp(timelineEvents[i].Timestamp)
		tj := normalizeTimestamp(timelineEvents[j].Timestamp)
		// Empty (unparseable) timestamps sort last
		if ti == "" && tj == "" {
			return false
		}
		if ti == "" {
			return false
		}
		if tj == "" {
			return true
		}
		return ti < tj
	})

	return detectedLogErrors, timelineEvents
}

// buildSortedLogErrors converts the detected log errors map into a sorted slice of SSSDLogError.
func buildSortedLogErrors(detectedLogErrors map[string][]string) []SSSDLogError {
	var errors []SSSDLogError
	for desc, lines := range detectedLogErrors {
		errors = append(errors, SSSDLogError{Description: desc, Examples: lines})
	}
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Description < errors[j].Description
	})
	return errors
}


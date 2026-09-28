# SSSD Configuration Options Reference Catalog

Auto-generated from SSSD 2.14.0 on 2026-09-28.
Sources: `ad_modified_defaults.xml`, `autofs_attributes.xml`, `autofs_restart.xml`, `debug_levels.xml`, `debug_levels_tools.xml`, `failover.xml`, `homedir_substring.xml`, `idmap_sss.8.xml`, `ipa_modified_defaults.xml`, `krb5_options.xml`, `ldap_id_mapping.xml`, `ldap_search_bases.xml`, `override_homedir.xml`, `pam_sss.8.xml`, `pam_sss_gss.8.xml`, `param_help.xml`, `param_help_py.xml`, `seealso.xml`, `service_discovery.xml`, `sss-certmap.5.xml`, `sss_cache.8.xml`, `sss_debuglevel.8.xml`, `sss_obfuscate.8.xml`, `sss_override.8.xml`, `sss_rpcidmapd.5.xml`, `sss_seed.8.xml`, `sss_ssh_authorizedkeys.1.xml`, `sss_ssh_knownhosts.1.xml`, `sssctl.8.xml`, `sssd-ad.5.xml`, `sssd-ad.conf`, `sssd-idp.5.xml`, `sssd-ifp.5.xml`, `sssd-ipa.5.xml`, `sssd-ipa.conf`, `sssd-kcm.8.xml`, `sssd-krb5.5.xml`, `sssd-krb5.conf`, `sssd-ldap-attributes.5.xml`, `sssd-ldap.5.xml`, `sssd-ldap.conf`, `sssd-proxy.conf`, `sssd-session-recording.5.xml`, `sssd-simple.5.xml`, `sssd-simple.conf`, `sssd-sudo.5.xml`, `sssd-systemtap.5.xml`, `sssd.8.xml`, `sssd.api.conf`, `sssd.conf.5.xml`, `sssd_krb5_localauth_plugin.8.xml`, `sssd_krb5_locator_plugin.8.xml`, `upstream.xml`, `version.m4`

Total Options: **519** across **44** sections.

## Sections

- [`[*]`](#section-*) (10 options)
- [`[autofs]`](#section-autofs) (1 options)
- [`[certmap]`](#section-certmap) (4 options)
- [`[domain]`](#section-domain) (106 options)
- [`[domain/ad]`](#section-domain-ad) (83 options)
- [`[domain/ad/auth]`](#section-domain-ad-auth) (15 options)
- [`[domain/ad/autofs]`](#section-domain-ad-autofs) (7 options)
- [`[domain/ad/chpass]`](#section-domain-ad-chpass) (2 options)
- [`[domain/ad/id]`](#section-domain-ad-id) (71 options)
- [`[domain/ad/resolver]`](#section-domain-ad-resolver) (10 options)
- [`[domain/ad/sudo]`](#section-domain-ad-sudo) (22 options)
- [`[domain/ipa]`](#section-domain-ipa) (86 options)
- [`[domain/ipa/access]`](#section-domain-ipa-access) (15 options)
- [`[domain/ipa/auth]`](#section-domain-ipa-auth) (15 options)
- [`[domain/ipa/autofs]`](#section-domain-ipa-autofs) (8 options)
- [`[domain/ipa/id]`](#section-domain-ipa-id) (91 options)
- [`[domain/ipa/session]`](#section-domain-ipa-session) (19 options)
- [`[domain/ipa/subdomains]`](#section-domain-ipa-subdomains) (1 options)
- [`[domain/ipa/sudo]`](#section-domain-ipa-sudo) (54 options)
- [`[domain/krb5]`](#section-domain-krb5) (19 options)
- [`[domain/krb5/auth]`](#section-domain-krb5-auth) (15 options)
- [`[domain/ldap]`](#section-domain-ldap) (90 options)
- [`[domain/ldap/access]`](#section-domain-ldap-access) (3 options)
- [`[domain/ldap/auth]`](#section-domain-ldap-auth) (1 options)
- [`[domain/ldap/autofs]`](#section-domain-ldap-autofs) (7 options)
- [`[domain/ldap/chpass]`](#section-domain-ldap-chpass) (4 options)
- [`[domain/ldap/id]`](#section-domain-ldap-id) (90 options)
- [`[domain/ldap/resolver]`](#section-domain-ldap-resolver) (10 options)
- [`[domain/ldap/sudo]`](#section-domain-ldap-sudo) (22 options)
- [`[domain/proxy]`](#section-domain-proxy) (1 options)
- [`[domain/proxy/auth]`](#section-domain-proxy-auth) (1 options)
- [`[domain/proxy/id]`](#section-domain-proxy-id) (2 options)
- [`[domain/simple]`](#section-domain-simple) (4 options)
- [`[domain/simple/access]`](#section-domain-simple-access) (4 options)
- [`[ifp]`](#section-ifp) (2 options)
- [`[kcm]`](#section-kcm) (6 options)
- [`[nss]`](#section-nss) (22 options)
- [`[pac]`](#section-pac) (3 options)
- [`[pam]`](#section-pam) (29 options)
- [`[service]`](#section-service) (12 options)
- [`[session_recording]`](#section-session_recording) (5 options)
- [`[ssh]`](#section-ssh) (5 options)
- [`[sssd]`](#section-sssd) (18 options)
- [`[sudo]`](#section-sudo) (3 options)

---

## Section `[*]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `cache_first` | `int` | `true` | - |
| `client_idle_timeout` | `int` | `60, KCM: 300` | - |
| `debug` | `int` | `-` | - |
| `debug_backtrace_enabled` | `bool` | `true` | - |
| `debug_level` | `int` | `-` | - |
| `debug_microseconds` | `bool` | `false` | - |
| `debug_timestamps` | `bool` | `true` | - |
| `fd_limit` | `int` | `8192 (or limits.conf "hard" limit)` | - |
| `responder_idle_timeout` | `int` | `300` | - |
| `timeout` | `int` | `10` | - |

## Section `[autofs]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `autofs_negative_timeout` | `int` | `15` | - |

## Section `[certmap]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `domains` | `list` | `the configured domain in sssd.conf` | - |
| `maprule` | `string` | `-` | - |
| `matchrule` | `string` | `KRB5:&lt;EKU&gt;clientAuth, i.e. only` | - |
| `priority` | `int` | `the lowest priority` | - |

## Section `[domain]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `access_provider` | `string` | `-` | permit, deny, ad, ipa, ldap, simple, proxy |
| `account_cache_expiration` | `int` | `0 (unlimited)` | - |
| `auth_provider` | `string` | `-` | none, ad, ipa, ldap, krb5, proxy |
| `auto_private_groups` | `string` | `-` | true, false, hybrid |
| `autofs_provider` | `string` | `The value of` | none, ad, ipa, ldap |
| `avoid_by_id_lookups` | `bool` | `False (True for IdP provider)` | - |
| `cache_credentials` | `bool` | `FALSE` | - |
| `cache_credentials_minimal_first_factor_length` | `int` | `8` | - |
| `cached_auth_timeout` | `int` | `0` | - |
| `case_sensitive` | `string` | `-` | true, false, preserving |
| `chpass_provider` | `string` | `-` | none, ad, ipa, ldap, krb5, proxy |
| `command` | `string` | `-` | - |
| `debug` | `int` | `-` | - |
| `debug_level` | `int` | `-` | - |
| `debug_timestamps` | `bool` | `true` | - |
| `default_shell` | `string` | `not set (Return NULL if no shell is` | - |
| `description` | `string` | `-` | - |
| `dns_discovery_domain` | `string` | `Use the domain part of machine's hostname` | - |
| `dns_resolver_op_timeout` | `int` | `3` | - |
| `dns_resolver_server_timeout` | `int` | `1000` | - |
| `dns_resolver_timeout` | `int` | `6` | - |
| `dns_resolver_use_search_list` | `bool` | `TRUE` | - |
| `domain_type` | `string` | `posix` | - |
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - |
| `dyndns_auth` | `string` | `GSS-TSIG` | - |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - |
| `dyndns_refresh_interval_offset` | `int` | `-` | - |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - |
| `dyndns_update` | `bool` | `true` | - |
| `dyndns_update_per_family` | `bool` | `true` | - |
| `dyndns_update_ptr` | `bool` | `True` | - |
| `enabled` | `bool` | `-` | - |
| `entry_cache_autofs_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_group_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_netgroup_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_resolver_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_service_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_ssh_host_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_sudo_timeout` | `int` | `entry_cache_timeout` | - |
| `entry_cache_timeout` | `int` | `5400` | - |
| `entry_cache_user_timeout` | `int` | `entry_cache_timeout` | - |
| `enumerate` | `bool` | `FALSE` | - |
| `failover_primary_timeout` | `int` | `31` | - |
| `fallback_homedir` | `string` | `not set (no substitution for unset home` | - |
| `filter_groups` | `list` | `root` | - |
| `filter_users` | `list` | `root` | - |
| `full_name_format` | `string` | `-` | - |
| `homedir_substring` | `string` | `/home` | - |
| `hostid_provider` | `string` | `The value of` | - |
| `id_provider` | `string` | `Not set (This is a mandatory setting that` | ad, ipa, ldap, proxy, simple, files |
| `idmap_range_max` | `int` | `2000200000` | - |
| `idmap_range_min` | `int` | `200000` | - |
| `idmap_range_size` | `int` | `200000` | - |
| `idp_auth_scope` | `string` | `Not set` | - |
| `idp_auto_refresh` | `bool` | `false` | - |
| `idp_client_id` | `string` | `Not set (Required)` | - |
| `idp_client_secret` | `string` | `Not set` | - |
| `idp_device_auth_endpoint` | `string` | `Not set` | - |
| `idp_id_scope` | `string` | `Not set` | - |
| `idp_request_timeout` | `int` | `10` | - |
| `idp_token_endpoint` | `string` | `Not set (Required)` | - |
| `idp_type` | `string` | `Not set (Required)` | - |
| `idp_userinfo_endpoint` | `string` | `Not set` | - |
| `ignore_group_members` | `bool` | `FALSE` | - |
| `inherit_from` | `string` | `Not set` | - |
| `local_auth_policy` | `string` | `match` | match, only, enable, disable |
| `lookup_family_order` | `string` | `ipv4_first` | - |
| `max_id` | `int` | `1 for min_id, 0 (no limit) for max_id` | - |
| `min_id` | `int` | `1 for min_id, 0 (no limit) for max_id` | - |
| `offline_timeout` | `int` | `60` | - |
| `offline_timeout_max` | `int` | `3600` | - |
| `offline_timeout_random_offset` | `int` | `30` | - |
| `override_gid` | `int` | `not set (Use the primary GID value retrieved` | - |
| `override_homedir` | `string` | `-` | - |
| `override_shell` | `string` | `not set (SSSD will use the value` | - |
| `pam_gssapi_check_upn` | `bool` | `True` | - |
| `pam_gssapi_indicators_apply` | `string` | `not set` | - |
| `pam_gssapi_indicators_map` | `string` | `not set (use of authentication indicators is not required)` | - |
| `pam_gssapi_services` | `string` | `- (GSSAPI authentication is disabled)` | - |
| `proxy_fast_alias` | `bool` | `false` | - |
| `proxy_lib_name` | `string` | `-` | - |
| `proxy_max_children` | `int` | `10` | - |
| `proxy_pam_target` | `string` | `not set by default, you have to take an` | - |
| `proxy_resolver_lib_name` | `string` | `-` | - |
| `pwd_expiration_warning` | `int` | `7 (Kerberos), 0 (LDAP)` | - |
| `re_expression` | `string` | `-` | - |
| `realmd_tags` | `string` | `-` | - |
| `refresh_expired_interval` | `int` | `0 (disabled)` | - |
| `refresh_expired_interval_offset` | `int` | `-` | - |
| `resolver_provider` | `string` | `The value of` | - |
| `selinux_provider` | `string` | `the value of` | - |
| `session_provider` | `string` | `-` | - |
| `subdomain_homedir` | `string` | `-` | - |
| `subdomain_inherit` | `string` | `none` | - |
| `subdomain_refresh_interval` | `int` | `-` | - |
| `subdomain_refresh_interval_offset` | `int` | `-` | - |
| `subdomains_provider` | `string` | `The value of` | - |
| `sudo_provider` | `string` | `The value of` | none, ad, ipa, ldap |
| `timeout` | `int` | `10` | - |
| `use_fully_qualified_names` | `bool` | `FALSE (TRUE for trusted` | - |

## Section `[domain/ad]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ad_access_filter` | `string` | `Not set` | - |
| `ad_backup_server` | `string` | `-` | - |
| `ad_domain` | `string` | `-` | - |
| `ad_enable_dns_sites` | `bool` | `true` | - |
| `ad_enable_gc` | `bool` | `true` | - |
| `ad_enabled_domains` | `string` | `Not set` | - |
| `ad_gpo_access_control` | `string` | `permissive` | disabled, enforcing, permissive |
| `ad_gpo_cache_timeout` | `int` | `5 (seconds)` | - |
| `ad_gpo_default_right` | `string` | `deny` | interactive, remote_interactive, network, batch, service, permit, deny |
| `ad_gpo_ignore_unreadable` | `bool` | `False` | - |
| `ad_gpo_implicit_deny` | `bool` | `False` | - |
| `ad_gpo_map_batch` | `string` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_deny` | `string` | `not set` | - |
| `ad_gpo_map_interactive` | `string` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_network` | `string` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_permit` | `string` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_remote_interactive` | `string` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_service` | `string` | `not set` | - |
| `ad_hostname` | `string` | `-` | - |
| `ad_machine_account_password_renewal_opts` | `string` | `86400:750:300:realm (24h, 12m30s and 5m)` | - |
| `ad_maximum_machine_account_password_age` | `int` | `30 days` | - |
| `ad_server` | `string` | `-` | - |
| `ad_site` | `string` | `Not set` | - |
| `ad_update_samba_machine_account_password` | `bool` | `false` | - |
| `ad_use_ldaps` | `bool` | `False` | - |
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - |
| `dyndns_auth` | `string` | `GSS-TSIG` | - |
| `dyndns_auth_ptr` | `string` | `Same as dyndns_auth` | - |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - |
| `dyndns_update` | `bool` | `true` | - |
| `dyndns_update_per_family` | `bool` | `true` | - |
| `dyndns_update_ptr` | `bool` | `True` | - |
| `krb5_auth_timeout` | `int` | `-` | - |
| `krb5_backup_server` | `string` | `-` | - |
| `krb5_canonicalize` | `bool` | `false` | - |
| `krb5_confd_path` | `string` | `not set (krb5.include.d subdirectory of` | - |
| `krb5_kdcip` | `string` | `-` | - |
| `krb5_realm` | `string` | `System defaults, see` | - |
| `krb5_server` | `string` | `-` | - |
| `krb5_use_kdcinfo` | `bool` | `true` | - |
| `ldap_backup_uri` | `string` | `-` | - |
| `ldap_connection_expire_offset` | `int` | `0` | - |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_default_authtok` | `string` | `-` | - |
| `ldap_default_authtok_type` | `string` | `password` | - |
| `ldap_default_bind_dn` | `string` | `-` | - |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always |
| `ldap_deref_threshold` | `int` | `10` | - |
| `ldap_disable_paging` | `bool` | `False` | - |
| `ldap_dns_service_name` | `string` | `ldap` | - |
| `ldap_entry_usn` | `string` | `-` | - |
| `ldap_krb5_init_creds` | `bool` | `true` | - |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - |
| `ldap_network_timeout` | `int` | `6` | - |
| `ldap_offline_timeout` | `int` | `-` | - |
| `ldap_opt_timeout` | `int` | `8` | - |
| `ldap_page_size` | `int` | `1000` | - |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop |
| `ldap_referrals` | `bool` | `true` | - |
| `ldap_rootdse_last_usn` | `string` | `-` | - |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_sasl_mech` | `string` | `not set` | - |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad |
| `ldap_search_base` | `string` | `If not set, the value of the` | - |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cert` | `string` | `not set` | - |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_key` | `string` | `not set` | - |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard |
| `ldap_uri` | `string` | `-` | - |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - |

## Section `[domain/ad/auth]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_ccachedir` | `string` | `/tmp` | - |
| `krb5_ccname_template` | `string` | `-` | - |
| `krb5_fast_principal` | `string` | `-` | - |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - |
| `krb5_keytab` | `string` | `System keytab, normally` | - |
| `krb5_lifetime` | `string` | `-` | - |
| `krb5_map_user` | `string` | `not set` | - |
| `krb5_renew_interval` | `string` | `-` | - |
| `krb5_renewable_lifetime` | `string` | `-` | - |
| `krb5_store_password_if_offline` | `bool` | `false` | - |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand |
| `krb5_use_subdomain_realm` | `bool` | `false` | - |
| `krb5_validate` | `bool` | `-` | - |
| `ldap_pwd_policy` | `string` | `none` | - |

## Section `[domain/ad/autofs]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_autofs_entry_key` | `string` | `-` | - |
| `ldap_autofs_entry_object_class` | `string` | `-` | - |
| `ldap_autofs_entry_value` | `string` | `-` | - |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - |
| `ldap_autofs_map_name` | `string` | `-` | - |
| `ldap_autofs_map_object_class` | `string` | `-` | - |
| `ldap_autofs_search_base` | `string` | `-` | - |

## Section `[domain/ad/chpass]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - |
| `krb5_kpasswd` | `string` | `Use the KDC` | - |

## Section `[domain/ad/id]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - |
| `ldap_force_upper_case_realm` | `bool` | `false` | - |
| `ldap_group_entry_usn` | `string` | `-` | - |
| `ldap_group_external_member` | `string` | `-` | - |
| `ldap_group_gid_number` | `string` | `-` | - |
| `ldap_group_member` | `string` | `-` | - |
| `ldap_group_modify_timestamp` | `string` | `-` | - |
| `ldap_group_name` | `string` | `-` | - |
| `ldap_group_nesting_level` | `int` | `2` | - |
| `ldap_group_object_class` | `string` | `-` | - |
| `ldap_group_objectsid` | `string` | `-` | - |
| `ldap_group_search_base` | `string` | `-` | - |
| `ldap_group_search_filter` | `string` | `-` | - |
| `ldap_group_search_scope` | `string` | `-` | - |
| `ldap_group_type` | `string` | `-` | - |
| `ldap_group_uuid` | `string` | `-` | - |
| `ldap_id_mapping` | `bool` | `false` | - |
| `ldap_id_use_start_tls` | `bool` | `true` | - |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - |
| `ldap_idmap_default_domain` | `string` | `-` | - |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - |
| `ldap_idmap_helper_table_size` | `int` | `-` | - |
| `ldap_idmap_range_max` | `int` | `-` | - |
| `ldap_idmap_range_min` | `int` | `-` | - |
| `ldap_idmap_range_size` | `int` | `-` | - |
| `ldap_netgroup_search_base` | `string` | `-` | - |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - |
| `ldap_pwd_attribute` | `string` | `-` | - |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - |
| `ldap_search_timeout` | `int` | `6` | - |
| `ldap_service_entry_usn` | `string` | `-` | - |
| `ldap_service_name` | `string` | `-` | - |
| `ldap_service_object_class` | `string` | `-` | - |
| `ldap_service_port` | `string` | `-` | - |
| `ldap_service_proto` | `string` | `-` | - |
| `ldap_service_search_base` | `string` | `-` | - |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - |
| `ldap_user_auth_type` | `string` | `-` | - |
| `ldap_user_certificate` | `string` | `-` | - |
| `ldap_user_email` | `string` | `-` | - |
| `ldap_user_entry_usn` | `string` | `-` | - |
| `ldap_user_extra_attrs` | `string` | `-` | - |
| `ldap_user_fullname` | `string` | `-` | - |
| `ldap_user_gecos` | `string` | `-` | - |
| `ldap_user_gid_number` | `string` | `-` | - |
| `ldap_user_home_directory` | `string` | `-` | - |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - |
| `ldap_user_krb_password_expiration` | `string` | `-` | - |
| `ldap_user_member_of` | `string` | `-` | - |
| `ldap_user_modify_timestamp` | `string` | `-` | - |
| `ldap_user_name` | `string` | `-` | - |
| `ldap_user_object_class` | `string` | `-` | - |
| `ldap_user_objectsid` | `string` | `-` | - |
| `ldap_user_passkey` | `string` | `-` | - |
| `ldap_user_primary_group` | `string` | `-` | - |
| `ldap_user_principal` | `string` | `-` | - |
| `ldap_user_search_base` | `string` | `-` | - |
| `ldap_user_search_filter` | `string` | `-` | - |
| `ldap_user_search_scope` | `string` | `-` | - |
| `ldap_user_shadow_expire` | `string` | `-` | - |
| `ldap_user_shadow_flag` | `string` | `-` | - |
| `ldap_user_shadow_inactive` | `string` | `-` | - |
| `ldap_user_shadow_last_change` | `string` | `-` | - |
| `ldap_user_shadow_max` | `string` | `-` | - |
| `ldap_user_shadow_min` | `string` | `-` | - |
| `ldap_user_shadow_warning` | `string` | `-` | - |
| `ldap_user_shell` | `string` | `-` | - |
| `ldap_user_ssh_public_key` | `string` | `-` | - |
| `ldap_user_uid_number` | `string` | `-` | - |
| `ldap_user_uuid` | `string` | `-` | - |

## Section `[domain/ad/resolver]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_iphost_entry_usn` | `string` | `-` | - |
| `ldap_iphost_name` | `string` | `-` | - |
| `ldap_iphost_number` | `string` | `-` | - |
| `ldap_iphost_object_class` | `string` | `-` | - |
| `ldap_iphost_search_base` | `string` | `-` | - |
| `ldap_ipnetwork_entry_usn` | `string` | `-` | - |
| `ldap_ipnetwork_name` | `string` | `-` | - |
| `ldap_ipnetwork_number` | `string` | `-` | - |
| `ldap_ipnetwork_object_class` | `string` | `-` | - |
| `ldap_ipnetwork_search_base` | `string` | `-` | - |

## Section `[domain/ad/sudo]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - |
| `ldap_sudo_hostnames` | `string` | `not specified` | - |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - |
| `ldap_sudo_include_regexp` | `bool` | `false` | - |
| `ldap_sudo_ip` | `string` | `not specified` | - |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - |
| `ldap_sudo_search_base` | `string` | `-` | - |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - |
| `ldap_sudorule_command` | `string` | `-` | - |
| `ldap_sudorule_host` | `string` | `-` | - |
| `ldap_sudorule_name` | `string` | `-` | - |
| `ldap_sudorule_notafter` | `string` | `-` | - |
| `ldap_sudorule_notbefore` | `string` | `-` | - |
| `ldap_sudorule_object_class` | `string` | `-` | - |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - |
| `ldap_sudorule_option` | `string` | `-` | - |
| `ldap_sudorule_order` | `string` | `-` | - |
| `ldap_sudorule_runas` | `string` | `-` | - |
| `ldap_sudorule_runasgroup` | `string` | `-` | - |
| `ldap_sudorule_runasuser` | `string` | `-` | - |
| `ldap_sudorule_user` | `string` | `-` | - |

## Section `[domain/ipa]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - |
| `dyndns_auth` | `string` | `GSS-TSIG` | - |
| `dyndns_auth_ptr` | `string` | `Same as dyndns_auth` | - |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - |
| `dyndns_update` | `bool` | `true` | - |
| `dyndns_update_per_family` | `bool` | `true` | - |
| `dyndns_update_ptr` | `bool` | `True` | - |
| `ipa_access_order` | `string` | `-` | - |
| `ipa_anchor_uuid` | `string` | `ipaAnchorUUID` | - |
| `ipa_automount_location` | `string` | `The location named "default"` | - |
| `ipa_backup_server` | `string` | `-` | - |
| `ipa_deskprofile_refresh` | `int` | `5 (seconds)` | - |
| `ipa_deskprofile_request_interval` | `int` | `60 (minutes)` | - |
| `ipa_deskprofile_search_base` | `string` | `Use base DN` | - |
| `ipa_domain` | `string` | `-` | - |
| `ipa_group_override_object_class` | `string` | `ipaGroupOverride` | - |
| `ipa_hbac_refresh` | `int` | `5 (seconds)` | - |
| `ipa_hbac_search_base` | `string` | `Use base DN` | - |
| `ipa_host_search_base` | `string` | `-` | - |
| `ipa_hostname` | `string` | `-` | - |
| `ipa_master_domain_search_base` | `string` | `the value of` | - |
| `ipa_override_object_class` | `string` | `ipaOverrideAnchor` | - |
| `ipa_ranges_search_base` | `string` | `-` | - |
| `ipa_selinux_refresh` | `int` | `5 (seconds)` | - |
| `ipa_selinux_search_base` | `string` | `the value of` | - |
| `ipa_server` | `string` | `-` | - |
| `ipa_server_mode` | `bool` | `false` | - |
| `ipa_subdomains_search_base` | `string` | `the value of` | - |
| `ipa_subid_ranges_search_base` | `string` | `-` | - |
| `ipa_user_override_object_class` | `string` | `ipaUserOverride` | - |
| `ipa_view_class` | `string` | `nsContainer` | - |
| `ipa_view_name` | `string` | `cn` | - |
| `ipa_views_search_base` | `string` | `the value of` | - |
| `krb5_auth_timeout` | `int` | `-` | - |
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - |
| `krb5_backup_server` | `string` | `-` | - |
| `krb5_canonicalize` | `bool` | `false` | - |
| `krb5_confd_path` | `string` | `not set (krb5.include.d subdirectory of` | - |
| `krb5_kdcip` | `string` | `-` | - |
| `krb5_kpasswd` | `string` | `Use the KDC` | - |
| `krb5_realm` | `string` | `System defaults, see` | - |
| `krb5_server` | `string` | `-` | - |
| `krb5_use_kdcinfo` | `bool` | `true` | - |
| `ldap_backup_uri` | `string` | `-` | - |
| `ldap_connection_expire_offset` | `int` | `0` | - |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_default_authtok` | `string` | `-` | - |
| `ldap_default_authtok_type` | `string` | `password` | - |
| `ldap_default_bind_dn` | `string` | `-` | - |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always |
| `ldap_deref_threshold` | `int` | `10` | - |
| `ldap_disable_paging` | `bool` | `False` | - |
| `ldap_dns_service_name` | `string` | `ldap` | - |
| `ldap_entry_usn` | `string` | `-` | - |
| `ldap_krb5_init_creds` | `bool` | `true` | - |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - |
| `ldap_network_timeout` | `int` | `6` | - |
| `ldap_offline_timeout` | `int` | `-` | - |
| `ldap_opt_timeout` | `int` | `8` | - |
| `ldap_page_size` | `int` | `1000` | - |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop |
| `ldap_referrals` | `bool` | `true` | - |
| `ldap_rootdse_last_usn` | `string` | `-` | - |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_sasl_mech` | `string` | `not set` | - |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad |
| `ldap_search_base` | `string` | `If not set, the value of the` | - |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cert` | `string` | `not set` | - |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_key` | `string` | `not set` | - |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard |
| `ldap_uri` | `string` | `-` | - |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - |

## Section `[domain/ipa/access]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_hbac_refresh` | `int` | `5 (seconds)` | - |
| `ipa_hbac_support_srchost` | `bool` | `-` | - |
| `ipa_host_fqdn` | `string` | `-` | - |
| `ipa_host_member_of` | `string` | `-` | - |
| `ipa_host_name` | `string` | `-` | - |
| `ipa_host_object_class` | `string` | `-` | - |
| `ipa_host_serverhostname` | `string` | `-` | - |
| `ipa_host_ssh_public_key` | `string` | `-` | - |
| `ipa_host_uuid` | `string` | `-` | - |
| `ipa_hostgroup_member` | `string` | `-` | - |
| `ipa_hostgroup_memberof` | `string` | `-` | - |
| `ipa_hostgroup_name` | `string` | `-` | - |
| `ipa_hostgroup_objectclass` | `string` | `-` | - |
| `ipa_hostgroup_uuid` | `string` | `-` | - |
| `ipa_selinux_refresh` | `int` | `5 (seconds)` | - |

## Section `[domain/ipa/auth]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_ccachedir` | `string` | `/tmp` | - |
| `krb5_ccname_template` | `string` | `-` | - |
| `krb5_fast_principal` | `string` | `-` | - |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - |
| `krb5_keytab` | `string` | `System keytab, normally` | - |
| `krb5_lifetime` | `string` | `-` | - |
| `krb5_map_user` | `string` | `not set` | - |
| `krb5_renew_interval` | `string` | `-` | - |
| `krb5_renewable_lifetime` | `string` | `-` | - |
| `krb5_store_password_if_offline` | `bool` | `false` | - |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand |
| `krb5_use_subdomain_realm` | `bool` | `false` | - |
| `krb5_validate` | `bool` | `-` | - |
| `ldap_pwd_policy` | `string` | `none` | - |

## Section `[domain/ipa/autofs]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_automount_location` | `string` | `The location named "default"` | - |
| `ldap_autofs_entry_key` | `string` | `-` | - |
| `ldap_autofs_entry_object_class` | `string` | `-` | - |
| `ldap_autofs_entry_value` | `string` | `-` | - |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - |
| `ldap_autofs_map_name` | `string` | `-` | - |
| `ldap_autofs_map_object_class` | `string` | `-` | - |
| `ldap_autofs_search_base` | `string` | `-` | - |

## Section `[domain/ipa/id]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_anchor_uuid` | `string` | `ipaAnchorUUID` | - |
| `ipa_group_override_object_class` | `string` | `ipaGroupOverride` | - |
| `ipa_host_fqdn` | `string` | `-` | - |
| `ipa_host_object_class` | `string` | `-` | - |
| `ipa_host_ssh_public_key` | `string` | `-` | - |
| `ipa_netgroup_domain` | `string` | `-` | - |
| `ipa_netgroup_member` | `string` | `-` | - |
| `ipa_netgroup_member_ext_host` | `string` | `-` | - |
| `ipa_netgroup_member_host` | `string` | `-` | - |
| `ipa_netgroup_member_of` | `string` | `-` | - |
| `ipa_netgroup_member_user` | `string` | `-` | - |
| `ipa_netgroup_name` | `string` | `-` | - |
| `ipa_netgroup_object_class` | `string` | `-` | - |
| `ipa_netgroup_uuid` | `string` | `-` | - |
| `ipa_override_object_class` | `string` | `ipaOverrideAnchor` | - |
| `ipa_server_mode` | `bool` | `false` | - |
| `ipa_user_override_object_class` | `string` | `ipaUserOverride` | - |
| `ipa_view_class` | `string` | `nsContainer` | - |
| `ipa_view_name` | `string` | `cn` | - |
| `ipa_views_search_base` | `string` | `the value of` | - |
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - |
| `ldap_force_upper_case_realm` | `bool` | `false` | - |
| `ldap_group_entry_usn` | `string` | `-` | - |
| `ldap_group_external_member` | `string` | `-` | - |
| `ldap_group_gid_number` | `string` | `-` | - |
| `ldap_group_member` | `string` | `-` | - |
| `ldap_group_modify_timestamp` | `string` | `-` | - |
| `ldap_group_name` | `string` | `-` | - |
| `ldap_group_nesting_level` | `int` | `2` | - |
| `ldap_group_object_class` | `string` | `-` | - |
| `ldap_group_objectsid` | `string` | `-` | - |
| `ldap_group_search_base` | `string` | `-` | - |
| `ldap_group_search_filter` | `string` | `-` | - |
| `ldap_group_search_scope` | `string` | `-` | - |
| `ldap_group_type` | `string` | `-` | - |
| `ldap_group_uuid` | `string` | `-` | - |
| `ldap_id_mapping` | `bool` | `false` | - |
| `ldap_id_use_start_tls` | `bool` | `true` | - |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - |
| `ldap_idmap_default_domain` | `string` | `-` | - |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - |
| `ldap_idmap_helper_table_size` | `int` | `-` | - |
| `ldap_idmap_range_max` | `int` | `-` | - |
| `ldap_idmap_range_min` | `int` | `-` | - |
| `ldap_idmap_range_size` | `int` | `-` | - |
| `ldap_netgroup_search_base` | `string` | `-` | - |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - |
| `ldap_pwd_attribute` | `string` | `-` | - |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - |
| `ldap_search_timeout` | `int` | `6` | - |
| `ldap_service_entry_usn` | `string` | `-` | - |
| `ldap_service_name` | `string` | `-` | - |
| `ldap_service_object_class` | `string` | `-` | - |
| `ldap_service_port` | `string` | `-` | - |
| `ldap_service_proto` | `string` | `-` | - |
| `ldap_service_search_base` | `string` | `-` | - |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - |
| `ldap_user_auth_type` | `string` | `-` | - |
| `ldap_user_certificate` | `string` | `-` | - |
| `ldap_user_email` | `string` | `-` | - |
| `ldap_user_entry_usn` | `string` | `-` | - |
| `ldap_user_extra_attrs` | `string` | `-` | - |
| `ldap_user_fullname` | `string` | `-` | - |
| `ldap_user_gecos` | `string` | `-` | - |
| `ldap_user_gid_number` | `string` | `-` | - |
| `ldap_user_home_directory` | `string` | `-` | - |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - |
| `ldap_user_krb_password_expiration` | `string` | `-` | - |
| `ldap_user_member_of` | `string` | `-` | - |
| `ldap_user_modify_timestamp` | `string` | `-` | - |
| `ldap_user_name` | `string` | `-` | - |
| `ldap_user_object_class` | `string` | `-` | - |
| `ldap_user_objectsid` | `string` | `-` | - |
| `ldap_user_passkey` | `string` | `-` | - |
| `ldap_user_primary_group` | `string` | `-` | - |
| `ldap_user_principal` | `string` | `-` | - |
| `ldap_user_search_base` | `string` | `-` | - |
| `ldap_user_search_filter` | `string` | `-` | - |
| `ldap_user_search_scope` | `string` | `-` | - |
| `ldap_user_shadow_expire` | `string` | `-` | - |
| `ldap_user_shadow_flag` | `string` | `-` | - |
| `ldap_user_shadow_inactive` | `string` | `-` | - |
| `ldap_user_shadow_last_change` | `string` | `-` | - |
| `ldap_user_shadow_max` | `string` | `-` | - |
| `ldap_user_shadow_min` | `string` | `-` | - |
| `ldap_user_shadow_warning` | `string` | `-` | - |
| `ldap_user_shell` | `string` | `-` | - |
| `ldap_user_ssh_public_key` | `string` | `-` | - |
| `ldap_user_uid_number` | `string` | `-` | - |
| `ldap_user_uuid` | `string` | `-` | - |

## Section `[domain/ipa/session]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_deskprofile_refresh` | `int` | `5 (seconds)` | - |
| `ipa_deskprofile_request_interval` | `int` | `60 (minutes)` | - |
| `ipa_host_fqdn` | `string` | `-` | - |
| `ipa_host_member_of` | `string` | `-` | - |
| `ipa_host_name` | `string` | `-` | - |
| `ipa_host_object_class` | `string` | `-` | - |
| `ipa_host_serverhostname` | `string` | `-` | - |
| `ipa_host_ssh_public_key` | `string` | `-` | - |
| `ipa_host_uuid` | `string` | `-` | - |
| `ipa_selinux_usermap_enabled` | `string` | `-` | - |
| `ipa_selinux_usermap_host_category` | `string` | `-` | - |
| `ipa_selinux_usermap_member_host` | `string` | `-` | - |
| `ipa_selinux_usermap_member_user` | `string` | `-` | - |
| `ipa_selinux_usermap_name` | `string` | `-` | - |
| `ipa_selinux_usermap_object_class` | `string` | `-` | - |
| `ipa_selinux_usermap_see_also` | `string` | `-` | - |
| `ipa_selinux_usermap_selinux_user` | `string` | `-` | - |
| `ipa_selinux_usermap_user_category` | `string` | `-` | - |
| `ipa_selinux_usermap_uuid` | `string` | `-` | - |

## Section `[domain/ipa/subdomains]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_subdomains_search_base` | `string` | `the value of` | - |

## Section `[domain/ipa/sudo]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ipa_sudocmd_memberof` | `string` | `-` | - |
| `ipa_sudocmd_object_class` | `string` | `-` | - |
| `ipa_sudocmd_sudocmd` | `string` | `-` | - |
| `ipa_sudocmd_uuid` | `string` | `-` | - |
| `ipa_sudocmdgroup_entry_usn` | `string` | `-` | - |
| `ipa_sudocmdgroup_member` | `string` | `-` | - |
| `ipa_sudocmdgroup_name` | `string` | `-` | - |
| `ipa_sudocmdgroup_object_class` | `string` | `-` | - |
| `ipa_sudocmdgroup_uuid` | `string` | `-` | - |
| `ipa_sudorule_allowcmd` | `string` | `-` | - |
| `ipa_sudorule_cmdcategory` | `string` | `-` | - |
| `ipa_sudorule_denycmd` | `string` | `-` | - |
| `ipa_sudorule_enabled_flag` | `string` | `-` | - |
| `ipa_sudorule_entry_usn` | `string` | `-` | - |
| `ipa_sudorule_externaluser` | `string` | `-` | - |
| `ipa_sudorule_host` | `string` | `-` | - |
| `ipa_sudorule_hostcategory` | `string` | `-` | - |
| `ipa_sudorule_name` | `string` | `-` | - |
| `ipa_sudorule_notafter` | `string` | `-` | - |
| `ipa_sudorule_notbefore` | `string` | `-` | - |
| `ipa_sudorule_object_class` | `string` | `-` | - |
| `ipa_sudorule_option` | `string` | `-` | - |
| `ipa_sudorule_runasextgroup` | `string` | `-` | - |
| `ipa_sudorule_runasextuser` | `string` | `-` | - |
| `ipa_sudorule_runasextusergroup` | `string` | `-` | - |
| `ipa_sudorule_runasgroup` | `string` | `-` | - |
| `ipa_sudorule_runasgroupcategory` | `string` | `-` | - |
| `ipa_sudorule_runasusercategory` | `string` | `-` | - |
| `ipa_sudorule_sudoorder` | `string` | `-` | - |
| `ipa_sudorule_user` | `string` | `-` | - |
| `ipa_sudorule_usercategory` | `string` | `-` | - |
| `ipa_sudorule_uuid` | `string` | `-` | - |
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - |
| `ldap_sudo_hostnames` | `string` | `not specified` | - |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - |
| `ldap_sudo_include_regexp` | `bool` | `false` | - |
| `ldap_sudo_ip` | `string` | `not specified` | - |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - |
| `ldap_sudo_search_base` | `string` | `-` | - |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - |
| `ldap_sudorule_command` | `string` | `-` | - |
| `ldap_sudorule_host` | `string` | `-` | - |
| `ldap_sudorule_name` | `string` | `-` | - |
| `ldap_sudorule_notafter` | `string` | `-` | - |
| `ldap_sudorule_notbefore` | `string` | `-` | - |
| `ldap_sudorule_object_class` | `string` | `-` | - |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - |
| `ldap_sudorule_option` | `string` | `-` | - |
| `ldap_sudorule_order` | `string` | `-` | - |
| `ldap_sudorule_runas` | `string` | `-` | - |
| `ldap_sudorule_runasgroup` | `string` | `-` | - |
| `ldap_sudorule_runasuser` | `string` | `-` | - |
| `ldap_sudorule_user` | `string` | `-` | - |

## Section `[domain/krb5]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_auth_timeout` | `int` | `-` | - |
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - |
| `krb5_backup_server` | `string` | `-` | - |
| `krb5_ccachedir` | `string` | `/tmp` | - |
| `krb5_ccname_template` | `string` | `-` | - |
| `krb5_fast_principal` | `string` | `-` | - |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - |
| `krb5_kdcinfo_lookahead` | `string` | `3:1` | - |
| `krb5_kdcip` | `string` | `-` | - |
| `krb5_keytab` | `string` | `System keytab, normally` | - |
| `krb5_kpasswd` | `string` | `Use the KDC` | - |
| `krb5_map_user` | `string` | `not set` | - |
| `krb5_realm` | `string` | `System defaults, see` | - |
| `krb5_server` | `string` | `-` | - |
| `krb5_store_password_if_offline` | `bool` | `false` | - |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand |
| `krb5_use_kdcinfo` | `bool` | `true` | - |
| `krb5_use_subdomain_realm` | `bool` | `false` | - |

## Section `[domain/krb5/auth]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_canonicalize` | `bool` | `false` | - |
| `krb5_ccachedir` | `string` | `/tmp` | - |
| `krb5_ccname_template` | `string` | `-` | - |
| `krb5_fast_principal` | `string` | `-` | - |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - |
| `krb5_keytab` | `string` | `System keytab, normally` | - |
| `krb5_lifetime` | `string` | `-` | - |
| `krb5_map_user` | `string` | `not set` | - |
| `krb5_renew_interval` | `string` | `-` | - |
| `krb5_renewable_lifetime` | `string` | `-` | - |
| `krb5_store_password_if_offline` | `bool` | `false` | - |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand |
| `krb5_use_subdomain_realm` | `bool` | `false` | - |
| `krb5_validate` | `bool` | `-` | - |

## Section `[domain/ldap]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `krb5_backup_server` | `string` | `-` | - |
| `krb5_canonicalize` | `bool` | `false` | - |
| `krb5_kdcip` | `string` | `-` | - |
| `krb5_realm` | `string` | `System defaults, see` | - |
| `krb5_server` | `string` | `-` | - |
| `krb5_use_kdcinfo` | `bool` | `true` | - |
| `ldap_access_filter` | `string` | `Empty` | - |
| `ldap_access_order` | `string` | `filter` | - |
| `ldap_account_expire_policy` | `string` | `Empty` | - |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - |
| `ldap_autofs_search_base` | `string` | `-` | - |
| `ldap_backup_uri` | `string` | `-` | - |
| `ldap_chpass_backup_uri` | `string` | `empty, i.e. ldap_uri is used.` | - |
| `ldap_chpass_dns_service_name` | `string` | `not set, i.e. service discovery is disabled` | - |
| `ldap_chpass_update_last_change` | `bool` | `False` | - |
| `ldap_chpass_uri` | `string` | `empty, i.e. ldap_uri is used.` | - |
| `ldap_connection_expire_offset` | `int` | `0` | - |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - |
| `ldap_default_authtok` | `string` | `-` | - |
| `ldap_default_authtok_type` | `string` | `password` | - |
| `ldap_default_bind_dn` | `string` | `-` | - |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always |
| `ldap_deref_threshold` | `int` | `10` | - |
| `ldap_disable_paging` | `bool` | `False` | - |
| `ldap_disable_range_retrieval` | `bool` | `False` | - |
| `ldap_dns_service_name` | `string` | `ldap` | - |
| `ldap_entry_usn` | `string` | `-` | - |
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - |
| `ldap_enumeration_search_timeout` | `int` | `60` | - |
| `ldap_force_upper_case_realm` | `bool` | `false` | - |
| `ldap_group_nesting_level` | `int` | `2` | - |
| `ldap_group_search_base` | `string` | `-` | - |
| `ldap_host_search_base` | `string` | `the value of` | - |
| `ldap_id_mapping` | `bool` | `false` | - |
| `ldap_id_use_start_tls` | `bool` | `true` | - |
| `ldap_ignore_unreadable_references` | `bool` | `False` | - |
| `ldap_iphost_search_base` | `string` | `-` | - |
| `ldap_ipnetwork_search_base` | `string` | `-` | - |
| `ldap_krb5_init_creds` | `bool` | `true` | - |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - |
| `ldap_library_debug_level` | `int` | `0 (libldap debugging disabled)` | - |
| `ldap_max_id` | `int` | `not set (both options are set to 0)` | - |
| `ldap_min_id` | `int` | `not set (both options are set to 0)` | - |
| `ldap_netgroup_search_base` | `string` | `-` | - |
| `ldap_network_timeout` | `int` | `6` | - |
| `ldap_offline_timeout` | `int` | `-` | - |
| `ldap_opt_timeout` | `int` | `8` | - |
| `ldap_page_size` | `int` | `1000` | - |
| `ldap_ppolicy_pwd_change_threshold` | `int` | `0` | - |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - |
| `ldap_pwd_policy` | `string` | `none` | - |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop |
| `ldap_read_rootdse` | `string` | `anonymous` | - |
| `ldap_referrals` | `bool` | `true` | - |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - |
| `ldap_rootdse_last_usn` | `string` | `-` | - |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - |
| `ldap_sasl_canonicalize` | `bool` | `false;` | - |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_sasl_mech` | `string` | `not set` | - |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - |
| `ldap_sasl_realm` | `string` | `the value of krb5_realm.` | - |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad |
| `ldap_search_base` | `string` | `If not set, the value of the` | - |
| `ldap_search_timeout` | `int` | `6` | - |
| `ldap_service_search_base` | `string` | `-` | - |
| `ldap_subid_ranges_search_base` | `string` | `the value of` | - |
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - |
| `ldap_sudo_hostnames` | `string` | `not specified` | - |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - |
| `ldap_sudo_include_regexp` | `bool` | `false` | - |
| `ldap_sudo_ip` | `string` | `not specified` | - |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - |
| `ldap_sudo_search_base` | `string` | `-` | - |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cert` | `string` | `not set` | - |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_key` | `string` | `not set` | - |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard |
| `ldap_uri` | `string` | `-` | - |
| `ldap_use_ppolicy` | `bool` | `true` | - |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - |
| `ldap_user_search_base` | `string` | `-` | - |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - |

## Section `[domain/ldap/access]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_access_filter` | `string` | `Empty` | - |
| `ldap_access_order` | `string` | `filter` | - |
| `ldap_account_expire_policy` | `string` | `Empty` | - |

## Section `[domain/ldap/auth]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_pwd_policy` | `string` | `none` | - |

## Section `[domain/ldap/autofs]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_autofs_entry_key` | `string` | `-` | - |
| `ldap_autofs_entry_object_class` | `string` | `-` | - |
| `ldap_autofs_entry_value` | `string` | `-` | - |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - |
| `ldap_autofs_map_name` | `string` | `-` | - |
| `ldap_autofs_map_object_class` | `string` | `-` | - |
| `ldap_autofs_search_base` | `string` | `-` | - |

## Section `[domain/ldap/chpass]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_chpass_backup_uri` | `string` | `empty, i.e. ldap_uri is used.` | - |
| `ldap_chpass_dns_service_name` | `string` | `not set, i.e. service discovery is disabled` | - |
| `ldap_chpass_update_last_change` | `bool` | `False` | - |
| `ldap_chpass_uri` | `string` | `empty, i.e. ldap_uri is used.` | - |

## Section `[domain/ldap/id]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - |
| `ldap_enumeration_search_timeout` | `int` | `60` | - |
| `ldap_force_upper_case_realm` | `bool` | `false` | - |
| `ldap_group_entry_usn` | `string` | `-` | - |
| `ldap_group_external_member` | `string` | `-` | - |
| `ldap_group_gid_number` | `string` | `-` | - |
| `ldap_group_member` | `string` | `-` | - |
| `ldap_group_modify_timestamp` | `string` | `-` | - |
| `ldap_group_name` | `string` | `-` | - |
| `ldap_group_nesting_level` | `int` | `2` | - |
| `ldap_group_object_class` | `string` | `-` | - |
| `ldap_group_objectsid` | `string` | `-` | - |
| `ldap_group_search_base` | `string` | `-` | - |
| `ldap_group_search_filter` | `string` | `-` | - |
| `ldap_group_search_scope` | `string` | `-` | - |
| `ldap_group_type` | `string` | `-` | - |
| `ldap_group_uuid` | `string` | `-` | - |
| `ldap_id_mapping` | `bool` | `false` | - |
| `ldap_id_use_start_tls` | `bool` | `true` | - |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - |
| `ldap_idmap_default_domain` | `string` | `-` | - |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - |
| `ldap_idmap_helper_table_size` | `int` | `-` | - |
| `ldap_idmap_range_max` | `int` | `-` | - |
| `ldap_idmap_range_min` | `int` | `-` | - |
| `ldap_idmap_range_size` | `int` | `-` | - |
| `ldap_library_debug_level` | `int` | `0 (libldap debugging disabled)` | - |
| `ldap_max_id` | `int` | `not set (both options are set to 0)` | - |
| `ldap_min_id` | `int` | `not set (both options are set to 0)` | - |
| `ldap_netgroup_member` | `string` | `-` | - |
| `ldap_netgroup_modify_timestamp` | `string` | `-` | - |
| `ldap_netgroup_name` | `string` | `-` | - |
| `ldap_netgroup_object_class` | `string` | `-` | - |
| `ldap_netgroup_search_base` | `string` | `-` | - |
| `ldap_netgroup_triple` | `string` | `-` | - |
| `ldap_ns_account_lock` | `string` | `-` | - |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - |
| `ldap_pwd_attribute` | `string` | `-` | - |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - |
| `ldap_search_timeout` | `int` | `6` | - |
| `ldap_service_entry_usn` | `string` | `-` | - |
| `ldap_service_name` | `string` | `-` | - |
| `ldap_service_object_class` | `string` | `-` | - |
| `ldap_service_port` | `string` | `-` | - |
| `ldap_service_proto` | `string` | `-` | - |
| `ldap_service_search_base` | `string` | `-` | - |
| `ldap_subid_ranges_search_base` | `string` | `the value of` | - |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - |
| `ldap_user_ad_account_expires` | `string` | `-` | - |
| `ldap_user_ad_user_account_control` | `string` | `-` | - |
| `ldap_user_auth_type` | `string` | `-` | - |
| `ldap_user_authorized_host` | `string` | `-` | - |
| `ldap_user_authorized_rhost` | `string` | `-` | - |
| `ldap_user_authorized_service` | `string` | `-` | - |
| `ldap_user_certificate` | `string` | `-` | - |
| `ldap_user_email` | `string` | `-` | - |
| `ldap_user_entry_usn` | `string` | `-` | - |
| `ldap_user_extra_attrs` | `string` | `-` | - |
| `ldap_user_fullname` | `string` | `-` | - |
| `ldap_user_gecos` | `string` | `-` | - |
| `ldap_user_gid_number` | `string` | `-` | - |
| `ldap_user_home_directory` | `string` | `-` | - |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - |
| `ldap_user_krb_password_expiration` | `string` | `-` | - |
| `ldap_user_member_of` | `string` | `-` | - |
| `ldap_user_modify_timestamp` | `string` | `-` | - |
| `ldap_user_name` | `string` | `-` | - |
| `ldap_user_nds_login_allowed_time_map` | `string` | `-` | - |
| `ldap_user_nds_login_disabled` | `string` | `-` | - |
| `ldap_user_nds_login_expiration_time` | `string` | `-` | - |
| `ldap_user_object_class` | `string` | `-` | - |
| `ldap_user_objectsid` | `string` | `-` | - |
| `ldap_user_passkey` | `string` | `-` | - |
| `ldap_user_primary_group` | `string` | `-` | - |
| `ldap_user_principal` | `string` | `-` | - |
| `ldap_user_search_base` | `string` | `-` | - |
| `ldap_user_search_filter` | `string` | `-` | - |
| `ldap_user_search_scope` | `string` | `-` | - |
| `ldap_user_shadow_expire` | `string` | `-` | - |
| `ldap_user_shadow_flag` | `string` | `-` | - |
| `ldap_user_shadow_inactive` | `string` | `-` | - |
| `ldap_user_shadow_last_change` | `string` | `-` | - |
| `ldap_user_shadow_max` | `string` | `-` | - |
| `ldap_user_shadow_min` | `string` | `-` | - |
| `ldap_user_shadow_warning` | `string` | `-` | - |
| `ldap_user_shell` | `string` | `-` | - |
| `ldap_user_ssh_public_key` | `string` | `-` | - |
| `ldap_user_uid_number` | `string` | `-` | - |
| `ldap_user_uuid` | `string` | `-` | - |

## Section `[domain/ldap/resolver]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_iphost_entry_usn` | `string` | `-` | - |
| `ldap_iphost_name` | `string` | `-` | - |
| `ldap_iphost_number` | `string` | `-` | - |
| `ldap_iphost_object_class` | `string` | `-` | - |
| `ldap_iphost_search_base` | `string` | `-` | - |
| `ldap_ipnetwork_entry_usn` | `string` | `-` | - |
| `ldap_ipnetwork_name` | `string` | `-` | - |
| `ldap_ipnetwork_number` | `string` | `-` | - |
| `ldap_ipnetwork_object_class` | `string` | `-` | - |
| `ldap_ipnetwork_search_base` | `string` | `-` | - |

## Section `[domain/ldap/sudo]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - |
| `ldap_sudo_hostnames` | `string` | `not specified` | - |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - |
| `ldap_sudo_include_regexp` | `bool` | `false` | - |
| `ldap_sudo_ip` | `string` | `not specified` | - |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - |
| `ldap_sudo_search_base` | `string` | `-` | - |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - |
| `ldap_sudorule_command` | `string` | `-` | - |
| `ldap_sudorule_host` | `string` | `-` | - |
| `ldap_sudorule_name` | `string` | `-` | - |
| `ldap_sudorule_notafter` | `string` | `-` | - |
| `ldap_sudorule_notbefore` | `string` | `-` | - |
| `ldap_sudorule_object_class` | `string` | `-` | - |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - |
| `ldap_sudorule_option` | `string` | `-` | - |
| `ldap_sudorule_order` | `string` | `-` | - |
| `ldap_sudorule_runas` | `string` | `-` | - |
| `ldap_sudorule_runasgroup` | `string` | `-` | - |
| `ldap_sudorule_runasuser` | `string` | `-` | - |
| `ldap_sudorule_user` | `string` | `-` | - |

## Section `[domain/proxy]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `proxy_max_children` | `int` | `10` | - |

## Section `[domain/proxy/auth]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `proxy_pam_target` | `string` | `not set by default, you have to take an` | - |

## Section `[domain/proxy/id]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `proxy_fast_alias` | `bool` | `false` | - |
| `proxy_lib_name` | `string` | `-` | - |

## Section `[domain/simple]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `simple_allow_groups` | `string` | `-` | - |
| `simple_allow_users` | `string` | `-` | - |
| `simple_deny_groups` | `string` | `-` | - |
| `simple_deny_users` | `string` | `-` | - |

## Section `[domain/simple/access]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `simple_allow_groups` | `string` | `-` | - |
| `simple_allow_users` | `string` | `-` | - |
| `simple_deny_groups` | `string` | `-` | - |
| `simple_deny_users` | `string` | `-` | - |

## Section `[ifp]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `allowed_uids` | `string` | `0, &sssd_user_name; (only root and SSSD` | - |
| `user_attributes` | `string` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell |

## Section `[kcm]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `max_ccache_size` | `int` | `65536` | - |
| `max_ccaches` | `int` | `0 (unlimited, only the per-UID quota is enforced)` | - |
| `max_uid_ccaches` | `int` | `64` | - |
| `socket_path` | `string` | `-` | - |
| `tgt_renewal` | `bool` | `False (Automatic renewals disabled)` | - |
| `tgt_renewal_inherit` | `string` | `NULL` | - |

## Section `[nss]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `allowed_shells` | `list` | `Not set. The user shell is automatically used.` | - |
| `default_shell` | `string` | `not set (Return NULL if no shell is` | - |
| `entry_cache_nowait_percentage` | `int` | `50` | - |
| `entry_negative_timeout` | `int` | `15` | - |
| `enum_cache_timeout` | `int` | `120` | - |
| `fallback_homedir` | `string` | `not set (no substitution for unset home` | - |
| `filter_groups` | `list` | `root` | - |
| `filter_users` | `list` | `root` | - |
| `filter_users_in_groups` | `bool` | `true` | - |
| `get_domains_timeout` | `int` | `60` | - |
| `homedir_substring` | `string` | `/home` | - |
| `memcache_size_group` | `int` | `6` | - |
| `memcache_size_initgroups` | `int` | `10` | - |
| `memcache_size_passwd` | `int` | `8` | - |
| `memcache_size_sid` | `int` | `6` | - |
| `memcache_timeout` | `int` | `300` | - |
| `override_homedir` | `string` | `-` | - |
| `override_shell` | `string` | `not set (SSSD will use the value` | - |
| `pwfield` | `string` | `-` | - |
| `shell_fallback` | `string` | `/bin/sh` | - |
| `user_attributes` | `string` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell |
| `vetoed_shells` | `list` | `Not set` | - |

## Section `[pac]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `allowed_uids` | `string` | `0, &sssd_user_name; (only root and SSSD` | - |
| `pac_check` | `string` | `-` | no_check, pac_present, check_upn, check_upn_allow_missing, upn_dns_info_present, check_upn_dns_info_ex, upn_dns_info_ex_present |
| `pac_lifetime` | `int` | `300` | - |

## Section `[pam]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `get_domains_timeout` | `int` | `60` | - |
| `offline_credentials_expiration` | `int` | `0 (No limit)` | - |
| `offline_failed_login_attempts` | `int` | `0 (No limit)` | - |
| `offline_failed_login_delay` | `int` | `5` | - |
| `p11_child_timeout` | `int` | `10` | - |
| `p11_uri` | `string` | `none` | - |
| `p11_wait_for_card_timeout` | `int` | `60` | - |
| `pam_account_expired_message` | `string` | `none` | - |
| `pam_account_locked_message` | `string` | `none` | - |
| `pam_app_services` | `string` | `Not set` | - |
| `pam_cert_auth` | `bool` | `False` | - |
| `pam_cert_db_path` | `string` | `-` | - |
| `pam_cert_verification` | `string` | `not set, i.e. use default` | - |
| `pam_gssapi_check_upn` | `bool` | `True` | - |
| `pam_gssapi_indicators_apply` | `string` | `not set` | - |
| `pam_gssapi_indicators_map` | `string` | `not set (use of authentication indicators is not required)` | - |
| `pam_gssapi_services` | `string` | `- (GSSAPI authentication is disabled)` | - |
| `pam_id_timeout` | `int` | `5` | - |
| `pam_initgroups_scheme` | `string` | `-` | always, no_session, never |
| `pam_json_services` | `string` | `- (JSON protocol is disabled)` | - |
| `pam_p11_allowed_services` | `string` | `the default set of PAM service names` | - |
| `pam_passkey_auth` | `bool` | `True` | - |
| `pam_public_domains` | `string` | `none` | - |
| `pam_pwd_expiration_warning` | `int` | `0` | - |
| `pam_response_filter` | `string` | `-` | env |
| `pam_trusted_users` | `string` | `All users are considered trusted` | - |
| `pam_verbosity` | `int` | `1` | - |
| `passkey_child_timeout` | `int` | `15` | - |
| `passkey_debug_libfido2` | `bool` | `False` | - |

## Section `[service]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `cache_first` | `int` | `true` | - |
| `client_idle_timeout` | `int` | `60, KCM: 300` | - |
| `command` | `string` | `-` | - |
| `debug` | `int` | `-` | - |
| `debug_backtrace_enabled` | `bool` | `true` | - |
| `debug_level` | `int` | `-` | - |
| `debug_microseconds` | `bool` | `false` | - |
| `debug_timestamps` | `bool` | `true` | - |
| `description` | `string` | `-` | - |
| `fd_limit` | `int` | `8192 (or limits.conf "hard" limit)` | - |
| `responder_idle_timeout` | `int` | `300` | - |
| `timeout` | `int` | `10` | - |

## Section `[session_recording]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `exclude_groups` | `list` | `Empty. No groups excluded.` | - |
| `exclude_users` | `list` | `Empty. No users excluded.` | - |
| `groups` | `list` | `Empty. Matches no groups.` | - |
| `scope` | `string` | `-` | - |
| `users` | `list` | `Empty. Matches no users.` | - |

## Section `[ssh]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `ca_db` | `string` | `-` | - |
| `ssh_hash_known_hosts` | `bool` | `-` | - |
| `ssh_known_hosts_timeout` | `int` | `-` | - |
| `ssh_use_certificate_keys` | `bool` | `true` | - |
| `ssh_use_certificate_matching_rules` | `string` | `not set, equivalent to 'all_rules',` | - |

## Section `[sssd]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `certificate_verification` | `string` | `-` | no_verification, ocsp_dgst, ocsp, crl |
| `config_file_version` | `int` | `-` | - |
| `core_dumpable` | `bool` | `true` | - |
| `default_domain_suffix` | `string` | `not set` | - |
| `disable_netlink` | `bool` | `false (netlink changes are detected)` | - |
| `domain_resolution_order` | `list` | `Not set` | - |
| `domains` | `list` | `the configured domain in sssd.conf` | - |
| `enable_files_domain` | `string` | `-` | - |
| `full_name_format` | `string` | `-` | - |
| `implicit_pac_responder` | `bool` | `true` | - |
| `krb5_rcache_dir` | `string` | `Distribution-specific and specified` | - |
| `monitor_resolv_conf` | `bool` | `true` | - |
| `override_space` | `string` | `not set (spaces will not be replaced)` | - |
| `passkey_verification` | `string` | `-` | user_verification |
| `re_expression` | `string` | `-` | - |
| `services` | `list` | `nss` | - |
| `try_inotify` | `bool` | `true on platforms where inotify is` | - |
| `user` | `string` | `-` | - |

## Section `[sudo]`

| Option | Type | Default | Allowed Values |
|--------|------|---------|----------------|
| `sudo_inverse_order` | `bool` | `-` | - |
| `sudo_threshold` | `int` | `50` | - |
| `sudo_timed` | `bool` | `false` | - |

## All Options (Alphabetical)

| Option | Type | Sections | Default | Allowed Values |
|--------|------|----------|---------|----------------|
| `access_provider` | `string` | `domain` | `-` | permit, deny, ad, ipa, ldap, simple, proxy |
| `account_cache_expiration` | `int` | `domain` | `0 (unlimited)` | - |
| `ad_access_filter` | `string` | `domain/ad` | `Not set` | - |
| `ad_backup_server` | `string` | `domain/ad` | `-` | - |
| `ad_domain` | `string` | `domain/ad` | `-` | - |
| `ad_enable_dns_sites` | `bool` | `domain/ad` | `true` | - |
| `ad_enable_gc` | `bool` | `domain/ad` | `true` | - |
| `ad_enabled_domains` | `string` | `domain/ad` | `Not set` | - |
| `ad_gpo_access_control` | `string` | `domain/ad` | `permissive` | disabled, enforcing, permissive |
| `ad_gpo_cache_timeout` | `int` | `domain/ad` | `5 (seconds)` | - |
| `ad_gpo_default_right` | `string` | `domain/ad` | `deny` | interactive, remote_interactive, network, batch, service, permit, deny |
| `ad_gpo_ignore_unreadable` | `bool` | `domain/ad` | `False` | - |
| `ad_gpo_implicit_deny` | `bool` | `domain/ad` | `False` | - |
| `ad_gpo_map_batch` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_deny` | `string` | `domain/ad` | `not set` | - |
| `ad_gpo_map_interactive` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_network` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_permit` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_remote_interactive` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - |
| `ad_gpo_map_service` | `string` | `domain/ad` | `not set` | - |
| `ad_hostname` | `string` | `domain/ad` | `-` | - |
| `ad_machine_account_password_renewal_opts` | `string` | `domain/ad` | `86400:750:300:realm (24h, 12m30s and 5m)` | - |
| `ad_maximum_machine_account_password_age` | `int` | `domain/ad` | `30 days` | - |
| `ad_server` | `string` | `domain/ad` | `-` | - |
| `ad_site` | `string` | `domain/ad` | `Not set` | - |
| `ad_update_samba_machine_account_password` | `bool` | `domain/ad` | `false` | - |
| `ad_use_ldaps` | `bool` | `domain/ad` | `False` | - |
| `allowed_shells` | `list` | `nss` | `Not set. The user shell is automatically used.` | - |
| `allowed_uids` | `string` | `pac, ifp` | `0, &sssd_user_name; (only root and SSSD` | - |
| `auth_provider` | `string` | `domain` | `-` | none, ad, ipa, ldap, krb5, proxy |
| `auto_private_groups` | `string` | `domain` | `-` | true, false, hybrid |
| `autofs_negative_timeout` | `int` | `autofs` | `15` | - |
| `autofs_provider` | `string` | `domain` | `The value of` | none, ad, ipa, ldap |
| `avoid_by_id_lookups` | `bool` | `domain` | `False (True for IdP provider)` | - |
| `ca_db` | `string` | `ssh` | `-` | - |
| `cache_credentials` | `bool` | `domain` | `FALSE` | - |
| `cache_credentials_minimal_first_factor_length` | `int` | `domain` | `8` | - |
| `cache_first` | `int` | `service, *` | `true` | - |
| `cached_auth_timeout` | `int` | `domain` | `0` | - |
| `case_sensitive` | `string` | `domain` | `-` | true, false, preserving |
| `certificate_verification` | `string` | `sssd` | `-` | no_verification, ocsp_dgst, ocsp, crl |
| `chpass_provider` | `string` | `domain` | `-` | none, ad, ipa, ldap, krb5, proxy |
| `client_idle_timeout` | `int` | `service, *` | `60, KCM: 300` | - |
| `command` | `string` | `service, domain` | `-` | - |
| `config_file_version` | `int` | `sssd` | `-` | - |
| `core_dumpable` | `bool` | `sssd` | `true` | - |
| `debug` | `int` | `service, domain, *` | `-` | - |
| `debug_backtrace_enabled` | `bool` | `service, *` | `true` | - |
| `debug_level` | `int` | `service, domain, *` | `-` | - |
| `debug_microseconds` | `bool` | `service, *` | `false` | - |
| `debug_timestamps` | `bool` | `service, domain, *` | `true` | - |
| `default_domain_suffix` | `string` | `sssd` | `not set` | - |
| `default_shell` | `string` | `nss, domain` | `not set (Return NULL if no shell is` | - |
| `description` | `string` | `service, domain` | `-` | - |
| `disable_netlink` | `bool` | `sssd` | `false (netlink changes are detected)` | - |
| `dns_discovery_domain` | `string` | `domain` | `Use the domain part of machine's hostname` | - |
| `dns_resolver_op_timeout` | `int` | `domain` | `3` | - |
| `dns_resolver_server_timeout` | `int` | `domain` | `1000` | - |
| `dns_resolver_timeout` | `int` | `domain` | `6` | - |
| `dns_resolver_use_search_list` | `bool` | `domain` | `TRUE` | - |
| `domain_resolution_order` | `list` | `sssd` | `Not set` | - |
| `domain_type` | `string` | `domain` | `posix` | - |
| `domains` | `list` | `sssd, certmap` | `the configured domain in sssd.conf` | - |
| `dyndns_address` | `string` | `domain, domain/ad, domain/ipa` | `No filtering of IP addresses.` | - |
| `dyndns_auth` | `string` | `domain, domain/ad, domain/ipa` | `GSS-TSIG` | - |
| `dyndns_auth_ptr` | `string` | `domain/ad, domain/ipa` | `Same as dyndns_auth` | - |
| `dyndns_dot_cacert` | `string` | `domain, domain/ad, domain/ipa` | `None (use global certificate store)` | - |
| `dyndns_dot_cert` | `string` | `domain, domain/ad, domain/ipa` | `None (Do not use TLS authentication)` | - |
| `dyndns_dot_key` | `string` | `domain, domain/ad, domain/ipa` | `None (Do not use TLS authentication)` | - |
| `dyndns_force_tcp` | `bool` | `domain, domain/ad, domain/ipa` | `False (let nsupdate choose the protocol)` | - |
| `dyndns_iface` | `string` | `domain, domain/ad, domain/ipa` | `Use the IP addresses of the interface which` | - |
| `dyndns_refresh_interval` | `int` | `domain, domain/ad, domain/ipa` | `86400 (24 hours)` | - |
| `dyndns_refresh_interval_offset` | `int` | `domain` | `-` | - |
| `dyndns_server` | `string` | `domain, domain/ad, domain/ipa` | `None (let nsupdate choose the server)` | - |
| `dyndns_ttl` | `int` | `domain, domain/ad, domain/ipa` | `3600 (seconds)` | - |
| `dyndns_update` | `bool` | `domain, domain/ad, domain/ipa` | `true` | - |
| `dyndns_update_per_family` | `bool` | `domain, domain/ad, domain/ipa` | `true` | - |
| `dyndns_update_ptr` | `bool` | `domain, domain/ad, domain/ipa` | `True` | - |
| `enable_files_domain` | `string` | `sssd` | `-` | - |
| `enabled` | `bool` | `domain` | `-` | - |
| `entry_cache_autofs_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_group_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_netgroup_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_nowait_percentage` | `int` | `nss` | `50` | - |
| `entry_cache_resolver_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_service_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_ssh_host_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_sudo_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_cache_timeout` | `int` | `domain` | `5400` | - |
| `entry_cache_user_timeout` | `int` | `domain` | `entry_cache_timeout` | - |
| `entry_negative_timeout` | `int` | `nss` | `15` | - |
| `enum_cache_timeout` | `int` | `nss` | `120` | - |
| `enumerate` | `bool` | `domain` | `FALSE` | - |
| `exclude_groups` | `list` | `session_recording` | `Empty. No groups excluded.` | - |
| `exclude_users` | `list` | `session_recording` | `Empty. No users excluded.` | - |
| `failover_primary_timeout` | `int` | `domain` | `31` | - |
| `fallback_homedir` | `string` | `nss, domain` | `not set (no substitution for unset home` | - |
| `fd_limit` | `int` | `service, *` | `8192 (or limits.conf "hard" limit)` | - |
| `filter_groups` | `list` | `nss, domain` | `root` | - |
| `filter_users` | `list` | `nss, domain` | `root` | - |
| `filter_users_in_groups` | `bool` | `nss` | `true` | - |
| `full_name_format` | `string` | `sssd, domain` | `-` | - |
| `get_domains_timeout` | `int` | `nss, pam` | `60` | - |
| `groups` | `list` | `session_recording` | `Empty. Matches no groups.` | - |
| `homedir_substring` | `string` | `nss, domain` | `/home` | - |
| `hostid_provider` | `string` | `domain` | `The value of` | - |
| `id_provider` | `string` | `domain` | `Not set (This is a mandatory setting that` | ad, ipa, ldap, proxy, simple, files |
| `idmap_range_max` | `int` | `domain` | `2000200000` | - |
| `idmap_range_min` | `int` | `domain` | `200000` | - |
| `idmap_range_size` | `int` | `domain` | `200000` | - |
| `idp_auth_scope` | `string` | `domain` | `Not set` | - |
| `idp_auto_refresh` | `bool` | `domain` | `false` | - |
| `idp_client_id` | `string` | `domain` | `Not set (Required)` | - |
| `idp_client_secret` | `string` | `domain` | `Not set` | - |
| `idp_device_auth_endpoint` | `string` | `domain` | `Not set` | - |
| `idp_id_scope` | `string` | `domain` | `Not set` | - |
| `idp_request_timeout` | `int` | `domain` | `10` | - |
| `idp_token_endpoint` | `string` | `domain` | `Not set (Required)` | - |
| `idp_type` | `string` | `domain` | `Not set (Required)` | - |
| `idp_userinfo_endpoint` | `string` | `domain` | `Not set` | - |
| `ignore_group_members` | `bool` | `domain` | `FALSE` | - |
| `implicit_pac_responder` | `bool` | `sssd` | `true` | - |
| `inherit_from` | `string` | `domain` | `Not set` | - |
| `ipa_access_order` | `string` | `domain/ipa` | `-` | - |
| `ipa_anchor_uuid` | `string` | `domain/ipa/id, domain/ipa` | `ipaAnchorUUID` | - |
| `ipa_automount_location` | `string` | `domain/ipa/autofs, domain/ipa` | `The location named "default"` | - |
| `ipa_backup_server` | `string` | `domain/ipa` | `-` | - |
| `ipa_deskprofile_refresh` | `int` | `domain/ipa/session, domain/ipa` | `5 (seconds)` | - |
| `ipa_deskprofile_request_interval` | `int` | `domain/ipa/session, domain/ipa` | `60 (minutes)` | - |
| `ipa_deskprofile_search_base` | `string` | `domain/ipa` | `Use base DN` | - |
| `ipa_domain` | `string` | `domain/ipa` | `-` | - |
| `ipa_group_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaGroupOverride` | - |
| `ipa_hbac_refresh` | `int` | `domain/ipa/access, domain/ipa` | `5 (seconds)` | - |
| `ipa_hbac_search_base` | `string` | `domain/ipa` | `Use base DN` | - |
| `ipa_hbac_support_srchost` | `bool` | `domain/ipa/access` | `-` | - |
| `ipa_host_fqdn` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_member_of` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_name` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_object_class` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_search_base` | `string` | `domain/ipa` | `-` | - |
| `ipa_host_serverhostname` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_ssh_public_key` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_host_uuid` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - |
| `ipa_hostgroup_member` | `string` | `domain/ipa/access` | `-` | - |
| `ipa_hostgroup_memberof` | `string` | `domain/ipa/access` | `-` | - |
| `ipa_hostgroup_name` | `string` | `domain/ipa/access` | `-` | - |
| `ipa_hostgroup_objectclass` | `string` | `domain/ipa/access` | `-` | - |
| `ipa_hostgroup_uuid` | `string` | `domain/ipa/access` | `-` | - |
| `ipa_hostname` | `string` | `domain/ipa` | `-` | - |
| `ipa_master_domain_search_base` | `string` | `domain/ipa` | `the value of` | - |
| `ipa_netgroup_domain` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_member` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_member_ext_host` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_member_host` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_member_of` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_member_user` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_name` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_object_class` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_netgroup_uuid` | `string` | `domain/ipa/id` | `-` | - |
| `ipa_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaOverrideAnchor` | - |
| `ipa_ranges_search_base` | `string` | `domain/ipa` | `-` | - |
| `ipa_selinux_refresh` | `int` | `domain/ipa/access, domain/ipa` | `5 (seconds)` | - |
| `ipa_selinux_search_base` | `string` | `domain/ipa` | `the value of` | - |
| `ipa_selinux_usermap_enabled` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_host_category` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_member_host` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_member_user` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_name` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_object_class` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_see_also` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_selinux_user` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_user_category` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_selinux_usermap_uuid` | `string` | `domain/ipa/session` | `-` | - |
| `ipa_server` | `string` | `domain/ipa` | `-` | - |
| `ipa_server_mode` | `bool` | `domain/ipa/id, domain/ipa` | `false` | - |
| `ipa_subdomains_search_base` | `string` | `domain/ipa/subdomains, domain/ipa` | `the value of` | - |
| `ipa_subid_ranges_search_base` | `string` | `domain/ipa` | `-` | - |
| `ipa_sudocmd_memberof` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmd_object_class` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmd_sudocmd` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmd_uuid` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmdgroup_entry_usn` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmdgroup_member` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmdgroup_name` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmdgroup_object_class` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudocmdgroup_uuid` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_allowcmd` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_cmdcategory` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_denycmd` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_enabled_flag` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_entry_usn` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_externaluser` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_host` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_hostcategory` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_name` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_notafter` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_notbefore` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_object_class` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_option` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasextgroup` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasextuser` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasextusergroup` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasgroup` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasgroupcategory` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_runasusercategory` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_sudoorder` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_user` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_usercategory` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_sudorule_uuid` | `string` | `domain/ipa/sudo` | `-` | - |
| `ipa_user_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaUserOverride` | - |
| `ipa_view_class` | `string` | `domain/ipa/id, domain/ipa` | `nsContainer` | - |
| `ipa_view_name` | `string` | `domain/ipa/id, domain/ipa` | `cn` | - |
| `ipa_views_search_base` | `string` | `domain/ipa/id, domain/ipa` | `the value of` | - |
| `krb5_auth_timeout` | `int` | `domain/ad, domain/ipa, domain/krb5` | `-` | - |
| `krb5_backup_kpasswd` | `string` | `domain/ad/chpass, domain/ipa, domain/krb5` | `Use the KDC` | - |
| `krb5_backup_server` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - |
| `krb5_canonicalize` | `bool` | `domain/ad, domain/ipa, domain/krb5/auth, domain/ldap` | `false` | - |
| `krb5_ccachedir` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `/tmp` | - |
| `krb5_ccname_template` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `-` | - |
| `krb5_confd_path` | `string` | `domain/ad, domain/ipa` | `not set (krb5.include.d subdirectory of` | - |
| `krb5_fast_principal` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `-` | - |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - |
| `krb5_kdcinfo_lookahead` | `string` | `domain/krb5` | `3:1` | - |
| `krb5_kdcip` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - |
| `krb5_keytab` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `System keytab, normally` | - |
| `krb5_kpasswd` | `string` | `domain/ad/chpass, domain/ipa, domain/krb5` | `Use the KDC` | - |
| `krb5_lifetime` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - |
| `krb5_map_user` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `not set` | - |
| `krb5_rcache_dir` | `string` | `sssd` | `Distribution-specific and specified` | - |
| `krb5_realm` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `System defaults, see` | - |
| `krb5_renew_interval` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - |
| `krb5_renewable_lifetime` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - |
| `krb5_server` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - |
| `krb5_store_password_if_offline` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - |
| `krb5_use_enterprise_principal` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false (AD provider: true)` | - |
| `krb5_use_fast` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `not set, i.e. FAST is not used.` | never, try, demand |
| `krb5_use_kdcinfo` | `bool` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `true` | - |
| `krb5_use_subdomain_realm` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - |
| `krb5_validate` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - |
| `ldap_access_filter` | `string` | `domain/ldap/access, domain/ldap` | `Empty` | - |
| `ldap_access_order` | `string` | `domain/ldap/access, domain/ldap` | `filter` | - |
| `ldap_account_expire_policy` | `string` | `domain/ldap/access, domain/ldap` | `Empty` | - |
| `ldap_autofs_entry_key` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - |
| `ldap_autofs_entry_object_class` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - |
| `ldap_autofs_entry_value` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - |
| `ldap_autofs_map_master_name` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs, domain/ldap` | `auto.master` | - |
| `ldap_autofs_map_name` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - |
| `ldap_autofs_map_object_class` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - |
| `ldap_autofs_search_base` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs, domain/ldap` | `-` | - |
| `ldap_backup_uri` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_chpass_backup_uri` | `string` | `domain/ldap/chpass, domain/ldap` | `empty, i.e. ldap_uri is used.` | - |
| `ldap_chpass_dns_service_name` | `string` | `domain/ldap/chpass, domain/ldap` | `not set, i.e. service discovery is disabled` | - |
| `ldap_chpass_update_last_change` | `bool` | `domain/ldap/chpass, domain/ldap` | `False` | - |
| `ldap_chpass_uri` | `string` | `domain/ldap/chpass, domain/ldap` | `empty, i.e. ldap_uri is used.` | - |
| `ldap_connection_expire_offset` | `int` | `domain/ad, domain/ipa, domain/ldap` | `0` | - |
| `ldap_connection_expire_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `900 (15 minutes)` | - |
| `ldap_connection_idle_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `900 (15 minutes)` | - |
| `ldap_default_authtok` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_default_authtok_type` | `string` | `domain/ad, domain/ipa, domain/ldap` | `password` | - |
| `ldap_default_bind_dn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_deref` | `string` | `domain/ad, domain/ipa, domain/ldap` | `Empty (this is handled as` | never, searching, finding, always |
| `ldap_deref_threshold` | `int` | `domain/ad, domain/ipa, domain/ldap` | `10` | - |
| `ldap_disable_paging` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `False` | - |
| `ldap_disable_range_retrieval` | `bool` | `domain/ldap` | `False` | - |
| `ldap_dns_service_name` | `string` | `domain/ad, domain/ipa, domain/ldap` | `ldap` | - |
| `ldap_entry_usn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_enumeration_refresh_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `300` | - |
| `ldap_enumeration_search_timeout` | `int` | `domain/ldap/id, domain/ldap` | `60` | - |
| `ldap_force_upper_case_realm` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - |
| `ldap_group_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_external_member` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_gid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_member` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_modify_timestamp` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_nesting_level` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `2` | - |
| `ldap_group_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_objectsid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - |
| `ldap_group_search_filter` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_search_scope` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_type` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_group_uuid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_host_search_base` | `string` | `domain/ldap` | `the value of` | - |
| `ldap_id_mapping` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - |
| `ldap_id_use_start_tls` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `true` | - |
| `ldap_idmap_autorid_compat` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_default_domain` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_default_domain_sid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_helper_table_size` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_range_max` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_range_min` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_idmap_range_size` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_ignore_unreadable_references` | `bool` | `domain/ldap` | `False` | - |
| `ldap_iphost_entry_usn` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_iphost_name` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_iphost_number` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_iphost_object_class` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_iphost_search_base` | `string` | `domain/ad/resolver, domain/ldap/resolver, domain/ldap` | `-` | - |
| `ldap_ipnetwork_entry_usn` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_ipnetwork_name` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_ipnetwork_number` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_ipnetwork_object_class` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - |
| `ldap_ipnetwork_search_base` | `string` | `domain/ad/resolver, domain/ldap/resolver, domain/ldap` | `-` | - |
| `ldap_krb5_init_creds` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `true` | - |
| `ldap_krb5_keytab` | `string` | `domain/ad, domain/ipa, domain/ldap` | `System keytab, normally` | - |
| `ldap_krb5_ticket_lifetime` | `int` | `domain/ad, domain/ipa, domain/ldap` | `86400 (24 hours)` | - |
| `ldap_library_debug_level` | `int` | `domain/ldap/id, domain/ldap` | `0 (libldap debugging disabled)` | - |
| `ldap_max_id` | `int` | `domain/ldap/id, domain/ldap` | `not set (both options are set to 0)` | - |
| `ldap_min_id` | `int` | `domain/ldap/id, domain/ldap` | `not set (both options are set to 0)` | - |
| `ldap_netgroup_member` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_netgroup_modify_timestamp` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_netgroup_name` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_netgroup_object_class` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_netgroup_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - |
| `ldap_netgroup_triple` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_network_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `6` | - |
| `ldap_ns_account_lock` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_offline_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_opt_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `8` | - |
| `ldap_page_size` | `int` | `domain/ad, domain/ipa, domain/ldap` | `1000` | - |
| `ldap_ppolicy_pwd_change_threshold` | `int` | `domain/ldap` | `0` | - |
| `ldap_purge_cache_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `0 (disabled)` | - |
| `ldap_pwd_attribute` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_pwd_policy` | `string` | `domain/ad/auth, domain/ipa/auth, domain/ldap/auth, domain/ldap` | `none` | - |
| `ldap_pwdlockout_dn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `cn=ppolicy,ou=policies,$ldap_search_base` | - |
| `ldap_pwmodify_mode` | `string` | `domain/ad, domain/ipa, domain/ldap` | `exop` | ldap_modify, exop |
| `ldap_read_rootdse` | `string` | `domain/ldap` | `anonymous` | - |
| `ldap_referrals` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `true` | - |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - |
| `ldap_rootdse_last_usn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_sasl_authid` | `string` | `domain/ad, domain/ipa, domain/ldap` | `host/hostname@REALM` | - |
| `ldap_sasl_canonicalize` | `bool` | `domain/ldap` | `false;` | - |
| `ldap_sasl_maxssf` | `int` | `domain/ad, domain/ipa, domain/ldap` | `Use the system default (usually specified` | - |
| `ldap_sasl_mech` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - |
| `ldap_sasl_minssf` | `int` | `domain/ad, domain/ipa, domain/ldap` | `Use the system default (usually specified` | - |
| `ldap_sasl_realm` | `string` | `domain/ldap` | `the value of krb5_realm.` | - |
| `ldap_schema` | `string` | `domain/ad, domain/ipa, domain/ldap` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad |
| `ldap_search_base` | `string` | `domain/ad, domain/ipa, domain/ldap` | `If not set, the value of the` | - |
| `ldap_search_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `6` | - |
| `ldap_service_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_service_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_service_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_service_port` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_service_proto` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_service_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - |
| `ldap_subid_ranges_search_base` | `string` | `domain/ldap/id, domain/ldap` | `the value of` | - |
| `ldap_sudo_full_refresh_interval` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `21600 (6 hours)` | - |
| `ldap_sudo_hostnames` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `not specified` | - |
| `ldap_sudo_include_netgroups` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `true` | - |
| `ldap_sudo_include_regexp` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `false` | - |
| `ldap_sudo_ip` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `not specified` | - |
| `ldap_sudo_random_offset` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `0 (disabled)` | - |
| `ldap_sudo_search_base` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `-` | - |
| `ldap_sudo_smart_refresh_interval` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `900 (15 minutes)` | - |
| `ldap_sudo_use_host_filter` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `true` | - |
| `ldap_sudorule_command` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_host` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_name` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_notafter` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_notbefore` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_object_class` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_object_class_attr` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_option` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_order` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_runas` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_runasgroup` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_runasuser` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_sudorule_user` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - |
| `ldap_tls_cacert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cacertdir` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_cert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - |
| `ldap_tls_cipher_suite` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - |
| `ldap_tls_key` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - |
| `ldap_tls_reqcert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `hard` | never, allow, try, demand, hard |
| `ldap_uri` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - |
| `ldap_use_ppolicy` | `bool` | `domain/ldap` | `true` | - |
| `ldap_use_tokengroups` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `True for AD and IPA otherwise False.` | - |
| `ldap_user_ad_account_expires` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_ad_user_account_control` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_auth_type` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_authorized_host` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_authorized_rhost` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_authorized_service` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_certificate` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_email` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_extra_attrs` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_fullname` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_gecos` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_gid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_home_directory` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_krb_last_pwd_change` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_krb_password_expiration` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_member_of` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_modify_timestamp` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_nds_login_allowed_time_map` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_nds_login_disabled` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_nds_login_expiration_time` | `string` | `domain/ldap/id` | `-` | - |
| `ldap_user_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_objectsid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_passkey` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_primary_group` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_principal` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - |
| `ldap_user_search_filter` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_search_scope` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_expire` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_flag` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_inactive` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_last_change` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_max` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_min` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shadow_warning` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_shell` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_ssh_public_key` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_uid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `ldap_user_uuid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - |
| `local_auth_policy` | `string` | `domain` | `match` | match, only, enable, disable |
| `lookup_family_order` | `string` | `domain` | `ipv4_first` | - |
| `maprule` | `string` | `certmap` | `-` | - |
| `matchrule` | `string` | `certmap` | `KRB5:&lt;EKU&gt;clientAuth, i.e. only` | - |
| `max_ccache_size` | `int` | `kcm` | `65536` | - |
| `max_ccaches` | `int` | `kcm` | `0 (unlimited, only the per-UID quota is enforced)` | - |
| `max_id` | `int` | `domain` | `1 for min_id, 0 (no limit) for max_id` | - |
| `max_uid_ccaches` | `int` | `kcm` | `64` | - |
| `memcache_size_group` | `int` | `nss` | `6` | - |
| `memcache_size_initgroups` | `int` | `nss` | `10` | - |
| `memcache_size_passwd` | `int` | `nss` | `8` | - |
| `memcache_size_sid` | `int` | `nss` | `6` | - |
| `memcache_timeout` | `int` | `nss` | `300` | - |
| `min_id` | `int` | `domain` | `1 for min_id, 0 (no limit) for max_id` | - |
| `monitor_resolv_conf` | `bool` | `sssd` | `true` | - |
| `offline_credentials_expiration` | `int` | `pam` | `0 (No limit)` | - |
| `offline_failed_login_attempts` | `int` | `pam` | `0 (No limit)` | - |
| `offline_failed_login_delay` | `int` | `pam` | `5` | - |
| `offline_timeout` | `int` | `domain` | `60` | - |
| `offline_timeout_max` | `int` | `domain` | `3600` | - |
| `offline_timeout_random_offset` | `int` | `domain` | `30` | - |
| `override_gid` | `int` | `domain` | `not set (Use the primary GID value retrieved` | - |
| `override_homedir` | `string` | `nss, domain` | `-` | - |
| `override_shell` | `string` | `nss, domain` | `not set (SSSD will use the value` | - |
| `override_space` | `string` | `sssd` | `not set (spaces will not be replaced)` | - |
| `p11_child_timeout` | `int` | `pam` | `10` | - |
| `p11_uri` | `string` | `pam` | `none` | - |
| `p11_wait_for_card_timeout` | `int` | `pam` | `60` | - |
| `pac_check` | `string` | `pac` | `-` | no_check, pac_present, check_upn, check_upn_allow_missing, upn_dns_info_present, check_upn_dns_info_ex, upn_dns_info_ex_present |
| `pac_lifetime` | `int` | `pac` | `300` | - |
| `pam_account_expired_message` | `string` | `pam` | `none` | - |
| `pam_account_locked_message` | `string` | `pam` | `none` | - |
| `pam_app_services` | `string` | `pam` | `Not set` | - |
| `pam_cert_auth` | `bool` | `pam` | `False` | - |
| `pam_cert_db_path` | `string` | `pam` | `-` | - |
| `pam_cert_verification` | `string` | `pam` | `not set, i.e. use default` | - |
| `pam_gssapi_check_upn` | `bool` | `pam, domain` | `True` | - |
| `pam_gssapi_indicators_apply` | `string` | `pam, domain` | `not set` | - |
| `pam_gssapi_indicators_map` | `string` | `pam, domain` | `not set (use of authentication indicators is not required)` | - |
| `pam_gssapi_services` | `string` | `pam, domain` | `- (GSSAPI authentication is disabled)` | - |
| `pam_id_timeout` | `int` | `pam` | `5` | - |
| `pam_initgroups_scheme` | `string` | `pam` | `-` | always, no_session, never |
| `pam_json_services` | `string` | `pam` | `- (JSON protocol is disabled)` | - |
| `pam_p11_allowed_services` | `string` | `pam` | `the default set of PAM service names` | - |
| `pam_passkey_auth` | `bool` | `pam` | `True` | - |
| `pam_public_domains` | `string` | `pam` | `none` | - |
| `pam_pwd_expiration_warning` | `int` | `pam` | `0` | - |
| `pam_response_filter` | `string` | `pam` | `-` | env |
| `pam_trusted_users` | `string` | `pam` | `All users are considered trusted` | - |
| `pam_verbosity` | `int` | `pam` | `1` | - |
| `passkey_child_timeout` | `int` | `pam` | `15` | - |
| `passkey_debug_libfido2` | `bool` | `pam` | `False` | - |
| `passkey_verification` | `string` | `sssd` | `-` | user_verification |
| `priority` | `int` | `certmap` | `the lowest priority` | - |
| `proxy_fast_alias` | `bool` | `domain/proxy/id, domain` | `false` | - |
| `proxy_lib_name` | `string` | `domain/proxy/id, domain` | `-` | - |
| `proxy_max_children` | `int` | `domain/proxy, domain` | `10` | - |
| `proxy_pam_target` | `string` | `domain/proxy/auth, domain` | `not set by default, you have to take an` | - |
| `proxy_resolver_lib_name` | `string` | `domain` | `-` | - |
| `pwd_expiration_warning` | `int` | `domain` | `7 (Kerberos), 0 (LDAP)` | - |
| `pwfield` | `string` | `nss` | `-` | - |
| `re_expression` | `string` | `sssd, domain` | `-` | - |
| `realmd_tags` | `string` | `domain` | `-` | - |
| `refresh_expired_interval` | `int` | `domain` | `0 (disabled)` | - |
| `refresh_expired_interval_offset` | `int` | `domain` | `-` | - |
| `resolver_provider` | `string` | `domain` | `The value of` | - |
| `responder_idle_timeout` | `int` | `service, *` | `300` | - |
| `scope` | `string` | `session_recording` | `-` | - |
| `selinux_provider` | `string` | `domain` | `the value of` | - |
| `services` | `list` | `sssd` | `nss` | - |
| `session_provider` | `string` | `domain` | `-` | - |
| `shell_fallback` | `string` | `nss` | `/bin/sh` | - |
| `simple_allow_groups` | `string` | `domain/simple/access, domain/simple` | `-` | - |
| `simple_allow_users` | `string` | `domain/simple/access, domain/simple` | `-` | - |
| `simple_deny_groups` | `string` | `domain/simple/access, domain/simple` | `-` | - |
| `simple_deny_users` | `string` | `domain/simple/access, domain/simple` | `-` | - |
| `socket_path` | `string` | `kcm` | `-` | - |
| `ssh_hash_known_hosts` | `bool` | `ssh` | `-` | - |
| `ssh_known_hosts_timeout` | `int` | `ssh` | `-` | - |
| `ssh_use_certificate_keys` | `bool` | `ssh` | `true` | - |
| `ssh_use_certificate_matching_rules` | `string` | `ssh` | `not set, equivalent to 'all_rules',` | - |
| `subdomain_homedir` | `string` | `domain` | `-` | - |
| `subdomain_inherit` | `string` | `domain` | `none` | - |
| `subdomain_refresh_interval` | `int` | `domain` | `-` | - |
| `subdomain_refresh_interval_offset` | `int` | `domain` | `-` | - |
| `subdomains_provider` | `string` | `domain` | `The value of` | - |
| `sudo_inverse_order` | `bool` | `sudo` | `-` | - |
| `sudo_provider` | `string` | `domain` | `The value of` | none, ad, ipa, ldap |
| `sudo_threshold` | `int` | `sudo` | `50` | - |
| `sudo_timed` | `bool` | `sudo` | `false` | - |
| `tgt_renewal` | `bool` | `kcm` | `False (Automatic renewals disabled)` | - |
| `tgt_renewal_inherit` | `string` | `kcm` | `NULL` | - |
| `timeout` | `int` | `service, domain, *` | `10` | - |
| `try_inotify` | `bool` | `sssd` | `true on platforms where inotify is` | - |
| `use_fully_qualified_names` | `bool` | `domain` | `FALSE (TRUE for trusted` | - |
| `user` | `string` | `sssd` | `-` | - |
| `user_attributes` | `string` | `nss, ifp` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell |
| `users` | `list` | `session_recording` | `Empty. Matches no users.` | - |
| `vetoed_shells` | `list` | `nss` | `Not set` | - |
| `wildcard_limit` | `int` | `domain/ad, domain/ipa, domain/ldap` | `1000 (often the size of one page)` | - |

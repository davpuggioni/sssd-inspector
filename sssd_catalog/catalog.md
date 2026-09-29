# SSSD Configuration Options Reference Catalog

Auto-generated from SSSD 2.14.0 on 2026-09-29.
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

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `cache_first` | `int` | `true` | - | sssd.api.conf (api) |
| `client_idle_timeout` | `int` | `60, KCM: 300` | - | sssd.api.conf (api) |
| `debug` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_backtrace_enabled` | `bool` | `true` | - | sssd.api.conf (api) |
| `debug_level` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_microseconds` | `bool` | `false` | - | sssd.api.conf (api) |
| `debug_timestamps` | `bool` | `true` | - | sssd.api.conf (api) |
| `fd_limit` | `int` | `8192 (or limits.conf "hard" limit)` | - | sssd.api.conf (api) |
| `responder_idle_timeout` | `int` | `300` | - | sssd.api.conf (api) |
| `timeout` | `int` | `10` | - | sssd.api.conf (api) |

## Section `[autofs]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `autofs_negative_timeout` | `int` | `15` | - | sssd.api.conf (api) |

## Section `[certmap]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `domains` | `list` | `the configured domain in sssd.conf` | - | sssd.api.conf (api) |
| `maprule` | `string` | `-` | - | sssd.conf.5.xml (man) |
| `matchrule` | `string` | `KRB5:&lt;EKU&gt;clientAuth, i.e. only` | - | sssd.conf.5.xml (man) |
| `priority` | `int` | `the lowest priority` | - | sssd.conf.5.xml (man) |

## Section `[domain]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `access_provider` | `string` | `-` | permit, deny, ad, ipa, ldap, simple, proxy | sssd.api.conf (api) |
| `account_cache_expiration` | `int` | `0 (unlimited)` | - | sssd.api.conf (api) |
| `auth_provider` | `string` | `-` | none, ad, ipa, ldap, krb5, proxy | sssd.api.conf (api) |
| `auto_private_groups` | `string` | `-` | true, false, hybrid | sssd.api.conf (api) |
| `autofs_provider` | `string` | `The value of` | none, ad, ipa, ldap | sssd.api.conf (api) |
| `avoid_by_id_lookups` | `bool` | `False (True for IdP provider)` | - | sssd.api.conf (api) |
| `cache_credentials` | `bool` | `FALSE` | - | sssd.api.conf (api) |
| `cache_credentials_minimal_first_factor_length` | `int` | `8` | - | sssd.api.conf (api) |
| `cached_auth_timeout` | `int` | `0` | - | sssd.api.conf (api) |
| `case_sensitive` | `string` | `-` | true, false, preserving | sssd.api.conf (api) |
| `chpass_provider` | `string` | `-` | none, ad, ipa, ldap, krb5, proxy | sssd.api.conf (api) |
| `command` | `string` | `-` | - | sssd.api.conf (api) |
| `debug` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_level` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_timestamps` | `bool` | `true` | - | sssd.api.conf (api) |
| `default_shell` | `string` | `not set (Return NULL if no shell is` | - | sssd.api.conf (api) |
| `description` | `string` | `-` | - | sssd.api.conf (api) |
| `dns_discovery_domain` | `string` | `Use the domain part of machine's hostname` | - | sssd.api.conf (api) |
| `dns_resolver_op_timeout` | `int` | `3` | - | sssd.api.conf (api) |
| `dns_resolver_server_timeout` | `int` | `1000` | - | sssd.api.conf (api) |
| `dns_resolver_timeout` | `int` | `6` | - | sssd.api.conf (api) |
| `dns_resolver_use_search_list` | `bool` | `TRUE` | - | sssd.conf.5.xml (man) |
| `domain_type` | `string` | `posix` | - | sssd.api.conf (api) |
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - | sssd.api.conf (api) |
| `dyndns_auth` | `string` | `GSS-TSIG` | - | sssd.api.conf (api) |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - | sssd.api.conf (api) |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - | sssd.api.conf (api) |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval_offset` | `int` | `-` | - | sssd.api.conf (api) |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - | sssd.api.conf (api) |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - | sssd.api.conf (api) |
| `dyndns_update` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_per_family` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_ptr` | `bool` | `True` | - | sssd.api.conf (api) |
| `enabled` | `bool` | `-` | - | sssd.api.conf (api) |
| `entry_cache_autofs_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_group_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_netgroup_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_resolver_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_service_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_ssh_host_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_sudo_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_timeout` | `int` | `5400` | - | sssd.api.conf (api) |
| `entry_cache_user_timeout` | `int` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `enumerate` | `bool` | `FALSE` | - | sssd.api.conf (api) |
| `failover_primary_timeout` | `int` | `31` | - | sssd.api.conf (api) |
| `fallback_homedir` | `string` | `not set (no substitution for unset home` | - | sssd.api.conf (api) |
| `filter_groups` | `list` | `root` | - | sssd.api.conf (api) |
| `filter_users` | `list` | `root` | - | sssd.api.conf (api) |
| `full_name_format` | `string` | `-` | - | sssd.api.conf (api) |
| `homedir_substring` | `string` | `/home` | - | sssd.api.conf (api) |
| `hostid_provider` | `string` | `The value of` | - | sssd.api.conf (api) |
| `id_provider` | `string` | `Not set (This is a mandatory setting that` | ad, ipa, ldap, proxy, simple, files | sssd.api.conf (api) |
| `idmap_range_max` | `int` | `2000200000` | - | sssd-idp.5.xml (man) |
| `idmap_range_min` | `int` | `200000` | - | sssd-idp.5.xml (man) |
| `idmap_range_size` | `int` | `200000` | - | sssd-idp.5.xml (man) |
| `idp_auth_scope` | `string` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_auto_refresh` | `bool` | `false` | - | sssd-idp.5.xml (man) |
| `idp_client_id` | `string` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_client_secret` | `string` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_device_auth_endpoint` | `string` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_id_scope` | `string` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_request_timeout` | `int` | `10` | - | sssd-idp.5.xml (man) |
| `idp_token_endpoint` | `string` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_type` | `string` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_userinfo_endpoint` | `string` | `Not set` | - | sssd-idp.5.xml (man) |
| `ignore_group_members` | `bool` | `FALSE` | - | sssd.api.conf (api) |
| `inherit_from` | `string` | `Not set` | - | sssd.conf.5.xml (man) |
| `local_auth_policy` | `string` | `match` | match, only, enable, disable | sssd.api.conf (api) |
| `lookup_family_order` | `string` | `ipv4_first` | - | sssd.api.conf (api) |
| `max_id` | `int` | `1 for min_id, 0 (no limit) for max_id` | - | sssd.api.conf (api) |
| `min_id` | `int` | `1 for min_id, 0 (no limit) for max_id` | - | sssd.api.conf (api) |
| `offline_timeout` | `int` | `60` | - | sssd.api.conf (api) |
| `offline_timeout_max` | `int` | `3600` | - | sssd.api.conf (api) |
| `offline_timeout_random_offset` | `int` | `30` | - | sssd.api.conf (api) |
| `override_gid` | `int` | `not set (Use the primary GID value retrieved` | - | sssd.api.conf (api) |
| `override_homedir` | `string` | `-` | - | sssd.api.conf (api) |
| `override_shell` | `string` | `not set (SSSD will use the value` | - | sssd.api.conf (api) |
| `pam_gssapi_check_upn` | `bool` | `True` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_apply` | `string` | `not set` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_map` | `string` | `not set (use of authentication indicators is not required)` | - | sssd.api.conf (api) |
| `pam_gssapi_services` | `string` | `- (GSSAPI authentication is disabled)` | - | sssd.api.conf (api) |
| `proxy_fast_alias` | `bool` | `false` | - | sssd-proxy.conf (api) |
| `proxy_lib_name` | `string` | `-` | - | sssd-proxy.conf (api) |
| `proxy_max_children` | `int` | `10` | - | sssd-proxy.conf (api) |
| `proxy_pam_target` | `string` | `not set by default, you have to take an` | - | sssd-proxy.conf (api) |
| `proxy_resolver_lib_name` | `string` | `-` | - | sssd.conf.5.xml (man) |
| `pwd_expiration_warning` | `int` | `7 (Kerberos), 0 (LDAP)` | - | sssd.api.conf (api) |
| `re_expression` | `string` | `-` | - | sssd.api.conf (api) |
| `realmd_tags` | `string` | `-` | - | sssd.api.conf (api) |
| `refresh_expired_interval` | `int` | `0 (disabled)` | - | sssd.api.conf (api) |
| `refresh_expired_interval_offset` | `int` | `-` | - | sssd.api.conf (api) |
| `resolver_provider` | `string` | `The value of` | - | sssd.api.conf (api) |
| `selinux_provider` | `string` | `the value of` | - | sssd.api.conf (api) |
| `session_provider` | `string` | `-` | - | sssd.api.conf (api) |
| `subdomain_homedir` | `string` | `-` | - | sssd.api.conf (api) |
| `subdomain_inherit` | `string` | `none` | - | sssd.api.conf (api) |
| `subdomain_refresh_interval` | `int` | `-` | - | sssd.api.conf (api) |
| `subdomain_refresh_interval_offset` | `int` | `-` | - | sssd.api.conf (api) |
| `subdomains_provider` | `string` | `The value of` | - | sssd.api.conf (api) |
| `sudo_provider` | `string` | `The value of` | none, ad, ipa, ldap | sssd.api.conf (api) |
| `timeout` | `int` | `10` | - | sssd.api.conf (api) |
| `use_fully_qualified_names` | `bool` | `FALSE (TRUE for trusted` | - | sssd.api.conf (api) |

## Section `[domain/ad]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ad_access_filter` | `string` | `Not set` | - | sssd-ad.conf (api) |
| `ad_backup_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `ad_domain` | `string` | `-` | - | sssd-ad.conf (api) |
| `ad_enable_dns_sites` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ad_enable_gc` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ad_enabled_domains` | `string` | `Not set` | - | sssd-ad.conf (api) |
| `ad_gpo_access_control` | `string` | `permissive` | disabled, enforcing, permissive | sssd-ad.conf (api) |
| `ad_gpo_cache_timeout` | `int` | `5 (seconds)` | - | sssd-ad.conf (api) |
| `ad_gpo_default_right` | `string` | `deny` | interactive, remote_interactive, network, batch, service, permit, deny | sssd-ad.conf (api) |
| `ad_gpo_ignore_unreadable` | `bool` | `False` | - | sssd-ad.5.xml (man) |
| `ad_gpo_implicit_deny` | `bool` | `False` | - | sssd-ad.5.xml (man) |
| `ad_gpo_map_batch` | `string` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_deny` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ad_gpo_map_interactive` | `string` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_network` | `string` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_permit` | `string` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_remote_interactive` | `string` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_service` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ad_hostname` | `string` | `-` | - | sssd-ad.conf (api) |
| `ad_machine_account_password_renewal_opts` | `string` | `86400:750:300:realm (24h, 12m30s and 5m)` | - | sssd-ad.conf (api) |
| `ad_maximum_machine_account_password_age` | `int` | `30 days` | - | sssd-ad.conf (api) |
| `ad_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `ad_site` | `string` | `Not set` | - | sssd-ad.conf (api) |
| `ad_update_samba_machine_account_password` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ad_use_ldaps` | `bool` | `False` | - | sssd-ad.conf (api) |
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - | sssd.api.conf (api) |
| `dyndns_auth` | `string` | `GSS-TSIG` | - | sssd.api.conf (api) |
| `dyndns_auth_ptr` | `string` | `Same as dyndns_auth` | - | sssd-ad.5.xml (man) |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - | sssd.api.conf (api) |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - | sssd.api.conf (api) |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - | sssd.api.conf (api) |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - | sssd.api.conf (api) |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - | sssd.api.conf (api) |
| `dyndns_update` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_per_family` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_ptr` | `bool` | `True` | - | sssd.api.conf (api) |
| `krb5_auth_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `krb5_backup_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_canonicalize` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_confd_path` | `string` | `not set (krb5.include.d subdirectory of` | - | sssd-ad.conf (api) |
| `krb5_kdcip` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_realm` | `string` | `System defaults, see` | - | sssd-ad.conf (api) |
| `krb5_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_use_kdcinfo` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_backup_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_offset` | `int` | `0` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_default_authtok` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_default_authtok_type` | `string` | `password` | - | sssd-ad.conf (api) |
| `ldap_default_bind_dn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always | sssd-ad.conf (api) |
| `ldap_deref_threshold` | `int` | `10` | - | sssd-ad.conf (api) |
| `ldap_disable_paging` | `bool` | `False` | - | sssd-ad.conf (api) |
| `ldap_dns_service_name` | `string` | `ldap` | - | sssd-ad.conf (api) |
| `ldap_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_krb5_init_creds` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - | sssd-ad.conf (api) |
| `ldap_network_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_offline_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_opt_timeout` | `int` | `8` | - | sssd-ad.conf (api) |
| `ldap_page_size` | `int` | `1000` | - | sssd-ad.conf (api) |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop | sssd-ad.conf (api) |
| `ldap_referrals` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_rootdse_last_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - | sssd-ad.conf (api) |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_mech` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad | sssd-ad.conf (api) |
| `ldap_search_base` | `string` | `If not set, the value of the` | - | sssd-ad.conf (api) |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cert` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_key` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard | sssd-ad.conf (api) |
| `ldap_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - | sssd-ad.conf (api) |

## Section `[domain/ad/auth]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_ccachedir` | `string` | `/tmp` | - | sssd-ad.conf (api) |
| `krb5_ccname_template` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `krb5_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_map_user` | `string` | `not set` | - | sssd-ad.conf (api) |
| `krb5_renew_interval` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_renewable_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_store_password_if_offline` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - | sssd-ad.conf (api) |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand | sssd-ad.conf (api) |
| `krb5_use_subdomain_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_validate` | `bool` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwd_policy` | `string` | `none` | - | sssd-ad.conf (api) |

## Section `[domain/ad/autofs]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_autofs_entry_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_value` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_search_base` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ad/chpass]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |

## Section `[domain/ad/id]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - | sssd-ad.conf (api) |
| `ldap_force_upper_case_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_group_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_external_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_nesting_level` | `int` | `2` | - | sssd-ad.conf (api) |
| `ldap_group_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_uuid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_id_mapping` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_id_use_start_tls` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_helper_table_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_max` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_min` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_netgroup_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_pwd_attribute` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - | sssd-ad.conf (api) |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_search_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_service_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_port` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_proto` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - | sssd-ad.conf (api) |
| `ldap_user_auth_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_certificate` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_email` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_extra_attrs` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_fullname` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gecos` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_home_directory` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_password_expiration` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_member_of` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_passkey` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_primary_group` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_expire` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_flag` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_inactive` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_last_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_max` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_min` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_warning` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shell` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_ssh_public_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uuid` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ad/resolver]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_iphost_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_search_base` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ad/sudo]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - | sssd-ad.conf (api) |
| `ldap_sudo_hostnames` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_regexp` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_sudo_ip` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_sudo_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudorule_command` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_host` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notafter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notbefore` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_option` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_order` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runas` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasgroup` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasuser` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_user` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ipa]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `dyndns_address` | `string` | `No filtering of IP addresses.` | - | sssd.api.conf (api) |
| `dyndns_auth` | `string` | `GSS-TSIG` | - | sssd.api.conf (api) |
| `dyndns_auth_ptr` | `string` | `Same as dyndns_auth` | - | sssd-ad.5.xml (man) |
| `dyndns_dot_cacert` | `string` | `None (use global certificate store)` | - | sssd.api.conf (api) |
| `dyndns_dot_cert` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_dot_key` | `string` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_force_tcp` | `bool` | `False (let nsupdate choose the protocol)` | - | sssd.api.conf (api) |
| `dyndns_iface` | `string` | `Use the IP addresses of the interface which` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval` | `int` | `86400 (24 hours)` | - | sssd.api.conf (api) |
| `dyndns_server` | `string` | `None (let nsupdate choose the server)` | - | sssd.api.conf (api) |
| `dyndns_ttl` | `int` | `3600 (seconds)` | - | sssd.api.conf (api) |
| `dyndns_update` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_per_family` | `bool` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_ptr` | `bool` | `True` | - | sssd.api.conf (api) |
| `ipa_access_order` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_anchor_uuid` | `string` | `ipaAnchorUUID` | - | sssd-ipa.conf (api) |
| `ipa_automount_location` | `string` | `The location named "default"` | - | sssd-ipa.conf (api) |
| `ipa_backup_server` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_request_interval` | `int` | `60 (minutes)` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_search_base` | `string` | `Use base DN` | - | sssd-ipa.conf (api) |
| `ipa_domain` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_group_override_object_class` | `string` | `ipaGroupOverride` | - | sssd-ipa.conf (api) |
| `ipa_hbac_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_hbac_search_base` | `string` | `Use base DN` | - | sssd-ipa.conf (api) |
| `ipa_host_search_base` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostname` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_master_domain_search_base` | `string` | `the value of` | - | sssd-ipa.conf (api) |
| `ipa_override_object_class` | `string` | `ipaOverrideAnchor` | - | sssd-ipa.conf (api) |
| `ipa_ranges_search_base` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_selinux_search_base` | `string` | `the value of` | - | sssd-ipa.5.xml (man) |
| `ipa_server` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_server_mode` | `bool` | `false` | - | sssd-ipa.conf (api) |
| `ipa_subdomains_search_base` | `string` | `the value of` | - | sssd-ipa.conf (api) |
| `ipa_subid_ranges_search_base` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_user_override_object_class` | `string` | `ipaUserOverride` | - | sssd-ipa.conf (api) |
| `ipa_view_class` | `string` | `nsContainer` | - | sssd-ipa.conf (api) |
| `ipa_view_name` | `string` | `cn` | - | sssd-ipa.conf (api) |
| `ipa_views_search_base` | `string` | `the value of` | - | sssd-ipa.conf (api) |
| `krb5_auth_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_backup_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_canonicalize` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_confd_path` | `string` | `not set (krb5.include.d subdirectory of` | - | sssd-ad.conf (api) |
| `krb5_kdcip` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_realm` | `string` | `System defaults, see` | - | sssd-ad.conf (api) |
| `krb5_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_use_kdcinfo` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_backup_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_offset` | `int` | `0` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_default_authtok` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_default_authtok_type` | `string` | `password` | - | sssd-ad.conf (api) |
| `ldap_default_bind_dn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always | sssd-ad.conf (api) |
| `ldap_deref_threshold` | `int` | `10` | - | sssd-ad.conf (api) |
| `ldap_disable_paging` | `bool` | `False` | - | sssd-ad.conf (api) |
| `ldap_dns_service_name` | `string` | `ldap` | - | sssd-ad.conf (api) |
| `ldap_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_krb5_init_creds` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - | sssd-ad.conf (api) |
| `ldap_network_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_offline_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_opt_timeout` | `int` | `8` | - | sssd-ad.conf (api) |
| `ldap_page_size` | `int` | `1000` | - | sssd-ad.conf (api) |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop | sssd-ad.conf (api) |
| `ldap_referrals` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_rootdse_last_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - | sssd-ad.conf (api) |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_mech` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad | sssd-ad.conf (api) |
| `ldap_search_base` | `string` | `If not set, the value of the` | - | sssd-ad.conf (api) |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cert` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_key` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard | sssd-ad.conf (api) |
| `ldap_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - | sssd-ad.conf (api) |

## Section `[domain/ipa/access]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_hbac_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_hbac_support_srchost` | `bool` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_fqdn` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_member_of` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_serverhostname` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_ssh_public_key` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_member` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_memberof` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_objectclass` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |

## Section `[domain/ipa/auth]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_ccachedir` | `string` | `/tmp` | - | sssd-ad.conf (api) |
| `krb5_ccname_template` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `krb5_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_map_user` | `string` | `not set` | - | sssd-ad.conf (api) |
| `krb5_renew_interval` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_renewable_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_store_password_if_offline` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - | sssd-ad.conf (api) |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand | sssd-ad.conf (api) |
| `krb5_use_subdomain_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_validate` | `bool` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwd_policy` | `string` | `none` | - | sssd-ad.conf (api) |

## Section `[domain/ipa/autofs]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_automount_location` | `string` | `The location named "default"` | - | sssd-ipa.conf (api) |
| `ldap_autofs_entry_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_value` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_search_base` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ipa/id]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_anchor_uuid` | `string` | `ipaAnchorUUID` | - | sssd-ipa.conf (api) |
| `ipa_group_override_object_class` | `string` | `ipaGroupOverride` | - | sssd-ipa.conf (api) |
| `ipa_host_fqdn` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_ssh_public_key` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_domain` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_ext_host` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_host` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_of` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_user` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_override_object_class` | `string` | `ipaOverrideAnchor` | - | sssd-ipa.conf (api) |
| `ipa_server_mode` | `bool` | `false` | - | sssd-ipa.conf (api) |
| `ipa_user_override_object_class` | `string` | `ipaUserOverride` | - | sssd-ipa.conf (api) |
| `ipa_view_class` | `string` | `nsContainer` | - | sssd-ipa.conf (api) |
| `ipa_view_name` | `string` | `cn` | - | sssd-ipa.conf (api) |
| `ipa_views_search_base` | `string` | `the value of` | - | sssd-ipa.conf (api) |
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - | sssd-ad.conf (api) |
| `ldap_force_upper_case_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_group_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_external_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_nesting_level` | `int` | `2` | - | sssd-ad.conf (api) |
| `ldap_group_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_uuid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_id_mapping` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_id_use_start_tls` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_helper_table_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_max` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_min` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_netgroup_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_pwd_attribute` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - | sssd-ad.conf (api) |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_search_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_service_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_port` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_proto` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - | sssd-ad.conf (api) |
| `ldap_user_auth_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_certificate` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_email` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_extra_attrs` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_fullname` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gecos` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_home_directory` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_password_expiration` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_member_of` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_passkey` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_primary_group` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_expire` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_flag` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_inactive` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_last_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_max` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_min` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_warning` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shell` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_ssh_public_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uuid` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ipa/session]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_deskprofile_refresh` | `int` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_request_interval` | `int` | `60 (minutes)` | - | sssd-ipa.conf (api) |
| `ipa_host_fqdn` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_member_of` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_serverhostname` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_ssh_public_key` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_enabled` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_host_category` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_member_host` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_member_user` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_see_also` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_selinux_user` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_user_category` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |

## Section `[domain/ipa/subdomains]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_subdomains_search_base` | `string` | `the value of` | - | sssd-ipa.conf (api) |

## Section `[domain/ipa/sudo]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ipa_sudocmd_memberof` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_sudocmd` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_entry_usn` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_member` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_allowcmd` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_cmdcategory` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_denycmd` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_enabled_flag` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_entry_usn` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_externaluser` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_host` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_hostcategory` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_name` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_notafter` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_notbefore` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_object_class` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_option` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextgroup` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextuser` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextusergroup` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasgroup` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasgroupcategory` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasusercategory` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_sudoorder` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_user` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_usercategory` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_uuid` | `string` | `-` | - | sssd-ipa.conf (api) |
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - | sssd-ad.conf (api) |
| `ldap_sudo_hostnames` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_regexp` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_sudo_ip` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_sudo_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudorule_command` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_host` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notafter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notbefore` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_option` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_order` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runas` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasgroup` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasuser` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_user` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/krb5]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_auth_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `krb5_backup_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_backup_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_ccachedir` | `string` | `/tmp` | - | sssd-ad.conf (api) |
| `krb5_ccname_template` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_kdcinfo_lookahead` | `string` | `3:1` | - | sssd-krb5.5.xml (man) |
| `krb5_kdcip` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `krb5_kpasswd` | `string` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_map_user` | `string` | `not set` | - | sssd-ad.conf (api) |
| `krb5_realm` | `string` | `System defaults, see` | - | sssd-ad.conf (api) |
| `krb5_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_store_password_if_offline` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - | sssd-ad.conf (api) |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand | sssd-ad.conf (api) |
| `krb5_use_kdcinfo` | `bool` | `true` | - | sssd-ad.conf (api) |
| `krb5_use_subdomain_realm` | `bool` | `false` | - | sssd-ad.conf (api) |

## Section `[domain/krb5/auth]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_canonicalize` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_ccachedir` | `string` | `/tmp` | - | sssd-ad.conf (api) |
| `krb5_ccname_template` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `krb5_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_map_user` | `string` | `not set` | - | sssd-ad.conf (api) |
| `krb5_renew_interval` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_renewable_lifetime` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_store_password_if_offline` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_use_enterprise_principal` | `bool` | `false (AD provider: true)` | - | sssd-ad.conf (api) |
| `krb5_use_fast` | `string` | `not set, i.e. FAST is not used.` | never, try, demand | sssd-ad.conf (api) |
| `krb5_use_subdomain_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_validate` | `bool` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ldap]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `krb5_backup_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_canonicalize` | `bool` | `false` | - | sssd-ad.conf (api) |
| `krb5_kdcip` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_realm` | `string` | `System defaults, see` | - | sssd-ad.conf (api) |
| `krb5_server` | `string` | `-` | - | sssd-ad.conf (api) |
| `krb5_use_kdcinfo` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_access_filter` | `string` | `Empty` | - | sssd-ldap.conf (api) |
| `ldap_access_order` | `string` | `filter` | - | sssd-ldap.conf (api) |
| `ldap_account_expire_policy` | `string` | `Empty` | - | sssd-ldap.conf (api) |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - | sssd-ad.conf (api) |
| `ldap_autofs_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_backup_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_chpass_backup_uri` | `string` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |
| `ldap_chpass_dns_service_name` | `string` | `not set, i.e. service discovery is disabled` | - | sssd-ldap.conf (api) |
| `ldap_chpass_update_last_change` | `bool` | `False` | - | sssd-ldap.conf (api) |
| `ldap_chpass_uri` | `string` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |
| `ldap_connection_expire_offset` | `int` | `0` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_connection_idle_timeout` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_default_authtok` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_default_authtok_type` | `string` | `password` | - | sssd-ad.conf (api) |
| `ldap_default_bind_dn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_deref` | `string` | `Empty (this is handled as` | never, searching, finding, always | sssd-ad.conf (api) |
| `ldap_deref_threshold` | `int` | `10` | - | sssd-ad.conf (api) |
| `ldap_disable_paging` | `bool` | `False` | - | sssd-ad.conf (api) |
| `ldap_disable_range_retrieval` | `bool` | `False` | - | sssd-ldap.conf (api) |
| `ldap_dns_service_name` | `string` | `ldap` | - | sssd-ad.conf (api) |
| `ldap_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - | sssd-ad.conf (api) |
| `ldap_enumeration_search_timeout` | `int` | `60` | - | sssd-ldap.conf (api) |
| `ldap_force_upper_case_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_group_nesting_level` | `int` | `2` | - | sssd-ad.conf (api) |
| `ldap_group_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_host_search_base` | `string` | `the value of` | - | sssd-ldap.5.xml (man) |
| `ldap_id_mapping` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_id_use_start_tls` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_ignore_unreadable_references` | `bool` | `False` | - | sssd-ldap.conf (api) |
| `ldap_iphost_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_krb5_init_creds` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_krb5_keytab` | `string` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `ldap_krb5_ticket_lifetime` | `int` | `86400 (24 hours)` | - | sssd-ad.conf (api) |
| `ldap_library_debug_level` | `int` | `0 (libldap debugging disabled)` | - | sssd-ldap.conf (api) |
| `ldap_max_id` | `int` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_min_id` | `int` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_network_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_offline_timeout` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_opt_timeout` | `int` | `8` | - | sssd-ad.conf (api) |
| `ldap_page_size` | `int` | `1000` | - | sssd-ad.conf (api) |
| `ldap_ppolicy_pwd_change_threshold` | `int` | `0` | - | sssd-ldap.conf (api) |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_pwd_policy` | `string` | `none` | - | sssd-ad.conf (api) |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - | sssd-ad.conf (api) |
| `ldap_pwmodify_mode` | `string` | `exop` | ldap_modify, exop | sssd-ad.conf (api) |
| `ldap_read_rootdse` | `string` | `anonymous` | - | sssd-ldap.5.xml (man) |
| `ldap_referrals` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_rootdse_last_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sasl_authid` | `string` | `host/hostname@REALM` | - | sssd-ad.conf (api) |
| `ldap_sasl_canonicalize` | `bool` | `false;` | - | sssd-ldap.conf (api) |
| `ldap_sasl_maxssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_mech` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_sasl_minssf` | `int` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_realm` | `string` | `the value of krb5_realm.` | - | sssd-ldap.5.xml (man) |
| `ldap_schema` | `string` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad | sssd-ad.conf (api) |
| `ldap_search_base` | `string` | `If not set, the value of the` | - | sssd-ad.conf (api) |
| `ldap_search_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_service_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_subid_ranges_search_base` | `string` | `the value of` | - | sssd-ldap.conf (api) |
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - | sssd-ad.conf (api) |
| `ldap_sudo_hostnames` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_regexp` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_sudo_ip` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_sudo_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_tls_cacert` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cacertdir` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cert` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_cipher_suite` | `string` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_key` | `string` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_reqcert` | `string` | `hard` | never, allow, try, demand, hard | sssd-ad.conf (api) |
| `ldap_uri` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_use_ppolicy` | `bool` | `true` | - | sssd-ldap.conf (api) |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - | sssd-ad.conf (api) |
| `ldap_user_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `wildcard_limit` | `int` | `1000 (often the size of one page)` | - | sssd-ad.conf (api) |

## Section `[domain/ldap/access]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_access_filter` | `string` | `Empty` | - | sssd-ldap.conf (api) |
| `ldap_access_order` | `string` | `filter` | - | sssd-ldap.conf (api) |
| `ldap_account_expire_policy` | `string` | `Empty` | - | sssd-ldap.conf (api) |

## Section `[domain/ldap/auth]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_pwd_policy` | `string` | `none` | - | sssd-ad.conf (api) |

## Section `[domain/ldap/autofs]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_autofs_entry_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_value` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_master_name` | `string` | `auto.master` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_search_base` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ldap/chpass]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_chpass_backup_uri` | `string` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |
| `ldap_chpass_dns_service_name` | `string` | `not set, i.e. service discovery is disabled` | - | sssd-ldap.conf (api) |
| `ldap_chpass_update_last_change` | `bool` | `False` | - | sssd-ldap.conf (api) |
| `ldap_chpass_uri` | `string` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |

## Section `[domain/ldap/id]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_enumeration_refresh_timeout` | `int` | `300` | - | sssd-ad.conf (api) |
| `ldap_enumeration_search_timeout` | `int` | `60` | - | sssd-ldap.conf (api) |
| `ldap_force_upper_case_realm` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_group_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_external_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_member` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_nesting_level` | `int` | `2` | - | sssd-ad.conf (api) |
| `ldap_group_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_uuid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_id_mapping` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_id_use_start_tls` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_idmap_autorid_compat` | `bool` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain_sid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_helper_table_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_max` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_min` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_size` | `int` | `-` | - | sssd-ad.conf (api) |
| `ldap_library_debug_level` | `int` | `0 (libldap debugging disabled)` | - | sssd-ldap.conf (api) |
| `ldap_max_id` | `int` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_min_id` | `int` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_member` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_modify_timestamp` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_name` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_object_class` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_netgroup_triple` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_ns_account_lock` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_purge_cache_timeout` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_pwd_attribute` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwdlockout_dn` | `string` | `cn=ppolicy,ou=policies,$ldap_search_base` | - | sssd-ad.conf (api) |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_search_timeout` | `int` | `6` | - | sssd-ad.conf (api) |
| `ldap_service_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_port` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_proto` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_subid_ranges_search_base` | `string` | `the value of` | - | sssd-ldap.conf (api) |
| `ldap_use_tokengroups` | `bool` | `True for AD and IPA otherwise False.` | - | sssd-ad.conf (api) |
| `ldap_user_ad_account_expires` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_ad_user_account_control` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_auth_type` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_authorized_host` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_authorized_rhost` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_authorized_service` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_certificate` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_email` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_extra_attrs` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_fullname` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gecos` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_home_directory` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_last_pwd_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_password_expiration` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_member_of` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_modify_timestamp` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_nds_login_allowed_time_map` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_nds_login_disabled` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_nds_login_expiration_time` | `string` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_objectsid` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_passkey` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_primary_group` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_principal` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_filter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_scope` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_expire` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_flag` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_inactive` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_last_change` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_max` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_min` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_warning` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shell` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_ssh_public_key` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uid_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uuid` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ldap/resolver]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_iphost_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_entry_usn` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_number` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_search_base` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/ldap/sudo]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ldap_sudo_full_refresh_interval` | `int` | `21600 (6 hours)` | - | sssd-ad.conf (api) |
| `ldap_sudo_hostnames` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_netgroups` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_regexp` | `bool` | `false` | - | sssd-ad.conf (api) |
| `ldap_sudo_ip` | `string` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_random_offset` | `int` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_sudo_search_base` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudo_smart_refresh_interval` | `int` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_sudo_use_host_filter` | `bool` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudorule_command` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_host` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_name` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notafter` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notbefore` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class_attr` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_option` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_order` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runas` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasgroup` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasuser` | `string` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_user` | `string` | `-` | - | sssd-ad.conf (api) |

## Section `[domain/proxy]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `proxy_max_children` | `int` | `10` | - | sssd-proxy.conf (api) |

## Section `[domain/proxy/auth]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `proxy_pam_target` | `string` | `not set by default, you have to take an` | - | sssd-proxy.conf (api) |

## Section `[domain/proxy/id]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `proxy_fast_alias` | `bool` | `false` | - | sssd-proxy.conf (api) |
| `proxy_lib_name` | `string` | `-` | - | sssd-proxy.conf (api) |

## Section `[domain/simple]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `simple_allow_groups` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_allow_users` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_groups` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_users` | `string` | `-` | - | sssd-simple.conf (api) |

## Section `[domain/simple/access]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `simple_allow_groups` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_allow_users` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_groups` | `string` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_users` | `string` | `-` | - | sssd-simple.conf (api) |

## Section `[ifp]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `allowed_uids` | `string` | `0, &sssd_user_name; (only root and SSSD` | - | sssd.api.conf (api) |
| `user_attributes` | `string` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell | sssd.api.conf (api) |

## Section `[kcm]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `max_ccache_size` | `int` | `65536` | - | sssd-kcm.8.xml (man) |
| `max_ccaches` | `int` | `0 (unlimited, only the per-UID quota is enforced)` | - | sssd-kcm.8.xml (man) |
| `max_uid_ccaches` | `int` | `64` | - | sssd-kcm.8.xml (man) |
| `socket_path` | `string` | `-` | - | sssd-kcm.8.xml (man) |
| `tgt_renewal` | `bool` | `False (Automatic renewals disabled)` | - | sssd-kcm.8.xml (man) |
| `tgt_renewal_inherit` | `string` | `NULL` | - | sssd-kcm.8.xml (man) |

## Section `[nss]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `allowed_shells` | `list` | `Not set. The user shell is automatically used.` | - | sssd.api.conf (api) |
| `default_shell` | `string` | `not set (Return NULL if no shell is` | - | sssd.api.conf (api) |
| `entry_cache_nowait_percentage` | `int` | `50` | - | sssd.api.conf (api) |
| `entry_negative_timeout` | `int` | `15` | - | sssd.api.conf (api) |
| `enum_cache_timeout` | `int` | `120` | - | sssd.api.conf (api) |
| `fallback_homedir` | `string` | `not set (no substitution for unset home` | - | sssd.api.conf (api) |
| `filter_groups` | `list` | `root` | - | sssd.api.conf (api) |
| `filter_users` | `list` | `root` | - | sssd.api.conf (api) |
| `filter_users_in_groups` | `bool` | `true` | - | sssd.api.conf (api) |
| `get_domains_timeout` | `int` | `60` | - | sssd.api.conf (api) |
| `homedir_substring` | `string` | `/home` | - | sssd.api.conf (api) |
| `memcache_size_group` | `int` | `6` | - | sssd.conf.5.xml (man) |
| `memcache_size_initgroups` | `int` | `10` | - | sssd.conf.5.xml (man) |
| `memcache_size_passwd` | `int` | `8` | - | sssd.conf.5.xml (man) |
| `memcache_size_sid` | `int` | `6` | - | sssd.conf.5.xml (man) |
| `memcache_timeout` | `int` | `300` | - | sssd.api.conf (api) |
| `override_homedir` | `string` | `-` | - | sssd.api.conf (api) |
| `override_shell` | `string` | `not set (SSSD will use the value` | - | sssd.api.conf (api) |
| `pwfield` | `string` | `-` | - | sssd.api.conf (api) |
| `shell_fallback` | `string` | `/bin/sh` | - | sssd.api.conf (api) |
| `user_attributes` | `string` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell | sssd.api.conf (api) |
| `vetoed_shells` | `list` | `Not set` | - | sssd.api.conf (api) |

## Section `[pac]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `allowed_uids` | `string` | `0, &sssd_user_name; (only root and SSSD` | - | sssd.api.conf (api) |
| `pac_check` | `string` | `-` | no_check, pac_present, check_upn, check_upn_allow_missing, upn_dns_info_present, check_upn_dns_info_ex, upn_dns_info_ex_present | sssd.api.conf (api) |
| `pac_lifetime` | `int` | `300` | - | sssd.api.conf (api) |

## Section `[pam]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `get_domains_timeout` | `int` | `60` | - | sssd.api.conf (api) |
| `offline_credentials_expiration` | `int` | `0 (No limit)` | - | sssd.api.conf (api) |
| `offline_failed_login_attempts` | `int` | `0 (No limit)` | - | sssd.api.conf (api) |
| `offline_failed_login_delay` | `int` | `5` | - | sssd.api.conf (api) |
| `p11_child_timeout` | `int` | `10` | - | sssd.api.conf (api) |
| `p11_uri` | `string` | `none` | - | sssd.api.conf (api) |
| `p11_wait_for_card_timeout` | `int` | `60` | - | sssd.api.conf (api) |
| `pam_account_expired_message` | `string` | `none` | - | sssd.api.conf (api) |
| `pam_account_locked_message` | `string` | `none` | - | sssd.api.conf (api) |
| `pam_app_services` | `string` | `Not set` | - | sssd.api.conf (api) |
| `pam_cert_auth` | `bool` | `False` | - | sssd.api.conf (api) |
| `pam_cert_db_path` | `string` | `-` | - | sssd.api.conf (api) |
| `pam_cert_verification` | `string` | `not set, i.e. use default` | - | sssd.api.conf (api) |
| `pam_gssapi_check_upn` | `bool` | `True` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_apply` | `string` | `not set` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_map` | `string` | `not set (use of authentication indicators is not required)` | - | sssd.api.conf (api) |
| `pam_gssapi_services` | `string` | `- (GSSAPI authentication is disabled)` | - | sssd.api.conf (api) |
| `pam_id_timeout` | `int` | `5` | - | sssd.api.conf (api) |
| `pam_initgroups_scheme` | `string` | `-` | always, no_session, never | sssd.api.conf (api) |
| `pam_json_services` | `string` | `- (JSON protocol is disabled)` | - | sssd.api.conf (api) |
| `pam_p11_allowed_services` | `string` | `the default set of PAM service names` | - | sssd.api.conf (api) |
| `pam_passkey_auth` | `bool` | `True` | - | sssd.api.conf (api) |
| `pam_public_domains` | `string` | `none` | - | sssd.api.conf (api) |
| `pam_pwd_expiration_warning` | `int` | `0` | - | sssd.api.conf (api) |
| `pam_response_filter` | `string` | `-` | env | sssd.api.conf (api) |
| `pam_trusted_users` | `string` | `All users are considered trusted` | - | sssd.api.conf (api) |
| `pam_verbosity` | `int` | `1` | - | sssd.api.conf (api) |
| `passkey_child_timeout` | `int` | `15` | - | sssd.api.conf (api) |
| `passkey_debug_libfido2` | `bool` | `False` | - | sssd.api.conf (api) |

## Section `[service]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `cache_first` | `int` | `true` | - | sssd.api.conf (api) |
| `client_idle_timeout` | `int` | `60, KCM: 300` | - | sssd.api.conf (api) |
| `command` | `string` | `-` | - | sssd.api.conf (api) |
| `debug` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_backtrace_enabled` | `bool` | `true` | - | sssd.api.conf (api) |
| `debug_level` | `int` | `-` | - | sssd.api.conf (api) |
| `debug_microseconds` | `bool` | `false` | - | sssd.api.conf (api) |
| `debug_timestamps` | `bool` | `true` | - | sssd.api.conf (api) |
| `description` | `string` | `-` | - | sssd.api.conf (api) |
| `fd_limit` | `int` | `8192 (or limits.conf "hard" limit)` | - | sssd.api.conf (api) |
| `responder_idle_timeout` | `int` | `300` | - | sssd.api.conf (api) |
| `timeout` | `int` | `10` | - | sssd.api.conf (api) |

## Section `[session_recording]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `exclude_groups` | `list` | `Empty. No groups excluded.` | - | sssd.api.conf (api) |
| `exclude_users` | `list` | `Empty. No users excluded.` | - | sssd.api.conf (api) |
| `groups` | `list` | `Empty. Matches no groups.` | - | sssd.api.conf (api) |
| `scope` | `string` | `-` | - | sssd.api.conf (api) |
| `users` | `list` | `Empty. Matches no users.` | - | sssd.api.conf (api) |

## Section `[ssh]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `ca_db` | `string` | `-` | - | sssd.api.conf (api) |
| `ssh_hash_known_hosts` | `bool` | `-` | - | sssd.api.conf (api) |
| `ssh_known_hosts_timeout` | `int` | `-` | - | sssd.api.conf (api) |
| `ssh_use_certificate_keys` | `bool` | `true` | - | sssd.api.conf (api) |
| `ssh_use_certificate_matching_rules` | `string` | `not set, equivalent to 'all_rules',` | - | sssd.api.conf (api) |

## Section `[sssd]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `certificate_verification` | `string` | `-` | no_verification, ocsp_dgst, ocsp, crl | sssd.api.conf (api) |
| `config_file_version` | `int` | `-` | - | curated (curated) |
| `core_dumpable` | `bool` | `true` | - | sssd.api.conf (api) |
| `default_domain_suffix` | `string` | `not set` | - | sssd.api.conf (api) |
| `disable_netlink` | `bool` | `false (netlink changes are detected)` | - | sssd.api.conf (api) |
| `domain_resolution_order` | `list` | `Not set` | - | sssd.api.conf (api) |
| `domains` | `list` | `the configured domain in sssd.conf` | - | sssd.api.conf (api) |
| `enable_files_domain` | `string` | `-` | - | sssd.api.conf (api) |
| `full_name_format` | `string` | `-` | - | sssd.api.conf (api) |
| `implicit_pac_responder` | `bool` | `true` | - | sssd.api.conf (api) |
| `krb5_rcache_dir` | `string` | `Distribution-specific and specified` | - | sssd.api.conf (api) |
| `monitor_resolv_conf` | `bool` | `true` | - | sssd.api.conf (api) |
| `override_space` | `string` | `not set (spaces will not be replaced)` | - | sssd.api.conf (api) |
| `passkey_verification` | `string` | `-` | user_verification | sssd.api.conf (api) |
| `re_expression` | `string` | `-` | - | sssd.api.conf (api) |
| `services` | `list` | `nss` | - | sssd.api.conf (api) |
| `try_inotify` | `bool` | `true on platforms where inotify is` | - | sssd.api.conf (api) |
| `user` | `string` | `-` | - | sssd.api.conf (api) |

## Section `[sudo]`

| Option | Type | Default | Allowed Values | Source |
|--------|------|---------|----------------|--------|
| `sudo_inverse_order` | `bool` | `-` | - | sssd.api.conf (api) |
| `sudo_threshold` | `int` | `50` | - | sssd.api.conf (api) |
| `sudo_timed` | `bool` | `false` | - | sssd.api.conf (api) |

## All Options (Alphabetical)

| Option | Type | Sections | Default | Allowed Values | Source |
|--------|------|----------|---------|----------------|--------|
| `access_provider` | `string` | `domain` | `-` | permit, deny, ad, ipa, ldap, simple, proxy | sssd.api.conf (api) |
| `account_cache_expiration` | `int` | `domain` | `0 (unlimited)` | - | sssd.api.conf (api) |
| `ad_access_filter` | `string` | `domain/ad` | `Not set` | - | sssd-ad.conf (api) |
| `ad_backup_server` | `string` | `domain/ad` | `-` | - | sssd-ad.conf (api) |
| `ad_domain` | `string` | `domain/ad` | `-` | - | sssd-ad.conf (api) |
| `ad_enable_dns_sites` | `bool` | `domain/ad` | `true` | - | sssd-ad.conf (api) |
| `ad_enable_gc` | `bool` | `domain/ad` | `true` | - | sssd-ad.conf (api) |
| `ad_enabled_domains` | `string` | `domain/ad` | `Not set` | - | sssd-ad.conf (api) |
| `ad_gpo_access_control` | `string` | `domain/ad` | `permissive` | disabled, enforcing, permissive | sssd-ad.conf (api) |
| `ad_gpo_cache_timeout` | `int` | `domain/ad` | `5 (seconds)` | - | sssd-ad.conf (api) |
| `ad_gpo_default_right` | `string` | `domain/ad` | `deny` | interactive, remote_interactive, network, batch, service, permit, deny | sssd-ad.conf (api) |
| `ad_gpo_ignore_unreadable` | `bool` | `domain/ad` | `False` | - | sssd-ad.5.xml (man) |
| `ad_gpo_implicit_deny` | `bool` | `domain/ad` | `False` | - | sssd-ad.5.xml (man) |
| `ad_gpo_map_batch` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_deny` | `string` | `domain/ad` | `not set` | - | sssd-ad.conf (api) |
| `ad_gpo_map_interactive` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_network` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_permit` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_remote_interactive` | `string` | `domain/ad` | `the default set of PAM service names includes:` | - | sssd-ad.conf (api) |
| `ad_gpo_map_service` | `string` | `domain/ad` | `not set` | - | sssd-ad.conf (api) |
| `ad_hostname` | `string` | `domain/ad` | `-` | - | sssd-ad.conf (api) |
| `ad_machine_account_password_renewal_opts` | `string` | `domain/ad` | `86400:750:300:realm (24h, 12m30s and 5m)` | - | sssd-ad.conf (api) |
| `ad_maximum_machine_account_password_age` | `int` | `domain/ad` | `30 days` | - | sssd-ad.conf (api) |
| `ad_server` | `string` | `domain/ad` | `-` | - | sssd-ad.conf (api) |
| `ad_site` | `string` | `domain/ad` | `Not set` | - | sssd-ad.conf (api) |
| `ad_update_samba_machine_account_password` | `bool` | `domain/ad` | `false` | - | sssd-ad.conf (api) |
| `ad_use_ldaps` | `bool` | `domain/ad` | `False` | - | sssd-ad.conf (api) |
| `allowed_shells` | `list` | `nss` | `Not set. The user shell is automatically used.` | - | sssd.api.conf (api) |
| `allowed_uids` | `string` | `pac, ifp` | `0, &sssd_user_name; (only root and SSSD` | - | sssd.api.conf (api) |
| `auth_provider` | `string` | `domain` | `-` | none, ad, ipa, ldap, krb5, proxy | sssd.api.conf (api) |
| `auto_private_groups` | `string` | `domain` | `-` | true, false, hybrid | sssd.api.conf (api) |
| `autofs_negative_timeout` | `int` | `autofs` | `15` | - | sssd.api.conf (api) |
| `autofs_provider` | `string` | `domain` | `The value of` | none, ad, ipa, ldap | sssd.api.conf (api) |
| `avoid_by_id_lookups` | `bool` | `domain` | `False (True for IdP provider)` | - | sssd.api.conf (api) |
| `ca_db` | `string` | `ssh` | `-` | - | sssd.api.conf (api) |
| `cache_credentials` | `bool` | `domain` | `FALSE` | - | sssd.api.conf (api) |
| `cache_credentials_minimal_first_factor_length` | `int` | `domain` | `8` | - | sssd.api.conf (api) |
| `cache_first` | `int` | `service, *` | `true` | - | sssd.api.conf (api) |
| `cached_auth_timeout` | `int` | `domain` | `0` | - | sssd.api.conf (api) |
| `case_sensitive` | `string` | `domain` | `-` | true, false, preserving | sssd.api.conf (api) |
| `certificate_verification` | `string` | `sssd` | `-` | no_verification, ocsp_dgst, ocsp, crl | sssd.api.conf (api) |
| `chpass_provider` | `string` | `domain` | `-` | none, ad, ipa, ldap, krb5, proxy | sssd.api.conf (api) |
| `client_idle_timeout` | `int` | `service, *` | `60, KCM: 300` | - | sssd.api.conf (api) |
| `command` | `string` | `service, domain` | `-` | - | sssd.api.conf (api) |
| `config_file_version` | `int` | `sssd` | `-` | - | curated (curated) |
| `core_dumpable` | `bool` | `sssd` | `true` | - | sssd.api.conf (api) |
| `debug` | `int` | `service, domain, *` | `-` | - | sssd.api.conf (api) |
| `debug_backtrace_enabled` | `bool` | `service, *` | `true` | - | sssd.api.conf (api) |
| `debug_level` | `int` | `service, domain, *` | `-` | - | sssd.api.conf (api) |
| `debug_microseconds` | `bool` | `service, *` | `false` | - | sssd.api.conf (api) |
| `debug_timestamps` | `bool` | `service, domain, *` | `true` | - | sssd.api.conf (api) |
| `default_domain_suffix` | `string` | `sssd` | `not set` | - | sssd.api.conf (api) |
| `default_shell` | `string` | `nss, domain` | `not set (Return NULL if no shell is` | - | sssd.api.conf (api) |
| `description` | `string` | `service, domain` | `-` | - | sssd.api.conf (api) |
| `disable_netlink` | `bool` | `sssd` | `false (netlink changes are detected)` | - | sssd.api.conf (api) |
| `dns_discovery_domain` | `string` | `domain` | `Use the domain part of machine's hostname` | - | sssd.api.conf (api) |
| `dns_resolver_op_timeout` | `int` | `domain` | `3` | - | sssd.api.conf (api) |
| `dns_resolver_server_timeout` | `int` | `domain` | `1000` | - | sssd.api.conf (api) |
| `dns_resolver_timeout` | `int` | `domain` | `6` | - | sssd.api.conf (api) |
| `dns_resolver_use_search_list` | `bool` | `domain` | `TRUE` | - | sssd.conf.5.xml (man) |
| `domain_resolution_order` | `list` | `sssd` | `Not set` | - | sssd.api.conf (api) |
| `domain_type` | `string` | `domain` | `posix` | - | sssd.api.conf (api) |
| `domains` | `list` | `sssd, certmap` | `the configured domain in sssd.conf` | - | sssd.api.conf (api) |
| `dyndns_address` | `string` | `domain, domain/ad, domain/ipa` | `No filtering of IP addresses.` | - | sssd.api.conf (api) |
| `dyndns_auth` | `string` | `domain, domain/ad, domain/ipa` | `GSS-TSIG` | - | sssd.api.conf (api) |
| `dyndns_auth_ptr` | `string` | `domain/ad, domain/ipa` | `Same as dyndns_auth` | - | sssd-ad.5.xml (man) |
| `dyndns_dot_cacert` | `string` | `domain, domain/ad, domain/ipa` | `None (use global certificate store)` | - | sssd.api.conf (api) |
| `dyndns_dot_cert` | `string` | `domain, domain/ad, domain/ipa` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_dot_key` | `string` | `domain, domain/ad, domain/ipa` | `None (Do not use TLS authentication)` | - | sssd.api.conf (api) |
| `dyndns_force_tcp` | `bool` | `domain, domain/ad, domain/ipa` | `False (let nsupdate choose the protocol)` | - | sssd.api.conf (api) |
| `dyndns_iface` | `string` | `domain, domain/ad, domain/ipa` | `Use the IP addresses of the interface which` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval` | `int` | `domain, domain/ad, domain/ipa` | `86400 (24 hours)` | - | sssd.api.conf (api) |
| `dyndns_refresh_interval_offset` | `int` | `domain` | `-` | - | sssd.api.conf (api) |
| `dyndns_server` | `string` | `domain, domain/ad, domain/ipa` | `None (let nsupdate choose the server)` | - | sssd.api.conf (api) |
| `dyndns_ttl` | `int` | `domain, domain/ad, domain/ipa` | `3600 (seconds)` | - | sssd.api.conf (api) |
| `dyndns_update` | `bool` | `domain, domain/ad, domain/ipa` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_per_family` | `bool` | `domain, domain/ad, domain/ipa` | `true` | - | sssd.api.conf (api) |
| `dyndns_update_ptr` | `bool` | `domain, domain/ad, domain/ipa` | `True` | - | sssd.api.conf (api) |
| `enable_files_domain` | `string` | `sssd` | `-` | - | sssd.api.conf (api) |
| `enabled` | `bool` | `domain` | `-` | - | sssd.api.conf (api) |
| `entry_cache_autofs_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_group_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_netgroup_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_nowait_percentage` | `int` | `nss` | `50` | - | sssd.api.conf (api) |
| `entry_cache_resolver_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_service_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_ssh_host_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_sudo_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_cache_timeout` | `int` | `domain` | `5400` | - | sssd.api.conf (api) |
| `entry_cache_user_timeout` | `int` | `domain` | `entry_cache_timeout` | - | sssd.api.conf (api) |
| `entry_negative_timeout` | `int` | `nss` | `15` | - | sssd.api.conf (api) |
| `enum_cache_timeout` | `int` | `nss` | `120` | - | sssd.api.conf (api) |
| `enumerate` | `bool` | `domain` | `FALSE` | - | sssd.api.conf (api) |
| `exclude_groups` | `list` | `session_recording` | `Empty. No groups excluded.` | - | sssd.api.conf (api) |
| `exclude_users` | `list` | `session_recording` | `Empty. No users excluded.` | - | sssd.api.conf (api) |
| `failover_primary_timeout` | `int` | `domain` | `31` | - | sssd.api.conf (api) |
| `fallback_homedir` | `string` | `nss, domain` | `not set (no substitution for unset home` | - | sssd.api.conf (api) |
| `fd_limit` | `int` | `service, *` | `8192 (or limits.conf "hard" limit)` | - | sssd.api.conf (api) |
| `filter_groups` | `list` | `nss, domain` | `root` | - | sssd.api.conf (api) |
| `filter_users` | `list` | `nss, domain` | `root` | - | sssd.api.conf (api) |
| `filter_users_in_groups` | `bool` | `nss` | `true` | - | sssd.api.conf (api) |
| `full_name_format` | `string` | `sssd, domain` | `-` | - | sssd.api.conf (api) |
| `get_domains_timeout` | `int` | `nss, pam` | `60` | - | sssd.api.conf (api) |
| `groups` | `list` | `session_recording` | `Empty. Matches no groups.` | - | sssd.api.conf (api) |
| `homedir_substring` | `string` | `nss, domain` | `/home` | - | sssd.api.conf (api) |
| `hostid_provider` | `string` | `domain` | `The value of` | - | sssd.api.conf (api) |
| `id_provider` | `string` | `domain` | `Not set (This is a mandatory setting that` | ad, ipa, ldap, proxy, simple, files | sssd.api.conf (api) |
| `idmap_range_max` | `int` | `domain` | `2000200000` | - | sssd-idp.5.xml (man) |
| `idmap_range_min` | `int` | `domain` | `200000` | - | sssd-idp.5.xml (man) |
| `idmap_range_size` | `int` | `domain` | `200000` | - | sssd-idp.5.xml (man) |
| `idp_auth_scope` | `string` | `domain` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_auto_refresh` | `bool` | `domain` | `false` | - | sssd-idp.5.xml (man) |
| `idp_client_id` | `string` | `domain` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_client_secret` | `string` | `domain` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_device_auth_endpoint` | `string` | `domain` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_id_scope` | `string` | `domain` | `Not set` | - | sssd-idp.5.xml (man) |
| `idp_request_timeout` | `int` | `domain` | `10` | - | sssd-idp.5.xml (man) |
| `idp_token_endpoint` | `string` | `domain` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_type` | `string` | `domain` | `Not set (Required)` | - | sssd-idp.5.xml (man) |
| `idp_userinfo_endpoint` | `string` | `domain` | `Not set` | - | sssd-idp.5.xml (man) |
| `ignore_group_members` | `bool` | `domain` | `FALSE` | - | sssd.api.conf (api) |
| `implicit_pac_responder` | `bool` | `sssd` | `true` | - | sssd.api.conf (api) |
| `inherit_from` | `string` | `domain` | `Not set` | - | sssd.conf.5.xml (man) |
| `ipa_access_order` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_anchor_uuid` | `string` | `domain/ipa/id, domain/ipa` | `ipaAnchorUUID` | - | sssd-ipa.conf (api) |
| `ipa_automount_location` | `string` | `domain/ipa/autofs, domain/ipa` | `The location named "default"` | - | sssd-ipa.conf (api) |
| `ipa_backup_server` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_refresh` | `int` | `domain/ipa/session, domain/ipa` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_request_interval` | `int` | `domain/ipa/session, domain/ipa` | `60 (minutes)` | - | sssd-ipa.conf (api) |
| `ipa_deskprofile_search_base` | `string` | `domain/ipa` | `Use base DN` | - | sssd-ipa.conf (api) |
| `ipa_domain` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_group_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaGroupOverride` | - | sssd-ipa.conf (api) |
| `ipa_hbac_refresh` | `int` | `domain/ipa/access, domain/ipa` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_hbac_search_base` | `string` | `domain/ipa` | `Use base DN` | - | sssd-ipa.conf (api) |
| `ipa_hbac_support_srchost` | `bool` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_fqdn` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_member_of` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_name` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_object_class` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_search_base` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_serverhostname` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_ssh_public_key` | `string` | `domain/ipa/id, domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_host_uuid` | `string` | `domain/ipa/access, domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_member` | `string` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_memberof` | `string` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_name` | `string` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_objectclass` | `string` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostgroup_uuid` | `string` | `domain/ipa/access` | `-` | - | sssd-ipa.conf (api) |
| `ipa_hostname` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_master_domain_search_base` | `string` | `domain/ipa` | `the value of` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_domain` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_ext_host` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_host` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_of` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_member_user` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_name` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_object_class` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_netgroup_uuid` | `string` | `domain/ipa/id` | `-` | - | sssd-ipa.conf (api) |
| `ipa_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaOverrideAnchor` | - | sssd-ipa.conf (api) |
| `ipa_ranges_search_base` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_refresh` | `int` | `domain/ipa/access, domain/ipa` | `5 (seconds)` | - | sssd-ipa.conf (api) |
| `ipa_selinux_search_base` | `string` | `domain/ipa` | `the value of` | - | sssd-ipa.5.xml (man) |
| `ipa_selinux_usermap_enabled` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_host_category` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_member_host` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_member_user` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_name` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_object_class` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_see_also` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_selinux_user` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_user_category` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_selinux_usermap_uuid` | `string` | `domain/ipa/session` | `-` | - | sssd-ipa.conf (api) |
| `ipa_server` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_server_mode` | `bool` | `domain/ipa/id, domain/ipa` | `false` | - | sssd-ipa.conf (api) |
| `ipa_subdomains_search_base` | `string` | `domain/ipa/subdomains, domain/ipa` | `the value of` | - | sssd-ipa.conf (api) |
| `ipa_subid_ranges_search_base` | `string` | `domain/ipa` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_memberof` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_object_class` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_sudocmd` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmd_uuid` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_entry_usn` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_member` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_name` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_object_class` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudocmdgroup_uuid` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_allowcmd` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_cmdcategory` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_denycmd` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_enabled_flag` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_entry_usn` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_externaluser` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_host` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_hostcategory` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_name` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_notafter` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_notbefore` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_object_class` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_option` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextgroup` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextuser` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasextusergroup` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasgroup` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasgroupcategory` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_runasusercategory` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_sudoorder` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_user` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_usercategory` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_sudorule_uuid` | `string` | `domain/ipa/sudo` | `-` | - | sssd-ipa.conf (api) |
| `ipa_user_override_object_class` | `string` | `domain/ipa/id, domain/ipa` | `ipaUserOverride` | - | sssd-ipa.conf (api) |
| `ipa_view_class` | `string` | `domain/ipa/id, domain/ipa` | `nsContainer` | - | sssd-ipa.conf (api) |
| `ipa_view_name` | `string` | `domain/ipa/id, domain/ipa` | `cn` | - | sssd-ipa.conf (api) |
| `ipa_views_search_base` | `string` | `domain/ipa/id, domain/ipa` | `the value of` | - | sssd-ipa.conf (api) |
| `krb5_auth_timeout` | `int` | `domain/ad, domain/ipa, domain/krb5` | `-` | - | sssd-ad.conf (api) |
| `krb5_backup_kpasswd` | `string` | `domain/ad/chpass, domain/ipa, domain/krb5` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_backup_server` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `krb5_canonicalize` | `bool` | `domain/ad, domain/ipa, domain/krb5/auth, domain/ldap` | `false` | - | sssd-ad.conf (api) |
| `krb5_ccachedir` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `/tmp` | - | sssd-ad.conf (api) |
| `krb5_ccname_template` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `-` | - | sssd-ad.conf (api) |
| `krb5_confd_path` | `string` | `domain/ad, domain/ipa` | `not set (krb5.include.d subdirectory of` | - | sssd-ad.conf (api) |
| `krb5_fast_principal` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `-` | - | sssd-ad.conf (api) |
| `krb5_fast_use_anonymous_pkinit` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - | sssd-ad.conf (api) |
| `krb5_kdcinfo_lookahead` | `string` | `domain/krb5` | `3:1` | - | sssd-krb5.5.xml (man) |
| `krb5_kdcip` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `krb5_keytab` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `krb5_kpasswd` | `string` | `domain/ad/chpass, domain/ipa, domain/krb5` | `Use the KDC` | - | sssd-ad.conf (api) |
| `krb5_lifetime` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - | sssd-ad.conf (api) |
| `krb5_map_user` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `not set` | - | sssd-ad.conf (api) |
| `krb5_rcache_dir` | `string` | `sssd` | `Distribution-specific and specified` | - | sssd.api.conf (api) |
| `krb5_realm` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `System defaults, see` | - | sssd-ad.conf (api) |
| `krb5_renew_interval` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - | sssd-ad.conf (api) |
| `krb5_renewable_lifetime` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - | sssd-ad.conf (api) |
| `krb5_server` | `string` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `krb5_store_password_if_offline` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - | sssd-ad.conf (api) |
| `krb5_use_enterprise_principal` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false (AD provider: true)` | - | sssd-ad.conf (api) |
| `krb5_use_fast` | `string` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `not set, i.e. FAST is not used.` | never, try, demand | sssd-ad.conf (api) |
| `krb5_use_kdcinfo` | `bool` | `domain/ad, domain/ipa, domain/krb5, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `krb5_use_subdomain_realm` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth, domain/krb5` | `false` | - | sssd-ad.conf (api) |
| `krb5_validate` | `bool` | `domain/ad/auth, domain/ipa/auth, domain/krb5/auth` | `-` | - | sssd-ad.conf (api) |
| `ldap_access_filter` | `string` | `domain/ldap/access, domain/ldap` | `Empty` | - | sssd-ldap.conf (api) |
| `ldap_access_order` | `string` | `domain/ldap/access, domain/ldap` | `filter` | - | sssd-ldap.conf (api) |
| `ldap_account_expire_policy` | `string` | `domain/ldap/access, domain/ldap` | `Empty` | - | sssd-ldap.conf (api) |
| `ldap_autofs_entry_key` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_object_class` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_entry_value` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_master_name` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs, domain/ldap` | `auto.master` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_name` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_map_object_class` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs` | `-` | - | sssd-ad.conf (api) |
| `ldap_autofs_search_base` | `string` | `domain/ad/autofs, domain/ipa/autofs, domain/ldap/autofs, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_backup_uri` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_chpass_backup_uri` | `string` | `domain/ldap/chpass, domain/ldap` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |
| `ldap_chpass_dns_service_name` | `string` | `domain/ldap/chpass, domain/ldap` | `not set, i.e. service discovery is disabled` | - | sssd-ldap.conf (api) |
| `ldap_chpass_update_last_change` | `bool` | `domain/ldap/chpass, domain/ldap` | `False` | - | sssd-ldap.conf (api) |
| `ldap_chpass_uri` | `string` | `domain/ldap/chpass, domain/ldap` | `empty, i.e. ldap_uri is used.` | - | sssd-ldap.conf (api) |
| `ldap_connection_expire_offset` | `int` | `domain/ad, domain/ipa, domain/ldap` | `0` | - | sssd-ad.conf (api) |
| `ldap_connection_expire_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_connection_idle_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_default_authtok` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_default_authtok_type` | `string` | `domain/ad, domain/ipa, domain/ldap` | `password` | - | sssd-ad.conf (api) |
| `ldap_default_bind_dn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_deref` | `string` | `domain/ad, domain/ipa, domain/ldap` | `Empty (this is handled as` | never, searching, finding, always | sssd-ad.conf (api) |
| `ldap_deref_threshold` | `int` | `domain/ad, domain/ipa, domain/ldap` | `10` | - | sssd-ad.conf (api) |
| `ldap_disable_paging` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `False` | - | sssd-ad.conf (api) |
| `ldap_disable_range_retrieval` | `bool` | `domain/ldap` | `False` | - | sssd-ldap.conf (api) |
| `ldap_dns_service_name` | `string` | `domain/ad, domain/ipa, domain/ldap` | `ldap` | - | sssd-ad.conf (api) |
| `ldap_entry_usn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_enumeration_refresh_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `300` | - | sssd-ad.conf (api) |
| `ldap_enumeration_search_timeout` | `int` | `domain/ldap/id, domain/ldap` | `60` | - | sssd-ldap.conf (api) |
| `ldap_force_upper_case_realm` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - | sssd-ad.conf (api) |
| `ldap_group_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_external_member` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_gid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_member` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_modify_timestamp` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_nesting_level` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `2` | - | sssd-ad.conf (api) |
| `ldap_group_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_objectsid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_filter` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_search_scope` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_type` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_group_uuid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_host_search_base` | `string` | `domain/ldap` | `the value of` | - | sssd-ldap.5.xml (man) |
| `ldap_id_mapping` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - | sssd-ad.conf (api) |
| `ldap_id_use_start_tls` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `ldap_idmap_autorid_compat` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_default_domain_sid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_helper_table_size` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_max` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_min` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_idmap_range_size` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_ignore_unreadable_references` | `bool` | `domain/ldap` | `False` | - | sssd-ldap.conf (api) |
| `ldap_iphost_entry_usn` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_name` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_number` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_object_class` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_iphost_search_base` | `string` | `domain/ad/resolver, domain/ldap/resolver, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_entry_usn` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_name` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_number` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_object_class` | `string` | `domain/ad/resolver, domain/ldap/resolver` | `-` | - | sssd-ad.conf (api) |
| `ldap_ipnetwork_search_base` | `string` | `domain/ad/resolver, domain/ldap/resolver, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_krb5_init_creds` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `ldap_krb5_keytab` | `string` | `domain/ad, domain/ipa, domain/ldap` | `System keytab, normally` | - | sssd-ad.conf (api) |
| `ldap_krb5_ticket_lifetime` | `int` | `domain/ad, domain/ipa, domain/ldap` | `86400 (24 hours)` | - | sssd-ad.conf (api) |
| `ldap_library_debug_level` | `int` | `domain/ldap/id, domain/ldap` | `0 (libldap debugging disabled)` | - | sssd-ldap.conf (api) |
| `ldap_max_id` | `int` | `domain/ldap/id, domain/ldap` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_min_id` | `int` | `domain/ldap/id, domain/ldap` | `not set (both options are set to 0)` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_member` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_modify_timestamp` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_name` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_object_class` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_netgroup_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_netgroup_triple` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_network_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `6` | - | sssd-ad.conf (api) |
| `ldap_ns_account_lock` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_offline_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_opt_timeout` | `int` | `domain/ad, domain/ipa, domain/ldap` | `8` | - | sssd-ad.conf (api) |
| `ldap_page_size` | `int` | `domain/ad, domain/ipa, domain/ldap` | `1000` | - | sssd-ad.conf (api) |
| `ldap_ppolicy_pwd_change_threshold` | `int` | `domain/ldap` | `0` | - | sssd-ldap.conf (api) |
| `ldap_purge_cache_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_pwd_attribute` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_pwd_policy` | `string` | `domain/ad/auth, domain/ipa/auth, domain/ldap/auth, domain/ldap` | `none` | - | sssd-ad.conf (api) |
| `ldap_pwdlockout_dn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `cn=ppolicy,ou=policies,$ldap_search_base` | - | sssd-ad.conf (api) |
| `ldap_pwmodify_mode` | `string` | `domain/ad, domain/ipa, domain/ldap` | `exop` | ldap_modify, exop | sssd-ad.conf (api) |
| `ldap_read_rootdse` | `string` | `domain/ldap` | `anonymous` | - | sssd-ldap.5.xml (man) |
| `ldap_referrals` | `bool` | `domain/ad, domain/ipa, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `ldap_rfc2307_fallback_to_local_users` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `false` | - | sssd-ad.conf (api) |
| `ldap_rootdse_last_usn` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_sasl_authid` | `string` | `domain/ad, domain/ipa, domain/ldap` | `host/hostname@REALM` | - | sssd-ad.conf (api) |
| `ldap_sasl_canonicalize` | `bool` | `domain/ldap` | `false;` | - | sssd-ldap.conf (api) |
| `ldap_sasl_maxssf` | `int` | `domain/ad, domain/ipa, domain/ldap` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_mech` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - | sssd-ad.conf (api) |
| `ldap_sasl_minssf` | `int` | `domain/ad, domain/ipa, domain/ldap` | `Use the system default (usually specified` | - | sssd-ad.conf (api) |
| `ldap_sasl_realm` | `string` | `domain/ldap` | `the value of krb5_realm.` | - | sssd-ldap.5.xml (man) |
| `ldap_schema` | `string` | `domain/ad, domain/ipa, domain/ldap` | `rfc2307` | rfc2307, rfc2307bis, ipa, ad | sssd-ad.conf (api) |
| `ldap_search_base` | `string` | `domain/ad, domain/ipa, domain/ldap` | `If not set, the value of the` | - | sssd-ad.conf (api) |
| `ldap_search_timeout` | `int` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `6` | - | sssd-ad.conf (api) |
| `ldap_service_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_port` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_proto` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_service_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_subid_ranges_search_base` | `string` | `domain/ldap/id, domain/ldap` | `the value of` | - | sssd-ldap.conf (api) |
| `ldap_sudo_full_refresh_interval` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `21600 (6 hours)` | - | sssd-ad.conf (api) |
| `ldap_sudo_hostnames` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_netgroups` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudo_include_regexp` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `false` | - | sssd-ad.conf (api) |
| `ldap_sudo_ip` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `not specified` | - | sssd-ad.conf (api) |
| `ldap_sudo_random_offset` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `0 (disabled)` | - | sssd-ad.conf (api) |
| `ldap_sudo_search_base` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudo_smart_refresh_interval` | `int` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `900 (15 minutes)` | - | sssd-ad.conf (api) |
| `ldap_sudo_use_host_filter` | `bool` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo, domain/ldap` | `true` | - | sssd-ad.conf (api) |
| `ldap_sudorule_command` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_host` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_name` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notafter` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_notbefore` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_object_class_attr` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_option` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_order` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runas` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasgroup` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_runasuser` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_sudorule_user` | `string` | `domain/ad/sudo, domain/ipa/sudo, domain/ldap/sudo` | `-` | - | sssd-ad.conf (api) |
| `ldap_tls_cacert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cacertdir` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_cert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_cipher_suite` | `string` | `domain/ad, domain/ipa, domain/ldap` | `use OpenLDAP defaults, typically in` | - | sssd-ad.conf (api) |
| `ldap_tls_key` | `string` | `domain/ad, domain/ipa, domain/ldap` | `not set` | - | sssd-ad.conf (api) |
| `ldap_tls_reqcert` | `string` | `domain/ad, domain/ipa, domain/ldap` | `hard` | never, allow, try, demand, hard | sssd-ad.conf (api) |
| `ldap_uri` | `string` | `domain/ad, domain/ipa, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_use_ppolicy` | `bool` | `domain/ldap` | `true` | - | sssd-ldap.conf (api) |
| `ldap_use_tokengroups` | `bool` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `True for AD and IPA otherwise False.` | - | sssd-ad.conf (api) |
| `ldap_user_ad_account_expires` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_ad_user_account_control` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_auth_type` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_authorized_host` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_authorized_rhost` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_authorized_service` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_certificate` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_email` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_entry_usn` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_extra_attrs` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_fullname` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gecos` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_gid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_home_directory` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_last_pwd_change` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_krb_password_expiration` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_member_of` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_modify_timestamp` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_name` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_nds_login_allowed_time_map` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_nds_login_disabled` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_nds_login_expiration_time` | `string` | `domain/ldap/id` | `-` | - | sssd-ldap.conf (api) |
| `ldap_user_object_class` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_objectsid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_passkey` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_primary_group` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_principal` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_base` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id, domain/ldap` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_filter` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_search_scope` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_expire` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_flag` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_inactive` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_last_change` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_max` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_min` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shadow_warning` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_shell` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_ssh_public_key` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uid_number` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `ldap_user_uuid` | `string` | `domain/ad/id, domain/ipa/id, domain/ldap/id` | `-` | - | sssd-ad.conf (api) |
| `local_auth_policy` | `string` | `domain` | `match` | match, only, enable, disable | sssd.api.conf (api) |
| `lookup_family_order` | `string` | `domain` | `ipv4_first` | - | sssd.api.conf (api) |
| `maprule` | `string` | `certmap` | `-` | - | sssd.conf.5.xml (man) |
| `matchrule` | `string` | `certmap` | `KRB5:&lt;EKU&gt;clientAuth, i.e. only` | - | sssd.conf.5.xml (man) |
| `max_ccache_size` | `int` | `kcm` | `65536` | - | sssd-kcm.8.xml (man) |
| `max_ccaches` | `int` | `kcm` | `0 (unlimited, only the per-UID quota is enforced)` | - | sssd-kcm.8.xml (man) |
| `max_id` | `int` | `domain` | `1 for min_id, 0 (no limit) for max_id` | - | sssd.api.conf (api) |
| `max_uid_ccaches` | `int` | `kcm` | `64` | - | sssd-kcm.8.xml (man) |
| `memcache_size_group` | `int` | `nss` | `6` | - | sssd.conf.5.xml (man) |
| `memcache_size_initgroups` | `int` | `nss` | `10` | - | sssd.conf.5.xml (man) |
| `memcache_size_passwd` | `int` | `nss` | `8` | - | sssd.conf.5.xml (man) |
| `memcache_size_sid` | `int` | `nss` | `6` | - | sssd.conf.5.xml (man) |
| `memcache_timeout` | `int` | `nss` | `300` | - | sssd.api.conf (api) |
| `min_id` | `int` | `domain` | `1 for min_id, 0 (no limit) for max_id` | - | sssd.api.conf (api) |
| `monitor_resolv_conf` | `bool` | `sssd` | `true` | - | sssd.api.conf (api) |
| `offline_credentials_expiration` | `int` | `pam` | `0 (No limit)` | - | sssd.api.conf (api) |
| `offline_failed_login_attempts` | `int` | `pam` | `0 (No limit)` | - | sssd.api.conf (api) |
| `offline_failed_login_delay` | `int` | `pam` | `5` | - | sssd.api.conf (api) |
| `offline_timeout` | `int` | `domain` | `60` | - | sssd.api.conf (api) |
| `offline_timeout_max` | `int` | `domain` | `3600` | - | sssd.api.conf (api) |
| `offline_timeout_random_offset` | `int` | `domain` | `30` | - | sssd.api.conf (api) |
| `override_gid` | `int` | `domain` | `not set (Use the primary GID value retrieved` | - | sssd.api.conf (api) |
| `override_homedir` | `string` | `nss, domain` | `-` | - | sssd.api.conf (api) |
| `override_shell` | `string` | `nss, domain` | `not set (SSSD will use the value` | - | sssd.api.conf (api) |
| `override_space` | `string` | `sssd` | `not set (spaces will not be replaced)` | - | sssd.api.conf (api) |
| `p11_child_timeout` | `int` | `pam` | `10` | - | sssd.api.conf (api) |
| `p11_uri` | `string` | `pam` | `none` | - | sssd.api.conf (api) |
| `p11_wait_for_card_timeout` | `int` | `pam` | `60` | - | sssd.api.conf (api) |
| `pac_check` | `string` | `pac` | `-` | no_check, pac_present, check_upn, check_upn_allow_missing, upn_dns_info_present, check_upn_dns_info_ex, upn_dns_info_ex_present | sssd.api.conf (api) |
| `pac_lifetime` | `int` | `pac` | `300` | - | sssd.api.conf (api) |
| `pam_account_expired_message` | `string` | `pam` | `none` | - | sssd.api.conf (api) |
| `pam_account_locked_message` | `string` | `pam` | `none` | - | sssd.api.conf (api) |
| `pam_app_services` | `string` | `pam` | `Not set` | - | sssd.api.conf (api) |
| `pam_cert_auth` | `bool` | `pam` | `False` | - | sssd.api.conf (api) |
| `pam_cert_db_path` | `string` | `pam` | `-` | - | sssd.api.conf (api) |
| `pam_cert_verification` | `string` | `pam` | `not set, i.e. use default` | - | sssd.api.conf (api) |
| `pam_gssapi_check_upn` | `bool` | `pam, domain` | `True` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_apply` | `string` | `pam, domain` | `not set` | - | sssd.api.conf (api) |
| `pam_gssapi_indicators_map` | `string` | `pam, domain` | `not set (use of authentication indicators is not required)` | - | sssd.api.conf (api) |
| `pam_gssapi_services` | `string` | `pam, domain` | `- (GSSAPI authentication is disabled)` | - | sssd.api.conf (api) |
| `pam_id_timeout` | `int` | `pam` | `5` | - | sssd.api.conf (api) |
| `pam_initgroups_scheme` | `string` | `pam` | `-` | always, no_session, never | sssd.api.conf (api) |
| `pam_json_services` | `string` | `pam` | `- (JSON protocol is disabled)` | - | sssd.api.conf (api) |
| `pam_p11_allowed_services` | `string` | `pam` | `the default set of PAM service names` | - | sssd.api.conf (api) |
| `pam_passkey_auth` | `bool` | `pam` | `True` | - | sssd.api.conf (api) |
| `pam_public_domains` | `string` | `pam` | `none` | - | sssd.api.conf (api) |
| `pam_pwd_expiration_warning` | `int` | `pam` | `0` | - | sssd.api.conf (api) |
| `pam_response_filter` | `string` | `pam` | `-` | env | sssd.api.conf (api) |
| `pam_trusted_users` | `string` | `pam` | `All users are considered trusted` | - | sssd.api.conf (api) |
| `pam_verbosity` | `int` | `pam` | `1` | - | sssd.api.conf (api) |
| `passkey_child_timeout` | `int` | `pam` | `15` | - | sssd.api.conf (api) |
| `passkey_debug_libfido2` | `bool` | `pam` | `False` | - | sssd.api.conf (api) |
| `passkey_verification` | `string` | `sssd` | `-` | user_verification | sssd.api.conf (api) |
| `priority` | `int` | `certmap` | `the lowest priority` | - | sssd.conf.5.xml (man) |
| `proxy_fast_alias` | `bool` | `domain/proxy/id, domain` | `false` | - | sssd-proxy.conf (api) |
| `proxy_lib_name` | `string` | `domain/proxy/id, domain` | `-` | - | sssd-proxy.conf (api) |
| `proxy_max_children` | `int` | `domain/proxy, domain` | `10` | - | sssd-proxy.conf (api) |
| `proxy_pam_target` | `string` | `domain/proxy/auth, domain` | `not set by default, you have to take an` | - | sssd-proxy.conf (api) |
| `proxy_resolver_lib_name` | `string` | `domain` | `-` | - | sssd.conf.5.xml (man) |
| `pwd_expiration_warning` | `int` | `domain` | `7 (Kerberos), 0 (LDAP)` | - | sssd.api.conf (api) |
| `pwfield` | `string` | `nss` | `-` | - | sssd.api.conf (api) |
| `re_expression` | `string` | `sssd, domain` | `-` | - | sssd.api.conf (api) |
| `realmd_tags` | `string` | `domain` | `-` | - | sssd.api.conf (api) |
| `refresh_expired_interval` | `int` | `domain` | `0 (disabled)` | - | sssd.api.conf (api) |
| `refresh_expired_interval_offset` | `int` | `domain` | `-` | - | sssd.api.conf (api) |
| `resolver_provider` | `string` | `domain` | `The value of` | - | sssd.api.conf (api) |
| `responder_idle_timeout` | `int` | `service, *` | `300` | - | sssd.api.conf (api) |
| `scope` | `string` | `session_recording` | `-` | - | sssd.api.conf (api) |
| `selinux_provider` | `string` | `domain` | `the value of` | - | sssd.api.conf (api) |
| `services` | `list` | `sssd` | `nss` | - | sssd.api.conf (api) |
| `session_provider` | `string` | `domain` | `-` | - | sssd.api.conf (api) |
| `shell_fallback` | `string` | `nss` | `/bin/sh` | - | sssd.api.conf (api) |
| `simple_allow_groups` | `string` | `domain/simple/access, domain/simple` | `-` | - | sssd-simple.conf (api) |
| `simple_allow_users` | `string` | `domain/simple/access, domain/simple` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_groups` | `string` | `domain/simple/access, domain/simple` | `-` | - | sssd-simple.conf (api) |
| `simple_deny_users` | `string` | `domain/simple/access, domain/simple` | `-` | - | sssd-simple.conf (api) |
| `socket_path` | `string` | `kcm` | `-` | - | sssd-kcm.8.xml (man) |
| `ssh_hash_known_hosts` | `bool` | `ssh` | `-` | - | sssd.api.conf (api) |
| `ssh_known_hosts_timeout` | `int` | `ssh` | `-` | - | sssd.api.conf (api) |
| `ssh_use_certificate_keys` | `bool` | `ssh` | `true` | - | sssd.api.conf (api) |
| `ssh_use_certificate_matching_rules` | `string` | `ssh` | `not set, equivalent to 'all_rules',` | - | sssd.api.conf (api) |
| `subdomain_homedir` | `string` | `domain` | `-` | - | sssd.api.conf (api) |
| `subdomain_inherit` | `string` | `domain` | `none` | - | sssd.api.conf (api) |
| `subdomain_refresh_interval` | `int` | `domain` | `-` | - | sssd.api.conf (api) |
| `subdomain_refresh_interval_offset` | `int` | `domain` | `-` | - | sssd.api.conf (api) |
| `subdomains_provider` | `string` | `domain` | `The value of` | - | sssd.api.conf (api) |
| `sudo_inverse_order` | `bool` | `sudo` | `-` | - | sssd.api.conf (api) |
| `sudo_provider` | `string` | `domain` | `The value of` | none, ad, ipa, ldap | sssd.api.conf (api) |
| `sudo_threshold` | `int` | `sudo` | `50` | - | sssd.api.conf (api) |
| `sudo_timed` | `bool` | `sudo` | `false` | - | sssd.api.conf (api) |
| `tgt_renewal` | `bool` | `kcm` | `False (Automatic renewals disabled)` | - | sssd-kcm.8.xml (man) |
| `tgt_renewal_inherit` | `string` | `kcm` | `NULL` | - | sssd-kcm.8.xml (man) |
| `timeout` | `int` | `service, domain, *` | `10` | - | sssd.api.conf (api) |
| `try_inotify` | `bool` | `sssd` | `true on platforms where inotify is` | - | sssd.api.conf (api) |
| `use_fully_qualified_names` | `bool` | `domain` | `FALSE (TRUE for trusted` | - | sssd.api.conf (api) |
| `user` | `string` | `sssd` | `-` | - | sssd.api.conf (api) |
| `user_attributes` | `string` | `nss, ifp` | `not set, fallback to InfoPipe option` | name, uidnumber, gidnumber, gecos, homedirectory, loginshell | sssd.api.conf (api) |
| `users` | `list` | `session_recording` | `Empty. Matches no users.` | - | sssd.api.conf (api) |
| `vetoed_shells` | `list` | `nss` | `Not set` | - | sssd.api.conf (api) |
| `wildcard_limit` | `int` | `domain/ad, domain/ipa, domain/ldap` | `1000 (often the size of one page)` | - | sssd-ad.conf (api) |

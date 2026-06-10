// Package types provides shared data types for SSSD Inspector
package types

// SSSDLogError holds the human-readable description and samples of the actual log lines
type SSSDLogError struct {
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

// TIDArticle represents a dynamically loaded Knowledge Base article
type TIDArticle struct {
	TIDID          string   `json:"tid_id"`
	Title          string   `json:"title"`
	URL            string   `json:"url"`
	Description    string   `json:"description"`
	LogPatterns    []string `json:"log_patterns"`
	ConfigPatterns []string `json:"config_patterns"`
	Evidence       []string `json:"-"` // Log lines that triggered the match
}

// ReportData holds the results of our analysis
type ReportData struct {
	AppVersion    string `json:"app_version"`
	Timestamp     string `json:"timestamp"`
	SupportCaseID string `json:"support_case_id"`
	KernelVersion string `json:"kernel_version"`
	SLESRlease    string `json:"sles_release"`
	SCCStatus     string `json:"scc_status"`

	// Virtualization & Hardware
	HardwareManufacturer string `json:"hardware_manufacturer"`
	HardwareModel        string `json:"hardware_model"`
	Hypervisor           string `json:"hypervisor"`
	VirtualIdentity      string `json:"virtual_identity"`
	MACType              string `json:"mac_type"`

	// Base Services
	SssdInstalled   bool     `json:"sssd_installed"`
	SssdConfigFound bool     `json:"sssd_config_found"`
	SssdService     string   `json:"sssd_service"`
	WinbindService  string   `json:"winbind_service"`
	NscdStatus      string   `json:"nscd_status"`
	NscdCaching     []string `json:"nscd_caching"`
	SSSDPackages    []string `json:"sssd_packages"`

	// Authentication Details
	NsswitchValid     bool     `json:"nsswitch_valid"`
	PamSssInstalled   bool     `json:"pam_sss_installed"`
	PamGDPRRestricted bool     `json:"pam_gdpr_restricted"`
	HostsIssues       []string `json:"hosts_issues"`
	Nameservers       []string `json:"nameservers"`
	SearchDomain      string   `json:"search_domain"`
	TimeService       string   `json:"time_service"`
	KerberosRealm     string   `json:"kerberos_realm"`
	KeytabFound       bool     `json:"keytab_found"`

	// SSSD Deep AD Configs
	ADProviderMode bool `json:"ad_provider_mode"`
	EnumerateIssue bool `json:"enumerate_issue"`
	UseFQDNSet     bool `json:"use_fqdn_set"`

	// Log Analysis & Problems
	SSSDLogErrors     []SSSDLogError `json:"sssd_log_errors"`
	SSSDConfigSnippet string         `json:"sssd_config_snippet"`
	MACDenialExamples []string       `json:"mac_denial_examples"`

	Problems    []string        `json:"problems"`
	Warnings    []string        `json:"warnings"`
	MatchedTIDs []TIDArticle    `json:"matched_tids"`
	Timeline    []TimelineEvent `json:"timeline"`
}

// TimelineEvent represents a single chronological log occurrence
type TimelineEvent struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	RawLog    string `json:"raw_log"`
}

// Package constants defines all application-wide constants for SSSD Inspector
package constants

import "time"

// Application constants
const (
	// App metadata
	AppName    = "SSSD Inspector"
	AppVersion = "0.2.2"

	// Window dimensions
	DefaultWindowWidth  = 1024
	DefaultWindowHeight = 768
	DefaultWindowTitle  = "SSSD Inspector"

	// Background color RGBA
	BackgroundR = 244
	BackgroundG = 244
	BackgroundB = 249
	BackgroundA = 255
)

// File processing constants
const (
	// Size limits
	DefaultMaxFileSize   = "100MB"
	DefaultMaxLineLength = "1MB"
	DefaultBufferSize    = "64KB"
	DefaultChunkSize     = "32KB"

	// Archive formats
	TXZFormat   = "txz"
	TarXZFormat = "tar.xz"
	TarGZFormat = "tar.gz"

	// Analysis timeout
	DefaultTimeout = "10m"

	// Progress steps
	DefaultProgressSteps = 10
)

// Performance constants
const (
	// Worker management
	DefaultMaxWorkers = 4
	DefaultGCPercent  = 100
)

// Report constants
const (
	// Default formats
	TXTFormat  = "txt"
	HTMLFormat = "html"
	PDFFormat  = "pdf"

	// Output settings
	DefaultOutputSuffix = "_report"
	DefaultOutputDir    = "./reports"

	// Template paths (now embedded via go:embed in pkg/report)
	DefaultHTMLTemplate = "embedded" // templates/report.html is embedded in the binary
	DefaultTXTTemplate  = "embedded" // built inline by report.BuildTextReport()
)

// CLI constants
const (
	// CLI flags
	FlagVersion   = "v"
	FlagAnalyze   = "analyze"
	FlagLogDir    = "logdir"
	FlagTXT       = "txt"
	FlagHTML      = "html"
	FlagAnonymize = "anonymize"

	// Flag descriptions
	DescVersion   = "Print program version"
	DescAnalyze   = "Path to the supportconfig directory or log file"
	DescLogDir    = "Path to a directory containing raw SSSD log files (e.g., /var/log/sssd)"
	DescTXT       = "Generate a TXT report"
	DescHTML      = "Generate an HTML report"
	DescAnonymize = "Redact PII (IPs, Domains) from the report"

	// Default behavior
	DefaultGenerateBothFormats = true
)

// Anonymization constants
const (
	// PII patterns
	IPv4Pattern  = `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	IPv6Pattern  = `(?i)\b(?:[a-f0-9]{1,4}:){7}[a-f0-9]{1,4}\b|\b(?:[a-f0-9]{1,4}:){1,7}:|\b:(?::[a-f0-9]{1,4}){1,7}\b`
	MACPattern   = `(?i)\b(?:[0-9a-f]{2}[:-]){5}[0-9a-f]{2}\b`
	EmailPattern = `(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b`

	// Replacement strings
	IPv4Replacement     = "XXX.XXX.XXX.XXX"
	IPv6Replacement     = "XXXX:XXXX::XXXX"
	MACReplacement      = "XX:XX:XX:XX:XX:XX"
	EmailReplacement    = "[REDACTED_USER]@example.com"
	DomainReplacement   = "example.com"
	HardwareReplacement = "[REDACTED]"

	// Anonymization markers
	RedactedMarker = "[REDACTED]"
)

// File patterns constants
const (
	// Relevant files for analysis
	NSSwitchConf        = "nsswitch.conf"
	HostsFile           = "hosts"
	NSCDConf            = "nscd.conf"
	SSSDConf            = "sssd.conf"
	SystemdTXT          = "systemd.txt"
	BasicEnvTXT         = "basic-environment.txt"
	UpdatesTXT          = "updates.txt"
	Y2LogTXT            = "y2log.txt"
	SSSDTXT             = "sssd.txt"
	RPMTXT              = "rpm.txt"
	ETCTXT              = "etc.txt"
	NetworkTXT          = "network.txt"
	NTPTXT              = "ntp.txt"
	PAMTXT              = "pam.txt"
	FSDiskIO_TXT        = "fs-diskio.txt"
	StorageTXT          = "storage.txt"
	SecurityAppArmorTXT = "security-apparmor.txt"
	SecuritySELinuxTXT  = "security-selinux.txt"
	MemoryTXT           = "memory.txt"
	SARTXT              = "sar.txt"
	MessagesFile        = "messages"
	MessagesTXT         = "messages.txt"
	BootTXT             = "boot.txt"
)

// Service status constants
const (
	StatusNotRunning         = "Not Running / Unknown"
	StatusNotRunningDisabled = "Not Running / Disabled"
	StatusDisabled           = "Disabled / Stopped"
	StatusNotConfigured      = "Not configured"
	StatusUnknown            = "Unknown"
	StatusUnknownNone        = "Unknown/None"
	StatusRunning            = "Running"
	StatusNARawLogMode       = "N/A (raw log mode)"
	StatusRedacted           = "[REDACTED]"
)

// Knowledge base constants
const (
	DefaultTIDDirectory    = "./kb_articles"
	DefaultRefreshInterval = "24h"
	DefaultAutoRefresh     = true
)

// Logging constants
const (
	// Log levels
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	// Log formats
	LogFormatJSON = "json"
	LogFormatText = "text"

	// Default settings
	DefaultLogLevel      = LogLevelInfo
	DefaultLogFormat     = LogFormatJSON
	DefaultLogPath       = "./logs/sssd-inspector.log"
	DefaultLogMaxSize    = "10MB"
	DefaultLogMaxBackups = 5
	DefaultLogCompress   = true
)

// File dialog constants
const (
	// Dialog titles
	TitleSelectArchive = "Select Supportconfig Archive"
	TitleSavePDF       = "Save PDF Report"
	TitleSaveTXT       = "Save TXT Report"

	// File filters
	FilterSupportconfig = "Supportconfig Archives (*.txz, *.tar.xz)"
	FilterAllFiles      = "All Files (*.*)"
	FilterPDF           = "PDF Document (*.pdf)"
	FilterText          = "Text Document (*.txt)"

	// Filter patterns
	PatternSupportconfig = "*.txz;*.tar.xz"
	PatternAllFiles      = "*.*"
	PatternPDF           = "*.pdf"
	PatternText          = "*.txt"

	// Default filenames
	DefaultPDFName = "SSSD_Analysis_Report.pdf"
	DefaultTXTName = "SSSD_Analysis_Report.txt"
)

// Event constants
const (
	// Event names
	EventAnalyzeProgress = "analyze-progress"

	// Event messages
	MsgInitializing     = "Initializing streaming engine..."
	MsgScanningHardware = "Scanning Hardware & OS Data..."
	MsgAnalyzingNetwork = "Analyzing Network & Kerberos state..."
	MsgEvaluatingConfig = "Evaluating SSSD Configurations..."
	MsgStreamingLogs    = "Streaming and Parsing SSSD Logs..."
	MsgMatchingKB       = "Matching Knowledge Base Articles..."
	MsgSanitizingPII    = "Sanitizing PII data..."
	MsgAnalysisComplete = "Analysis Complete!"
)

// Progress percentage constants
const (
	ProgressStart      = 0
	ProgressHardware   = 10
	ProgressNetwork    = 25
	ProgressConfig     = 40
	ProgressLogs       = 60
	ProgressKB         = 85
	ProgressSanitizing = 95
	ProgressComplete   = 100
)

// Timeout and extraction constants (single source of truth)
const (
	// File scan timeout per file
	DefaultFileScanTimeout = 30 * time.Second
	// Archive extraction timeout
	DefaultExtractionTimeout = 30 * time.Minute
	// Buffer size for extraction copies
	ExtractionBufferSize = 64 * 1024 // 64KB
	// Maximum allowed size for a single extracted file (tar bomb protection)
	MaxArchiveFileSize = 100 * 1024 * 1024 // 100MB
	// Threshold for triggering extra GC after extracting large archives
	LargeArchiveThreshold = 50 * 1024 * 1024 // 50MB
)

// Time format constants
const (
	// Timestamp formats
	TimestampFormat = "02:01:2006 15:04:05"

	// Default timeout duration
	DefaultTimeoutDuration = 30 * time.Minute
)

// LogFileNames returns the list of log files scanned for SSSD error patterns
func LogFileNames() []string {
	return []string{SSSDTXT, MessagesFile, MessagesTXT}
}

// Error messages
const (
	ErrConfigNotFound   = "configuration file not found, using defaults"
	ErrInvalidConfig    = "invalid configuration"
	ErrFileAccess       = "file access error"
	ErrInvalidFormat    = "invalid file format"
	ErrPermissionDenied = "permission denied"
	ErrParseError       = "parse error"
	ErrMemoryError      = "memory error"
	ErrTimeoutExceeded  = "timeout exceeded"
	ErrCancelled        = "operation cancelled"
)

// RelevantFiles returns the complete list of relevant files for analysis
func RelevantFiles() []string {
	return []string{
		NSSwitchConf, HostsFile, NSCDConf, SSSDConf,
		SystemdTXT, BasicEnvTXT, UpdatesTXT,
		Y2LogTXT, SSSDTXT, RPMTXT, ETCTXT,
		NetworkTXT, NTPTXT, PAMTXT, FSDiskIO_TXT,
		StorageTXT, SecurityAppArmorTXT, SecuritySELinuxTXT,
		MemoryTXT, SARTXT, MessagesFile, MessagesTXT, BootTXT,
	}
}

// ArchiveFormats returns the list of supported archive formats
func ArchiveFormats() []string {
	return []string{TXZFormat, TarXZFormat, TarGZFormat}
}

// DefaultReportFormats returns the default report formats
func DefaultReportFormats() []string {
	return []string{TXTFormat, HTMLFormat}
}

// GetProgressMessage returns the progress message for a given percentage
func GetProgressMessage(percentage int) string {
	switch {
	case percentage == ProgressStart:
		return MsgInitializing
	case percentage == ProgressHardware:
		return MsgScanningHardware
	case percentage == ProgressNetwork:
		return MsgAnalyzingNetwork
	case percentage == ProgressConfig:
		return MsgEvaluatingConfig
	case percentage == ProgressLogs:
		return MsgStreamingLogs
	case percentage == ProgressKB:
		return MsgMatchingKB
	case percentage == ProgressSanitizing:
		return MsgSanitizingPII
	case percentage == ProgressComplete:
		return MsgAnalysisComplete
	default:
		return "Processing..."
	}
}

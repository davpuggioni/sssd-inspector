// report.go
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
)

// reportAssetsFS pins the single NAMED "embed" import for package main.
//
// NOTE: main_gui.go also needs "embed" for `assets embed.FS`; Go forbids
// importing the same module twice in one package with different bindings,
// so every //go:embed Var in package main must share this one named import
// (never `import _ "embed"`).
var _ = embed.FS{}

// reportHTMLTemplate is the maintainable HTML report template (P6). It is
// embedded into the single binary via go:embed — no external files at
// runtime, no network, no JS framework.
//
//go:embed report_tpl/report.html.tpl
var reportHTMLTemplate string

func buildTextReport(report ReportData) string {
	var sb strings.Builder

	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString("             SUPPORTCONFIG SSSD ANALYSIS REPORT\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf(" Generated: %s\n", report.Timestamp))
	if report.SupportCaseID != "" {
		sb.WriteString(fmt.Sprintf(" Support Case (SR#): %s\n", report.SupportCaseID))
	}
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString("                  EXECUTIVE SUMMARY (TRIAGE)\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	for _, line := range summaryLines(report) {
		sb.WriteString(line + "\n")
	}
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	// Root-cause breakdown (P4a): group config + log signals per coarse cause.
	sb.WriteString("             ROOT-CAUSE BREAKDOWN (GROUPED)\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	for _, line := range buildRootCauseBreakdownLines(report) {
		sb.WriteString(line + "\n")
	}
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	sb.WriteString(fmt.Sprintf("[+] OS Release:        %s\n", report.SLESRlease))
	sb.WriteString(fmt.Sprintf("[+] Kernel:            %s\n", report.KernelVersion))
	sb.WriteString(fmt.Sprintf("[+] SCC Status:        %s\n", report.SCCStatus))
	sb.WriteString(fmt.Sprintf("[+] Hardware:          %s %s\n", report.HardwareManufacturer, report.HardwareModel))
	sb.WriteString(fmt.Sprintf("[+] Virtualization:    %s (Identity: %s)\n", report.Hypervisor, report.VirtualIdentity))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	sb.WriteString(fmt.Sprintf("[+] SSSD Installed:    %v\n", report.SssdInstalled))
	sb.WriteString(fmt.Sprintf("[+] SSSD Config:       %v\n", report.SssdConfigFound))
	sb.WriteString(fmt.Sprintf("[+] SSSD Service:      %s\n", report.SssdService))
	sb.WriteString(fmt.Sprintf("[+] Winbind Status:    %s\n", report.WinbindService))
	sb.WriteString(fmt.Sprintf("[+] NSCD Service:      %s\n", report.NscdStatus))
	if len(report.NscdCaching) > 0 {
		sb.WriteString(fmt.Sprintf("[!] NSCD Caching:      YES (%s)\n", strings.Join(report.NscdCaching, ", ")))
	} else {
		sb.WriteString("[+] NSCD Caching:      Safely Disabled for SSSD modules\n")
	}

	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString("               AD / KERBEROS INTEGRATION CHECKS\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	sb.WriteString(fmt.Sprintf("[+] DNS Nameservers:   %s\n", strings.Join(report.Nameservers, ", ")))
	sb.WriteString(fmt.Sprintf("[+] DNS Search Domain: %s\n", report.SearchDomain))
	sb.WriteString(fmt.Sprintf("[+] Time Service:      %s\n", report.TimeService))
	sb.WriteString(fmt.Sprintf("[+] Kerberos Realm:    %s\n", report.KerberosRealm))
	sb.WriteString(fmt.Sprintf("[+] Machine Keytab:    %v\n", report.KeytabFound))

	if report.PamGDPRRestricted {
		sb.WriteString("[!] PAM pam_sss.so:    Restricted (GDPR)\n")
	} else {
		sb.WriteString(fmt.Sprintf("[+] PAM pam_sss.so:    %v\n", report.PamSssInstalled))
	}
	sb.WriteString(fmt.Sprintf("[+] NSSwitch valid:    %v\n", report.NsswitchValid))

	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString(fmt.Sprintf("[+] SSSD AD Provider:  %v\n", report.ADProviderMode))
	sb.WriteString(fmt.Sprintf("[+] SSSD Use FQDN:     %v\n", report.UseFQDNSet))
	sb.WriteString(fmt.Sprintf("[!] SSSD Enumerate:    %v\n", report.EnumerateIssue))

	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString("                  INSTALLED SSSD PACKAGES\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	if len(report.SSSDPackages) > 0 {
		for _, pkg := range report.SSSDPackages {
			sb.WriteString(fmt.Sprintf(" [+] %s\n", pkg))
		}
	} else {
		sb.WriteString(" [-] No SSSD packages found.\n")
	}

	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString("                 SSSD LOG ERRORS (sssd.txt)\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")
	if len(report.SSSDLogErrors) == 0 {
		sb.WriteString(" [+] No critical AD/Kerberos errors found in SSSD logs.\n")
	} else {
		for _, errData := range report.SSSDLogErrors {
			sb.WriteString(fmt.Sprintf(" [!] %s\n", errData.Description))
			for _, line := range errData.Examples {
				sb.WriteString(fmt.Sprintf("     - %s\n", line))
			}
		}
	}

	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString("                     ACTIONABLE PROBLEMS\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	if len(report.Problems) == 0 {
		sb.WriteString(" No major SSSD/Auth issues detected based on the checks.\n")
	} else {
		for _, prob := range report.Problems {
			sb.WriteString(fmt.Sprintf(" [X] %s\n", prob))
		}
	}

	if len(report.MACDenialExamples) > 0 {
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		sb.WriteString("              APPARMOR / SELINUX DENIALS FOUND\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for _, line := range report.MACDenialExamples {
			sb.WriteString(fmt.Sprintf(" [!] %s\n", line))
		}
	}

	if len(report.Warnings) > 0 {
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		sb.WriteString("                 TUNING & DIAGNOSTIC HINTS\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		for _, warn := range report.Warnings {
			sb.WriteString(fmt.Sprintf(" [i] %s\n", warn))
		}
	}

	if len(report.MatchedTIDs) > 0 {
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		sb.WriteString("               KNOWLEDGE BASE ARTICLES (TIDs)\n")
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		for _, tid := range report.MatchedTIDs {
			sb.WriteString(fmt.Sprintf(" [KB] %s: %s\n", tid.TIDID, tid.Title))
			sb.WriteString(fmt.Sprintf("      Link: %s\n", tid.URL))
			sb.WriteString(fmt.Sprintf("      Desc: %s\n", tid.Description))
			if len(tid.Evidence) > 0 {
				sb.WriteString("      Log Evidence:\n")
				for _, ev := range tid.Evidence {
					sb.WriteString(fmt.Sprintf("       - %s\n", ev))
				}
			}
			sb.WriteString("\n")
		}
	}

	if len(report.TemporalClusters) > 0 {
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		sb.WriteString("        TEMPORAL CLUSTERS (RETRY LOOPS / FLAPPING)\n")
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		for _, c := range report.TemporalClusters {
			sb.WriteString(fmt.Sprintf(" [~] %s\n", c.Description))
			sb.WriteString(fmt.Sprintf("     %d occurrences between %s and %s\n", c.EventCount, c.WindowStart, c.WindowEnd))
			if c.SampleRawLog != "" {
				sb.WriteString(fmt.Sprintf("     Sample: %s\n", c.SampleRawLog))
			}
		}
		sb.WriteString("\n")
	}

	if len(report.KBSuggestions) > 0 {
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		sb.WriteString("        KB SUGGESTIONS (FUZZY TF-IDF MATCH)\n")
		sb.WriteString(strings.Repeat("=", 60) + "\n")
		for _, s := range report.KBSuggestions {
			sb.WriteString(fmt.Sprintf(" [?] %s: %s (similarity %.0f%%)\n", s.TIDID, s.Title, s.Score*100))
			sb.WriteString(fmt.Sprintf("      Link: %s\n", s.URL))
			if s.SampleLine != "" {
				sb.WriteString(fmt.Sprintf("      Matched log line: %s\n", s.SampleLine))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("=", 60) + "\n")
	sb.WriteString(fmt.Sprintf(" %s - SUSE Technical Support - Released under the GNU GPL v3 (see LICENSE).\n", toolSignature(report.AppVersion)))
	sb.WriteString(strings.Repeat("=", 60) + "\n")

	return sb.String()
}

// toolSignature returns the single, source-of-truth attribution for the footer
// of every report. The version is read from constants so it never drifts from
// the binary (main.go prints the same string for -v).
func toolSignature(version string) string {
	return fmt.Sprintf("sssd-inspector v%s", version)
}

// buildJSONReport serializes the full structured report (including the
// executive summary and the provenance-aware ConfigFindings) as pretty-printed
// JSON. This is the machine-readable export for tooling and case management.
func buildJSONReport(report ReportData) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func writeHTMLReportFile(report ReportData, filename string) {
	t, err := template.New("report").Funcs(template.FuncMap{
		"healthScoreClass": healthScoreClass,
		"mult":             func(a, b float64) float64 { return a * b },
		"sevClass": func(s Severity) string {
			switch s {
			case SevCritical:
				return "critical"
			case SevError:
				return "error"
			default:
				return "warning"
			}
		},
		"rootBreakdown":   rootCauseBreakdown,
		"timelineTotal":   timelineTotalOccurrences,
		"timelinePreview": timelinePreviewRows,
		"timelineShown":   timelineShownCount,
		"occurrenceCount": timelineRowOccurrences,
	}).Parse(reportHTMLTemplate)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create HTML report: %v", err)
		return
	}
	defer file.Close()

	if err := t.Execute(file, report); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

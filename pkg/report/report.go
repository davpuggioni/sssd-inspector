// Package report provides report generation for SSSD Inspector
package report

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"strings"

	"sssd-inspector/logger"
	"sssd-inspector/pkg/types"
)

//go:embed templates/report.html
var htmlReportTemplate string

// BuildTextReport generates a text report from analysis data
func BuildTextReport(report types.ReportData) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, "%s\n", "             SUPPORTCONFIG SSSD ANALYSIS REPORT")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, " Generated: %s\n", report.Timestamp)
	if report.SupportCaseID != "" {
		fmt.Fprintf(&sb, " Support Case (SR#): %s\n", report.SupportCaseID)
	}
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))

	fmt.Fprintf(&sb, "[+] OS Release:        %s\n", report.SLESRlease)
	fmt.Fprintf(&sb, "[+] Kernel:            %s\n", report.KernelVersion)
	fmt.Fprintf(&sb, "[+] SCC Status:        %s\n", report.SCCStatus)
	fmt.Fprintf(&sb, "[+] Hardware:          %s %s\n", report.HardwareManufacturer, report.HardwareModel)
	fmt.Fprintf(&sb, "[+] Virtualization:    %s (Identity: %s)\n", report.Hypervisor, report.VirtualIdentity)
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))

	fmt.Fprintf(&sb, "[+] SSSD Installed:    %v\n", report.SssdInstalled)
	fmt.Fprintf(&sb, "[+] SSSD Config:       %v\n", report.SssdConfigFound)
	fmt.Fprintf(&sb, "[+] SSSD Service:      %s\n", report.SssdService)
	fmt.Fprintf(&sb, "[+] Winbind Status:    %s\n", report.WinbindService)
	fmt.Fprintf(&sb, "[+] NSCD Service:      %s\n", report.NscdStatus)
	if len(report.NscdCaching) > 0 {
		fmt.Fprintf(&sb, "[!] NSCD Caching:      YES (%s)\n", strings.Join(report.NscdCaching, ", "))
	} else {
		fmt.Fprintf(&sb, "%s\n", "[+] NSCD Caching:      Safely Disabled for SSSD modules")
	}

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
	fmt.Fprintf(&sb, "%s\n", "               AD / KERBEROS INTEGRATION CHECKS")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))

	fmt.Fprintf(&sb, "[+] DNS Nameservers:   %s\n", strings.Join(report.Nameservers, ", "))
	fmt.Fprintf(&sb, "[+] DNS Search Domain: %s\n", report.SearchDomain)
	fmt.Fprintf(&sb, "[+] Time Service:      %s\n", report.TimeService)
	fmt.Fprintf(&sb, "[+] Kerberos Realm:    %s\n", report.KerberosRealm)
	fmt.Fprintf(&sb, "[+] Machine Keytab:    %v\n", report.KeytabFound)

	if report.PamGDPRRestricted {
		fmt.Fprintf(&sb, "%s\n", "[!] PAM pam_sss.so:    Restricted (GDPR)")
	} else {
		fmt.Fprintf(&sb, "[+] PAM pam_sss.so:    %v\n", report.PamSssInstalled)
	}
	fmt.Fprintf(&sb, "[+] NSSwitch valid:    %v\n", report.NsswitchValid)

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
	fmt.Fprintf(&sb, "[+] SSSD AD Provider:  %v\n", report.ADProviderMode)
	fmt.Fprintf(&sb, "[+] SSSD Use FQDN:     %v\n", report.UseFQDNSet)
	fmt.Fprintf(&sb, "[!] SSSD Enumerate:    %v\n", report.EnumerateIssue)

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, "%s\n", "                  INSTALLED SSSD PACKAGES")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	if len(report.SSSDPackages) > 0 {
		for _, pkg := range report.SSSDPackages {
			fmt.Fprintf(&sb, " [+] %s\n", pkg)
		}
	} else {
		fmt.Fprintf(&sb, "%s\n", " [-] No SSSD packages found.")
	}

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, "%s\n", "                 SSSD LOG ERRORS (sssd.txt)")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	if len(report.SSSDLogErrors) == 0 {
		fmt.Fprintf(&sb, "%s\n", " [+] No critical AD/Kerberos errors found in SSSD logs.")
	} else {
		for _, errData := range report.SSSDLogErrors {
			fmt.Fprintf(&sb, " [!] %s\n", errData.Description)
			for _, line := range errData.Examples {
				fmt.Fprintf(&sb, "     - %s\n", line)
			}
		}
	}

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, "%s\n", "                     ACTIONABLE PROBLEMS")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))

	if len(report.Problems) == 0 {
		fmt.Fprintf(&sb, "%s\n", " No major SSSD/Auth issues detected based on the checks.")
	} else {
		for _, prob := range report.Problems {
			fmt.Fprintf(&sb, " [X] %s\n", prob)
		}
	}

	if len(report.MACDenialExamples) > 0 {
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
		fmt.Fprintf(&sb, "%s\n", "              APPARMOR / SELINUX DENIALS FOUND")
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
		for _, line := range report.MACDenialExamples {
			fmt.Fprintf(&sb, " [!] %s\n", line)
		}
	}

	if len(report.Warnings) > 0 {
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
		fmt.Fprintf(&sb, "%s\n", "                 TUNING & DIAGNOSTIC HINTS")
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 60))
		for _, warn := range report.Warnings {
			fmt.Fprintf(&sb, " [i] %s\n", warn)
		}
	}

	if len(report.MatchedTIDs) > 0 {
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
		fmt.Fprintf(&sb, "%s\n", "               KNOWLEDGE BASE ARTICLES (TIDs)")
		fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
		for _, tid := range report.MatchedTIDs {
			fmt.Fprintf(&sb, " [KB] %s: %s\n", tid.TIDID, tid.Title)
			fmt.Fprintf(&sb, "      Link: %s\n", tid.URL)
			fmt.Fprintf(&sb, "      Desc: %s\n", tid.Description)
			if len(tid.Evidence) > 0 {
				fmt.Fprintf(&sb, "%s\n", "      Log Evidence:")
				for _, ev := range tid.Evidence {
					fmt.Fprintf(&sb, "       - %s\n", ev)
				}
			}
			fmt.Fprintf(&sb, "\n")
		}
	}

	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))
	fmt.Fprintf(&sb, " sssd-inspector v%s - SUSE Technical Support -Created by Davide M. Puggioni with Gemini Pro - 2026 - Released under the GNU GPL v3.\n", report.AppVersion)
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("=", 60))

	return sb.String()
}

// WriteHTMLReportFile generates an HTML report file using the embedded template
func WriteHTMLReportFile(report types.ReportData, filename string) {
	t, err := template.New("report").Parse(htmlReportTemplate)
	if err != nil {
		logger.Error("template parsing error", err, logger.Fields{"filename": filename})
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		logger.Error("failed to create HTML report", err, logger.Fields{"filename": filename})
		return
	}
	defer file.Close()

	if err := t.Execute(file, report); err != nil {
		logger.Error("template execution error", err, logger.Fields{"filename": filename})
	}
}

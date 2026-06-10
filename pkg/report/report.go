// Package report provides report generation for SSSD Inspector
package report

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"sssd-inspector/pkg/types"
)

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

// WriteHTMLReportFile generates an HTML report file
func WriteHTMLReportFile(report types.ReportData, filename string) {
	const tpl = `
<!DOCTYPE html>
<html>
<head>
<title>SSSD Supportconfig Analysis</title>
<style>
	body { font-family: Arial, sans-serif; background-color: #f4f4f9; color: #333; margin: 40px; }
	h1 { color: #0056b3; border-bottom: 2px solid #0056b3; padding-bottom: 10px; margin-bottom: 5px; }
	.timestamp { color: #666; font-size: 0.9em; margin-bottom: 20px; }
	h2 { color: #d9534f; border-bottom: 1px solid #d9534f; padding-bottom: 5px; margin-top: 30px;}
	h2.warn-header { color: #17a2b8; border-bottom: 1px solid #17a2b8; }
	h2.kb-header { color: #28a745; border-bottom: 1px solid #28a745; }
	.info-table { border-collapse: collapse; width: 100%; max-width: 900px; background: #fff; box-shadow: 0 0 10px rgba(0,0,0,0.1); margin-bottom: 20px;}
	.info-table th, .info-table td { padding: 12px 15px; border: 1px solid #ddd; text-align: left; }
	.info-table th { background-color: #0056b3; color: white; width: 35%; }
	.section-title { background-color: #e9ecef !important; color: #333 !important; font-weight: bold; text-align: center; }
	.problem-list { background: #ffebee; padding: 20px; border-left: 5px solid #d9534f; list-style-type: square; }
	.problem-list li { margin-bottom: 10px; font-weight: bold; }
	.warn-list { background: #e2f3f5; padding: 20px; border-left: 5px solid #17a2b8; list-style-type: square; }
	.warn-list li { margin-bottom: 10px; font-weight: normal; color: #0c5460; }
	.kb-list { background: #e8f5e9; padding: 20px; border-left: 5px solid #28a745; list-style-type: none; }
	.kb-list li { margin-bottom: 15px; }
	.kb-title { font-weight: bold; color: #155724; font-size: 1.1em; }
	.kb-desc { color: #333; margin-top: 5px; }
	.success { color: green; font-weight: bold;}
	.fail { color: red; font-weight: bold;}
	.warn { color: #d39e00; font-weight: bold;}

	details summary { cursor: pointer; font-weight: bold; color: #555; padding: 5px 0; outline: none; transition: color 0.2s;}
	details summary:hover { color: #000; }
	.log-block { margin-top: 5px; background: #f8f9fa; padding: 10px; border-left: 3px solid #d9534f; font-family: monospace; font-size: 0.85em; overflow-x: auto; color: #333;}
	.mac-block { background: #fff3cd; border-left: 3px solid #d39e00; }

	.pkg-list { margin: 0; padding-left: 20px; font-family: monospace; font-size: 0.9em; }
	.footer { margin-top: 40px; text-align: center; font-size: 0.85em; color: #777; border-top: 1px solid #ddd; padding-top: 10px; }
</style>
</head>
<body>
	<h1>Supportconfig SSSD Analysis Report</h1>
	<div class="timestamp">Generated on: {{.Timestamp}}</div>

	<table class="info-table">
		<tr><td colspan="2" class="section-title">System & Virtualization Information</td></tr>
		<tr><th>OS Release</th><td>{{.SLESRlease}}</td></tr>
		<tr><th>Kernel Version</th><td>{{.KernelVersion}}</td></tr>
		<tr><th>SCC Status</th><td>{{.SCCStatus}}</td></tr>
		<tr><th>Hardware</th><td>{{.HardwareManufacturer}} {{.HardwareModel}}</td></tr>
		<tr><th>Virtualization</th><td>{{.Hypervisor}} (Identity: {{.VirtualIdentity}})</td></tr>

		<tr><td colspan="2" class="section-title">Base Authentication Services</td></tr>
		<tr><th>SSSD Installed</th><td>{{.SssdInstalled}}</td></tr>
		<tr><th>SSSD Config Found</th><td>{{if .SssdConfigFound}}<span class="success">Yes</span>{{else}}<span class="fail">No</span>{{end}}</td></tr>
		<tr><th>SSSD Service Status</th><td>{{.SssdService}}</td></tr>
		<tr><th>Winbind Status</th><td>{{.WinbindService}}</td></tr>
		<tr><th>NSCD Service Status</th><td>{{.NscdStatus}}</td></tr>
		<tr><th>NSSwitch Valid</th><td>{{if .NsswitchValid}}<span class="success">Yes</span>{{else}}<span class="fail">No</span>{{end}}</td></tr>
		<tr><th>PAM pam_sss.so Found</th><td>{{if .PamGDPRRestricted}}<span class="warn">Restricted (GDPR)</span>{{else if .PamSssInstalled}}<span class="success">Yes</span>{{else}}<span class="fail">No</span>{{end}}</td></tr>

		<tr><td colspan="2" class="section-title">Installed SSSD Packages</td></tr>
		<tr><td colspan="2">
			{{if .SSSDPackages}}
			<ul class="pkg-list">
				{{range .SSSDPackages}}<li>{{.}}</li>{{end}}
			</ul>
			{{else}}
			<span class="warn">No SSSD packages found.</span>
			{{end}}
		</td></tr>

		<tr><td colspan="2" class="section-title">Active Directory & Kerberos Prerequisites</td></tr>
		<tr><th>DNS Nameservers</th><td>
			<ul style="margin: 0; padding-left: 20px;">
			{{range .Nameservers}}<li>{{.}}</li>{{end}}
			</ul>
		</td></tr>
		<tr><th>DNS Search Domain</th><td>{{.SearchDomain}}</td></tr>
		<tr><th>Time Sync Service</th><td>{{.TimeService}}</td></tr>
		<tr><th>Kerberos Realm</th><td>{{.KerberosRealm}}</td></tr>
		<tr><th>Machine Keytab Verified</th><td>{{if .KeytabFound}}<span class="success">Yes</span>{{else}}<span class="fail">No</span>{{end}}</td></tr>

		<tr><td colspan="2" class="section-title">Deep SSSD Configuration</td></tr>
		<tr><th>AD Provider Explicitly Set</th><td>{{.ADProviderMode}}</td></tr>
		<tr><th>Use Fully Qualified Names</th><td>{{.UseFQDNSet}}</td></tr>
		<tr><th>Enumerate Set to True</th><td>{{if .EnumerateIssue}}<span class="fail">Yes (Performance Risk)</span>{{else}}<span class="success">No</span>{{end}}</td></tr>

		<tr><td colspan="2" class="section-title">SSSD Log Errors (sssd.txt)</td></tr>
		{{if .SSSDLogErrors}}
			{{range .SSSDLogErrors}}
			<tr><td colspan="2"><span class="fail">Error Detected:</span> {{.Description}}
				<details>
					<summary>View Log Snippets</summary>
					<div class="log-block">
						{{range .Examples}}
						<div style="margin-bottom: 4px;">{{.}}</div>
						{{end}}
					</div>
				</details>
			</td></tr>
			{{end}}
		{{else}}
			<tr><td colspan="2"><span class="success">No critical AD/Kerberos errors found in SSSD logs.</span></td></tr>
		{{end}}
	</table>

	<h2>Actionable Problems Found</h2>
	{{if .Problems}}
		<ul class="problem-list">
		{{range .Problems}}
			<li>{{.}}</li>
		{{end}}
		</ul>
	{{else}}
		<p class="success" style="font-size: 1.2em;">No major SSSD/Auth issues detected based on the analysis!</p>
	{{end}}

	{{if .MACDenialExamples}}
		<h2 style="color: #d39e00; border-bottom: 1px solid #d39e00;">AppArmor / SELinux Denials Found</h2>
		<details>
			<summary style="color: #d39e00;">View Blocked Access Logs</summary>
			<div class="log-block mac-block">
				{{range .MACDenialExamples}}
					<div style="margin-bottom: 4px;">{{.}}</div>
				{{end}}
			</div>
		</details>
	{{end}}

	{{if .Warnings}}
		<h2 class="warn-header">Tuning & Diagnostic Hints</h2>
		<ul class="warn-list">
		{{range .Warnings}}
			<li>{{.}}</li>
		{{end}}
		</ul>
	{{end}}

	{{if .MatchedTIDs}}
<h2 style="color: #2e7d32; border-bottom: 2px solid #2e7d32; padding-bottom: 5px;">Knowledge Base Articles (TIDs)</h2>
<div style="background-color: #e8f5e9; border-left: 5px solid #2e7d32; padding: 15px; margin-bottom: 20px;">
    {{range .MatchedTIDs}}
    <div style="margin-bottom: 20px;">
        <strong><a href="{{.URL}}" target="_blank" style="color: #1565c0; text-decoration: none; font-size: 1.1em;">[{{.TIDID}}] {{.Title}}</a></strong>

        <div style="white-space: pre-wrap; font-family: 'Courier New', Courier, monospace; margin-top: 8px; font-size: 0.95em; color: #333; line-height: 1.4;">{{.Description}}</div>

        {{if .Evidence}}
        <div style="margin-top: 12px; padding: 10px; background-color: #fff; border: 1px solid #c8e6c9; border-radius: 4px;">
            <strong style="color: #d84315; font-size: 0.9em;">🔍 Log Evidence Found:</strong>
            <ul style="margin-top: 6px; margin-bottom: 0; padding-left: 20px; font-family: 'Courier New', Courier, monospace; font-size: 0.85em; color: #555;">
                {{range .Evidence}}
                <li style="margin-bottom: 4px;">{{.}}</li>
                {{end}}
            </ul>
        </div>
        {{end}}
    </div>
    {{end}}
</div>
{{end}}

	<div class="footer">
		sssd-inspector v{{.AppVersion}} - SUSE Technical Support -Created by Davide M. Puggioni with Gemini Pro - 2026 - Released under the GNU GPL v3.
	</div>
</body>
</html>
`
	t, err := template.New("report").Parse(tpl)
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

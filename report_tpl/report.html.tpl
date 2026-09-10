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

	/* P6: incident cards, TOC and timeline */
	.toc { background: #fff; border: 1px solid #ddd; border-radius: 6px; padding: 12px 16px; max-width: 900px; margin: 16px 0 8px 0; box-shadow: 0 0 6px rgba(0,0,0,0.06); }
	.toc ul { margin: 6px 0 2px 0; padding-left: 20px; font-size: 0.92em; }
	.toc li { margin-bottom: 3px; }
	.toc a { color: #0056b3; text-decoration: none; }
	.toc a:hover { text-decoration: underline; }
	.incident { margin-bottom: 12px; padding: 12px 14px; border-left: 5px solid #0056b3; background: #fff; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); max-width: 900px; }
	.incident.error { border-left-color: #d9534f; }
	.incident.warning { border-left-color: #d39e00; }
	.incident .incident-head { font-weight: bold; font-size: 1.02em; margin-bottom: 4px; }
	.incident .badge { display: inline-block; padding: 2px 8px; border-radius: 10px; color: #fff; font-size: 0.75em; font-weight: bold; margin-right: 6px; vertical-align: middle; }
	.badge-error { background: #d9534f; }
	.badge-warning { background: #d39e00; }
	.badge-info { background: #17a2b8; }
	.incident .incident-meta { font-size: 0.82em; color: #666; margin-top: 6px; }
	.incident ul { margin: 6px 0 2px 0; padding-left: 20px; font-size: 0.9em; }
	.timeline-table { border-collapse: collapse; width: 100%; max-width: 900px; background: #fff; margin-top: 8px; font-size: 0.88em; }
	.timeline-table th, .timeline-table td { padding: 8px 10px; border: 1px solid #ddd; text-align: left; vertical-align: top; }
	.timeline-table th { background: #0056b3; color: #fff; }
	.timeline-table td.ts { white-space: nowrap; font-family: monospace; }
	.timeline-table td.count { text-align: right; font-weight: bold; white-space: nowrap; }
	.timeline-note { font-size: 0.85em; color: #555; margin-top: 6px; }
	.sev-chip { display: inline-block; padding: 2px 8px; border-radius: 10px; color: #fff; font-size: 0.75em; font-weight: bold; }
	.sev-chip.critical { background: #b71c1c; }
	.sev-chip.error { background: #d9534f; }
	.sev-chip.warning { background: #d39e00; }

	/* Executive summary (triage) block */
	.exec-summary { display: flex; flex-wrap: wrap; gap: 15px; align-items: stretch; margin: 20px 0; max-width: 900px; }
	.exec-score { flex: 0 0 180px; background: #fff; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1); padding: 15px; text-align: center; }
	.exec-score .score { font-size: 2.6em; font-weight: bold; line-height: 1.1; }
	.exec-score .score-label { font-size: 0.8em; color: #666; text-transform: uppercase; letter-spacing: 1px; }
	.exec-score .score-bar { height: 8px; border-radius: 4px; background: #e9ecef; margin-top: 8px; overflow: hidden; }
	.exec-score .score-fill { height: 100%; border-radius: 4px; }
	.score-fill.success { background: green; }
	.score-fill.warn { background: #d39e00; }
	.score-fill.fail { background: #d9534f; }
	.exec-headline { flex: 1 1 300px; background: #fff; border-radius: 8px; box-shadow: 0 0 10px rgba(0,0,0,0.1); padding: 15px; }
	.exec-headline .headline { font-weight: bold; margin-bottom: 10px; }
	.exec-counts { display: flex; gap: 10px; flex-wrap: wrap; }
	.exec-count { padding: 6px 12px; border-radius: 6px; color: #fff; font-weight: bold; font-size: 0.9em; }
	.count-critical { background: #b71c1c; }
	.count-error { background: #d9534f; }
	.count-warning { background: #d39e00; }
	.count-info { background: #17a2b8; }

	/* Findings with provenance drill-down */
	.finding { margin-bottom: 12px; padding: 10px 12px; border-left: 5px solid #888; background: #fff; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.08); }
	.finding.critical { border-left-color: #b71c1c; }
	.finding.error { border-left-color: #d9534f; }
	.finding.warning { border-left-color: #d39e00; }
	.finding .finding-meta { font-family: monospace; font-size: 0.8em; color: #666; margin-top: 6px; }

	/* Print/PDF: clean pagination and no shadows. P6: keep collapsed
		content visible on paper and avoid splitting cards/tables. */
	@media print {
		body { margin: 10mm; background: #fff; }
		.info-table, .exec-score, .exec-headline, .finding, .incident, .toc { box-shadow: none; }
		h1 { font-size: 1.5em; }
		h2 { page-break-after: avoid; }
		table, .exec-summary, .finding, .incident, .log-block, details { page-break-inside: avoid; }
		.log-block { white-space: pre-wrap; }
		a { text-decoration: none; color: inherit; }
		details:not([open]) > *:not(summary) { display: block !important; }
		.toc { display: none; }
	}
</style>
</head>
<body>
	<h1>Supportconfig SSSD Analysis Report</h1>
	<div class="timestamp">Generated on: {{.Timestamp}}</div>

	{{$rows := rootBreakdown .}}
	<nav class="toc" aria-label="Report contents">
		<strong>Contents</strong>
		<ul>
			<li><a href="#triage">Triage</a></li>
			{{if $rows}}<li><a href="#incidents">Root-Cause Incidents ({{len $rows}})</a></li>{{end}}
			<li><a href="#findings">Findings &amp; Problems</a></li>
			{{if .Timeline}}<li><a href="#timeline">Event Timeline ({{timelineTotal .Timeline}} occurrence(s), {{len .Timeline}} shown)</a></li>{{end}}
			{{if .Timeline}}
	<h2 id="timeline">Event Timeline</h2>
	<div style="margin-bottom: 20px; max-width: 900px;">
		<div class="timeline-note">{{timelineTotal .Timeline}} total occurrence(s) across {{len .Timeline}} distinct event row(s). Repeats of the same diagnostic event are collapsed into one row with a count (P7 bounded timeline); showing the first {{timelineShown .Timeline}}.</div>
		<table class="timeline-table">
			<tr><th>Timestamp</th><th>Event</th><th>Occurrences</th></tr>
			{{range timelinePreview .Timeline}}
			<tr><td class="ts">{{.Timestamp}}</td><td>{{.Message}}{{if .RawLog}}<div class="timeline-note">{{.RawLog}}</div>{{end}}</td><td class="count">{{occurrenceCount .}}</td></tr>
			{{end}}
		</table>
	</div>
	{{end}}

	{{if .TemporalClusters}}<li><a href="#clusters">Temporal Clusters</a></li>{{end}}
			{{if or .MatchedTIDs .KBSuggestions}}<li><a href="#kb">Knowledge Base</a></li>{{end}}
		</ul>
	</nav>

	<h2 id="triage" style="color:#0056b3; border-bottom-color:#0056b3;">Triage</h2>
	<div class="exec-summary">
		<div class="exec-score">
			<div class="score-label">Health Score</div>
			<div class="score {{healthScoreClass .Summary.HealthScore}}">{{.Summary.HealthScore}}</div>
			<div class="score-bar"><div class="score-fill {{healthScoreClass .Summary.HealthScore}}" style="width: {{.Summary.HealthScore}}%;"></div></div>
		</div>
		<div class="exec-headline">
			<div class="headline">{{.Summary.Headline}}</div>
			<div class="exec-counts">
				<span class="exec-count count-critical">Critical: {{.Summary.CriticalCount}}</span>
				<span class="exec-count count-error">Errors: {{.Summary.ErrorCount}}</span>
				<span class="exec-count count-warning">Warnings: {{.Summary.WarningCount}}</span>
				<span class="exec-count count-info">Log Patterns: {{.Summary.LogErrorCount}}</span>
			</div>
		</div>
	</div>

	{{$rows := rootBreakdown .}}
	{{if $rows}}
	<h2 id="incidents">Root-Cause Incidents</h2>
	<div style="margin-bottom: 20px; max-width: 900px;">
		<p style="font-size:0.9em; color:#555; margin-top:2px;">One incident card per likely root cause, ranked by weighted strength (config + log signals). This is the "one root cause, many symptoms" view.</p>
		{{range $rows}}
		<div class="incident {{if lt .Severity 2}}warning{{else}}error{{end}}">
			<div class="incident-head"><span class="badge {{if lt .Severity 2}}badge-warning{{else}}badge-error{{end}}">{{if lt .Severity 2}}HINT{{else}}ROOT CAUSE{{end}}</span>[{{.Category}}] - {{len .ConfigSignals}} config signal(s), {{len .LogSignals}} log signal(s)</div>
			<div class="incident-meta">Weighted strength: {{.Severity}}</div>
			<details open>
				<summary>Evidence</summary>
				<div style="margin-top:8px; padding-left:6px;">
					{{if .ConfigSignals}}
					<div style="font-weight:bold; font-size:0.85em; color:#555;">Config / correlation signals</div>
					<ul>
						{{range .ConfigSignals}}<li style="margin-bottom:4px;">{{.}}</li>{{end}}
					</ul>
					{{end}}
					{{if .LogSignals}}
					<div style="font-weight:bold; font-size:0.85em; color:#555;">Log signals</div>
					<ul>
						{{range .LogSignals}}<li style="margin-bottom:4px;">{{.}}</li>{{end}}
					</ul>
					{{end}}
				</div>
			</details>
		</div>
		{{end}}
	</div>
	<h2>Root-Cause Breakdown</h2>
	<div style="margin-bottom: 20px; max-width: 900px;">
		<p style="font-size:0.9em; color:#555; margin-top:2px;">Signals grouped under the coarse root cause they point to (config + log), ranked by weighted strength.</p>
		{{range $rows}}
		<details class="finding {{if lt .Severity 2}}warning{{else}}error{{end}}">
			<summary>[{{.Category}}] - {{len .ConfigSignals}} config signal(s), {{len .LogSignals}} log signal(s)</summary>
			<div style="margin-top:8px; padding-left:6px;">
				{{if .ConfigSignals}}
				<div style="font-weight:bold; font-size:0.85em; color:#555;">Config / correlation signals</div>
				<ul style="margin:4px 0 10px 0; padding-left:20px;">
					{{range .ConfigSignals}}<li style="margin-bottom:4px;">{{.}}</li>{{end}}
				</ul>
				{{end}}
				{{if .LogSignals}}
				<div style="font-weight:bold; font-size:0.85em; color:#555;">Log signals</div>
				<ul style="margin:4px 0 0 0; padding-left:20px;">
					{{range .LogSignals}}<li style="margin-bottom:4px;">{{.}}</li>{{end}}
				</ul>
				{{end}}
			</div>
		</details>
		{{end}}
	</div>
	{{end}}

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

	<h2 id="findings">Actionable Problems Found</h2>
	{{if .ConfigFindings}}
		<h2 class="warn-header">Configuration Findings (with provenance)</h2>
		{{range .ConfigFindings}}
			<div class="finding {{sevClass .Severity}}">
				<div class="headline">{{.Message}}</div>
				<div class="finding-meta">Source: {{.SourcePath}} | Key: {{.SourceKey}} | Line: {{if gt .SourceLine 0}}{{.SourceLine}}{{else}}n/a{{end}}{{if .Evidence}}<br/>Evidence: <code>{{.Evidence}}</code>{{end}}</div>
			</div>
		{{end}}
	{{end}}
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
<h2 id="kb" style="color: #2e7d32; border-bottom: 2px solid #2e7d32; padding-bottom: 5px;">Knowledge Base Articles (TIDs)</h2>
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

	{{if .TemporalClusters}}
	<h2 id="clusters" class="warn-header">Temporal Clusters (Retry Loops / Flapping)</h2>
	<div style="background-color: #fff8e1; border-left: 5px solid #d39e00; padding: 15px; margin-bottom: 20px;">
		{{range .TemporalClusters}}
		<div style="margin-bottom: 12px;">
			<strong>{{.Description}}</strong>
			<div style="margin-top: 4px; font-size: 0.95em;">{{.EventCount}} occurrences between {{.WindowStart}} and {{.WindowEnd}}</div>
			{{if .SampleRawLog}}<div class="log-block" style="margin-top: 6px;">{{.SampleRawLog}}</div>{{end}}
		</div>
		{{end}}
	</div>
	{{end}}

	{{if .KBSuggestions}}
	<h2 class="kb-header">KB Suggestions (Fuzzy Match)</h2>
	<div style="background-color: #e3f2fd; border-left: 5px solid #1565c0; padding: 15px; margin-bottom: 20px;">
		<p style="margin-top:0; font-size:0.9em; color:#555;">Log lines that no known pattern matched, correlated with similar Knowledge Base articles (TF-IDF similarity):</p>
		{{range .KBSuggestions}}
		<div style="margin-bottom: 12px;">
			<strong><a href="{{.URL}}" target="_blank" style="color: #1565c0;">[{{.TIDID}}] {{.Title}}</a></strong>
			<span class="warn"> (similarity {{printf "%.0f" (mult .Score 100)}}%)</span>
			{{if .SampleLine}}<div class="log-block" style="margin-top: 6px;">{{.SampleLine}}</div>{{end}}
		</div>
		{{end}}
	</div>
	{{end}}

	<div class="footer">
		sssd-inspector v{{.AppVersion}} - SUSE Technical Support -Created by Davide M. Puggioni with Gemini Pro - 2026 - Released under the GNU GPL v3.
	</div>

	<script>
	// Vanilla-JS report interactivity (no external runtime): severity filter,
	// live text search over the DOM. The report is fully usable and printable
	// without JavaScript.
	(function () {
		var activeSev = "all", searchText = "";
		var sevBtns = {};

		function applyFilters() {
			var items = document.querySelectorAll(".finding");
			Array.prototype.forEach.call(items, function (it) {
				var cls = it.className || "";
				var sevOk = activeSev === "all" || cls.indexOf(activeSev) !== -1;
				var txt = (it.textContent || "").toLowerCase();
				var searchOk = searchText === "" || txt.indexOf(searchText) !== -1;
				it.style.display = (sevOk && searchOk) ? "" : "none";
			});
		}

		// Toolbar with a search box and severity filter buttons.
		var bar = document.createElement("div");
		bar.style.cssText = "max-width:900px; margin:12px 0; padding:10px; background:#fff; border-radius:6px; box-shadow:0 0 6px rgba(0,0,0,0.08); font-size:0.9em; display:flex; align-items:center; gap:8px; flex-wrap:wrap;";
		var inp = document.createElement("input");
		inp.type = "text";
		inp.placeholder = "Search findings & root causes...";
		inp.style.cssText = "flex:1 1 220px; min-width:200px; padding:6px; border:1px solid #ccc; border-radius:4px;";
		inp.addEventListener("input", function () {
			searchText = (inp.value || "").toLowerCase();
			applyFilters();
		});
		bar.appendChild(inp);

		var label = document.createElement("span");
		label.textContent = "Severity:";
		label.style.color = "#333";
		bar.appendChild(label);
		["all", "critical", "error", "warning"].forEach(function (sev) {
			var b = document.createElement("button");
			b.type = "button";
			b.textContent = sev.charAt(0).toUpperCase() + sev.slice(1);
			b.style.cssText = "padding:5px 10px; border:1px solid #aaa; border-radius:4px; cursor:pointer; background:#e9ecef; color:#000;";
			sevBtns[sev] = b;
			b.addEventListener("click", function () {
				activeSev = sev;
				for (var k in sevBtns) {
					var on = (k === sev);
					sevBtns[k].style.background = on ? "#0056b3" : "#e9ecef";
					sevBtns[k].style.color = on ? "#fff" : "#000";
				}
				applyFilters();
			});
			bar.appendChild(b);
		});
		document.body.insertBefore(bar, document.body.firstChild);
		applyFilters();
	})();
</script>
</body>
</html>

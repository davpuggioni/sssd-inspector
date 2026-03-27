import './style.css';
import './app.css';

// Import Go bridge functions
import { Analyze, OpenFileBrowser, SaveTXT } from '../wailsjs/go/main/App';
import { OnFileDrop, EventsOn } from '../wailsjs/runtime/runtime';

// 2. Build the main layout (Added the Export PDF button)
document.querySelector('#app').innerHTML = `
    <style>
        .report-wrapper { font-family: Arial, sans-serif; background-color: #f4f4f9; color: #333; padding: 20px; text-align: left; }
        h1 { color: #0056b3; border-bottom: 2px solid #0056b3; padding-bottom: 10px; margin-bottom: 5px; font-size: 1.8em; }
        h2 { color: #d9534f; border-bottom: 1px solid #d9534f; padding-bottom: 5px; margin-top: 30px; font-size: 1.4em;}
        h2.warn-header { color: #17a2b8; border-bottom: 1px solid #17a2b8; }
        .info-table { border-collapse: collapse; width: 100%; background: #fff; box-shadow: 0 0 10px rgba(0,0,0,0.1); margin-bottom: 20px;}
        .info-table th, .info-table td { padding: 10px 15px; border: 1px solid #ddd; text-align: left; font-size: 0.95em; }
        .info-table th { background-color: #0056b3; color: white; width: 35%; }
        .section-title { background-color: #e9ecef !important; color: #333 !important; font-weight: bold; text-align: center; }
        .problem-list { background: #ffebee; padding: 15px 15px 15px 35px; border-left: 5px solid #d9534f; list-style-type: square; }
        .problem-list li { margin-bottom: 8px; font-weight: bold; color: #b71c1c; }
        .warn-list { background: #e2f3f5; padding: 15px 15px 15px 35px; border-left: 5px solid #17a2b8; list-style-type: square; }
        .warn-list li { margin-bottom: 8px; color: #0c5460; }
        .success-text { color: green; font-weight: bold; font-size: 1.1em; }
        .success { color: green; font-weight: bold; }
        .fail { color: red; font-weight: bold; }
        .warn { color: #d39e00; font-weight: bold; }
        .log-block { margin-top: 5px; background: #f8f9fa; padding: 10px; border-left: 3px solid #d9534f; font-family: monospace; font-size: 0.85em; overflow-x: auto; color: #333; }
        .mac-block { background: #fff3cd; border-left: 3px solid #d39e00; }
        details summary { cursor: pointer; font-weight: bold; color: #555; padding: 5px 0; outline: none; transition: color 0.2s; }
        details summary:hover { color: #000; }
        @media print {
            #topBar { display: none !important; }
            body, #app, #pdfContentArea, .report-wrapper { background: #fff !important; margin: 0 !important; padding: 0 !important; }
            .report-wrapper { box-shadow: none !important; }
            a { text-decoration: none; color: black; }
        }
    </style>

    <div id="topBar" style="padding: 20px; background-color: #1e1e2e; color: white; text-align: center; border-bottom: 4px solid #0056b3;">
        <h2 style="margin: 0 0 10px 0; border: none; color: #fff;">SSSD Supportconfig Analyzer</h2>
        <div style="display: flex; justify-content: center; gap: 10px;">
            <input id="filePath" type="text" placeholder="Select or paste path to supportconfig.txz..." style="width: 500px; padding: 10px; font-size: 14px; border-radius: 4px; border: 1px solid #ccc;"/>
            <button id="browseBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 4px; font-weight: bold;">Browse...</button>
            <button id="analyzeBtn" style="padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #0056b3; color: white; border: none; border-radius: 4px; font-weight: bold;">Analyze</button>
            <button id="exportPdfBtn" style="display: none; padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #28a745; color: white; border: none; border-radius: 4px; font-weight: bold;">📄 Export PDF</button>
            <button id="exportTxtBtn" style="display: none; padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #17a2b8; color: white; border: none; border-radius: 4px; font-weight: bold;">📝 Export TXT</button>
            <div id="zoomControls" style="display: none; margin-left: 10px;">
                <button id="zoomOutBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 4px 0 0 4px; border-right: 1px solid #5a6268; font-weight: bold;">A-</button>
                <button id="zoomInBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 0 4px 4px 0; font-weight: bold;">A+</button>
            </div>
        </div>
    </div>

    <div id="pdfContentArea">
        <div id="resultBox" class="report-wrapper" style="min-height: 80vh;">
            <div style="text-align: center; color: #777; margin-top: 50px;">Waiting for supportconfig file...</div>
        </div>
    </div>
`;

// Formatting Helpers
const yesNo = (bool) => bool ? '<span class="success">Yes</span>' : '<span class="fail">No</span>';
const listItems = (arr) => arr && arr.length > 0 ? arr.map(i => `<li>${i}</li>`).join('') : '';

// The Render Function
function renderReportHTML(report) {
    let html = `<h1>Analysis Report</h1>`;

    html += `
    <div style="margin-bottom: 20px; font-size: 1.1em; color: #555;">
        <strong>Analysis Date:</strong> ${report.timestamp}<br/>
        ${report.support_case_id ? `<strong>Support Case (SR#):</strong> ${report.support_case_id}` : ''}
    </div>
    <table class="info-table">
        <tr><td colspan="2" class="section-title">System & Virtualization Information</td></tr>
        <tr><th>OS Release</th><td>${report.sles_release || 'Unknown'}</td></tr>
        <tr><th>Kernel Version</th><td>${report.kernel_version || 'Unknown'}</td></tr>
        <tr><th>SCC Status</th><td>${report.scc_status || 'Unknown'}</td></tr>
        <tr><th>Hardware</th><td>${report.hardware_manufacturer || ''} ${report.hardware_model || ''}</td></tr>
        <tr><th>Virtualization</th><td>${report.hypervisor || 'Unknown'} (Identity: ${report.virtual_identity || 'Unknown'})</td></tr>
        <tr><th>MAC Security</th><td>${report.mac_type || 'Unknown/None'}</td></tr>

        <tr><td colspan="2" class="section-title">Base Authentication Services</td></tr>
        <tr><th>SSSD Installed</th><td>${yesNo(report.sssd_installed)}</td></tr>
        <tr><th>SSSD Packages</th><td>
            ${report.sssd_packages && report.sssd_packages.length > 0 ? 
                `<pre style="margin: 0; font-family: monospace; font-size: 0.9em;">${report.sssd_packages.join('\n')}</pre>` : 
                'None Detected'}
        </td></tr>
        <tr><th>SSSD Config Found</th><td>${yesNo(report.sssd_config_found)}</td></tr>
        <tr><th>SSSD Service Status</th><td>${report.sssd_service || 'Unknown'}</td></tr>
        <tr><th>Winbind Status</th><td>${report.winbind_service || 'Unknown'}</td></tr>
        <tr><th>NSCD Service Status</th><td>${report.nscd_status || 'Unknown'}</td></tr>
        <tr><th>NSSwitch Valid</th><td>${yesNo(report.nsswitch_valid)}</td></tr>
        <tr><th>PAM pam_sss.so Found</th><td>${report.pam_gdpr_restricted ? '<span class="warn">Restricted (GDPR)</span>' : yesNo(report.pam_sss_installed)}</td></tr>

        <tr><td colspan="2" class="section-title">Active Directory & Kerberos Prerequisites</td></tr>
        <tr><th>DNS Nameservers</th><td><ul style="margin: 0; padding-left: 20px;">${listItems(report.nameservers)}</ul></td></tr>
        <tr><th>DNS Search Domain</th><td>${report.search_domain || 'None'}</td></tr>
        <tr><th>Time Sync Service</th><td>${report.time_service || 'Unknown'}</td></tr>
        <tr><th>Kerberos Realm</th><td>${report.kerberos_realm || 'Not configured'}</td></tr>
        <tr><th>Machine Keytab Verified</th><td>${yesNo(report.keytab_found)}</td></tr>
        
        <tr><td colspan="2" class="section-title">Deep SSSD Configuration</td></tr>
        <tr><th>AD Provider Explicitly Set</th><td>${yesNo(report.ad_provider_mode)}</td></tr>
        <tr><th>Use Fully Qualified Names</th><td>${yesNo(report.use_fqdn_set)}</td></tr>
        <tr><th>Enumerate Set to True</th><td>${report.enumerate_issue ? '<span class="fail">Yes (Performance Risk)</span>' : '<span class="success">No</span>'}</td></tr>

        <tr><td colspan="2" class="section-title">Configuration Files</td></tr>
        <tr><td colspan="2">
            ${report.sssd_config_snippet ? `
            <details>
                <summary style="cursor: pointer; font-weight: bold; padding: 5px 0;">View sssd.conf</summary>
                <pre class="log-block" style="white-space: pre-wrap; margin-top: 10px;">${report.sssd_config_snippet}</pre>
            </details>` : 'sssd.conf not found / not readable'}
        </td></tr>
    </table>`;

    html += `<h2>Actionable Problems Found</h2>`;
    if (report.problems && report.problems.length > 0) {
        html += `<ul class="problem-list">${listItems(report.problems)}</ul>`;
    } else {
        html += `<p class="success-text">No major SSSD/Auth issues detected based on the analysis!</p>`;
    }

    if (report.sssd_log_errors && report.sssd_log_errors.length > 0) {
        html += `<h2>SSSD Log Errors (sssd.txt)</h2><table class="info-table">`;
        report.sssd_log_errors.forEach(err => {
            html += `
            <tr><td colspan="2"><span class="fail">Error Detected:</span> ${err.description}
                <details open>
                    <summary>View Log Snippets</summary>
                    <div class="log-block">${err.examples.map(e => `<div style="margin-bottom: 4px;">${e}</div>`).join('')}</div>
                </details>
            </td></tr>`;
        });
        html += `</table>`;
    }

    if (report.mac_denial_examples && report.mac_denial_examples.length > 0) {
        html += `<h2 style="color: #d39e00; border-bottom: 1px solid #d39e00;">AppArmor / SELinux Denials Found</h2>
        <details open>
            <summary style="color: #d39e00;">View Blocked Access Logs</summary>
            <div class="log-block mac-block">${report.mac_denial_examples.map(e => `<div style="margin-bottom: 4px;">${e}</div>`).join('')}</div>
        </details>`;
    }

    if (report.warnings && report.warnings.length > 0) {
        html += `<h2 class="warn-header">Tuning & Diagnostic Hints</h2>
        <ul class="warn-list">${listItems(report.warnings)}</ul>`;
    }

    if (report.matched_tids && report.matched_tids.length > 0) {
        html += `<h2 style="color: #2e7d32; border-bottom: 2px solid #2e7d32;">Knowledge Base Articles (TIDs)</h2>
        <div style="background-color: #e8f5e9; border-left: 5px solid #2e7d32; padding: 15px; margin-bottom: 20px;">`;
        
        report.matched_tids.forEach(tid => {
            html += `<div style="margin-bottom: 20px;">
                <strong><a href="${tid.url}" target="_blank" style="color: #1565c0; text-decoration: none; font-size: 1.1em;">[${tid.tid_id}] ${tid.title}</a></strong>
                <div style="white-space: pre-wrap; font-family: monospace; margin-top: 8px; font-size: 0.95em; color: #333;">${tid.description}</div>`;
            
            if (tid.Evidence && tid.Evidence.length > 0) {
                html += `
                <div style="margin-top: 12px; padding: 10px; background-color: #fff; border: 1px solid #c8e6c9; border-radius: 4px;">
                    <strong style="color: #d84315; font-size: 0.9em;">🔍 Log Evidence Found:</strong>
                    <ul style="margin-top: 6px; padding-left: 20px; font-family: monospace; font-size: 0.85em; color: #555;">
                        ${tid.Evidence.map(e => `<li>${e}</li>`).join('')}
                    </ul>
                </div>`;
            }
            html += `</div>`;
        });
        html += `</div>`;
    }

    return html;
}

// Wire up the Browse Button
document.getElementById('browseBtn').addEventListener('click', () => {
    OpenFileBrowser().then((selectedPath) => {
        if (selectedPath) {
            document.getElementById('filePath').value = selectedPath;
        }
    }).catch((err) => {
        console.error("Failed to open file browser:", err);
    });
});

// Wire up the Analyze Button
document.getElementById('analyzeBtn').addEventListener('click', () => {
    let filePath = document.getElementById('filePath').value.trim();
    let resultBox = document.getElementById('resultBox');
    let exportPdfBtn = document.getElementById('exportPdfBtn');
    let exportTxtBtn = document.getElementById('exportTxtBtn');
    let zoomControls = document.getElementById('zoomControls');

    if (filePath === "") {
        resultBox.innerHTML = "<div style='color: red; text-align: center; margin-top: 20px;'>Please enter a path first or click Browse.</div>";
        return;
    }

    // Hide Export button during loading
    exportPdfBtn.style.display = 'none';
    exportTxtBtn.style.display = 'none';
    zoomControls.style.display = 'none';

    resultBox.innerHTML = `
        <div style="text-align: center; margin-top: 50px;">
            <h3 style="color: #0056b3;">Analyzing...</h3>
            <p style="color: #666;">Parsing supportconfig in Go! 🚀</p>
        </div>
    `;

    Analyze(filePath)
        .then((report) => {
            window.currentReport = report; // Store for TXT export
            resultBox.innerHTML = renderReportHTML(report);
            // Show the Export buttons now that we have data!
            exportPdfBtn.style.display = 'inline-block';
            exportTxtBtn.style.display = 'inline-block';
            zoomControls.style.display = 'inline-flex';
        })
        .catch((err) => {
            resultBox.innerHTML = `
            <div style="background: #ffebee; padding: 20px; border-left: 5px solid #d9534f; color: #b71c1c; margin-top: 20px;">
                <strong>Error during analysis:</strong><br><br>${err}
            </div>`;
        });
});

// 3. Wire up the Export PDF Button (Native Print Dialog)
document.getElementById('exportPdfBtn').addEventListener('click', () => {
    // Generate a timestamp
    const now = new Date();
    const timestamp = now.toISOString().replace(/T/, '_').replace(/:/g, '-').split('.')[0];
    
    // Temporarily change document title so the native Print to PDF picks it up as the filename
    const originalTitle = document.title;
    document.title = `SSSD_Analysis_Report_${timestamp}`;
    
    // Trigger OS native print dialog
    window.print();
    
    // Restore title immediately after the dialog opens
    document.title = originalTitle;
});

// 4. Wire up the Export TXT Button
document.getElementById('exportTxtBtn').addEventListener('click', () => {
    const btn = document.getElementById('exportTxtBtn');
    btn.innerText = "⏳ Saving...";
    
    SaveTXT(window.currentReport).then((savedPath) => {
        if (savedPath === "cancelled") {
            btn.innerText = "📝 Export TXT"; // User clicked cancel
        } else {
            btn.innerText = "✅ Saved!";
            setTimeout(() => { btn.innerText = "📝 Export TXT"; }, 3000);
        }
    }).catch((err) => {
        console.error("Save error:", err);
        btn.innerText = "❌ Error";
        setTimeout(() => { btn.innerText = "📝 Export TXT"; }, 3000);
    });
});

// 5. Wire up the Zoom Buttons
let currentFontSize = 1.0;
document.getElementById('zoomInBtn').addEventListener('click', () => {
    currentFontSize += 0.1;
    document.getElementById('resultBox').style.fontSize = currentFontSize + 'em';
});
document.getElementById('zoomOutBtn').addEventListener('click', () => {
    // Don't let it get completely unreadable
    if (currentFontSize > 0.4) {
        currentFontSize -= 0.1;
        document.getElementById('resultBox').style.fontSize = currentFontSize + 'em';
    }
});

// 6. Wire up Drag and Drop (Comprehensive Fallbacks)
function handleFilePath(path) {
    if (path) {
        document.getElementById('filePath').value = path;
        document.getElementById('analyzeBtn').click();
    }
}

// 6A. Wails Native method
OnFileDrop((x, y, paths) => {
    if (paths && paths.length > 0) handleFilePath(paths[0]);
}, false);

// 6B. Wails Event Bus method (bypass OnFileDrop wrapper)
EventsOn("wails:file-drop", (x, y, paths) => {
    if (paths && paths.length > 0) handleFilePath(paths[0]);
});

// 6C. HTML5 DOM Fallback (Needed for WebKitGTK on some Linux distros)
let dropZone = document.getElementById('app');
dropZone.addEventListener('dragover', (e) => {
    e.preventDefault(); 
});
dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    if (e.dataTransfer && e.dataTransfer.files.length > 0) {
        let file = e.dataTransfer.files[0];
        // Linux Webkit instances typically expose the absolute path here
        if (file.path) {
            handleFilePath(file.path);
        }
    }
});
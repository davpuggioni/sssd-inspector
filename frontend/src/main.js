// Hybrid main.js - preserves original synchronous rendering, imports modular CSS
import './styles/layout.css';
import './styles/components.css';
import './styles/report.css';

import { Analyze, OpenFileBrowser, SaveTXT, SavePDF } from '../wailsjs/go/main/App';
import { OnFileDrop, EventsOn } from '../wailsjs/runtime/runtime';

const yesNo = (bool) => bool ? '<span class="success">Yes</span>' : '<span class="fail">No</span>';
const listItems = (arr) => arr && arr.length > 0 ? arr.map(i => `<li>${i}</li>`).join('') : '';

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
        <tr><th>SSSD Packages</th><td>${report.sssd_packages && report.sssd_packages.length > 0 ? `<pre style="margin:0;font-family:monospace;font-size:0.9em;">${report.sssd_packages.join('\n')}</pre>` : 'None Detected'}</td></tr>
        <tr><th>SSSD Config Found</th><td>${yesNo(report.sssd_config_found)}</td></tr>
        <tr><th>SSSD Service Status</th><td>${report.sssd_service || 'Unknown'}</td></tr>
        <tr><th>Winbind Service</th><td>${report.winbind_service || 'Unknown'}</td></tr>
        <tr><th>nscd Status</th><td>${report.nscd_status || 'Unknown'}</td></tr>
        <tr><th>nscd Caching</th><td>${listItems(report.nscd_caching)}</td></tr>
        <tr><td colspan="2" class="section-title">Authentication Configuration</td></tr>
        <tr><th>nsswitch.conf Valid</th><td>${yesNo(report.nsswitch_valid)}</td></tr>
        <tr><th>pam_sss Installed</th><td>${yesNo(report.pam_sss_installed)}</td></tr>
        <tr><th>GDPR Restricted Mode</th><td>${yesNo(report.pam_gdpr_restricted)}</td></tr>
        <tr><th>Hosts File Issues</th><td>${listItems(report.hosts_issues)}</td></tr>
        <tr><th>Nameservers</th><td>${listItems(report.nameservers)}</td></tr>
        <tr><th>Search Domain</th><td>${report.search_domain || 'None'}</td></tr>
        <tr><th>Time Service</th><td>${report.time_service || 'Unknown'}</td></tr>
        <tr><th>Kerberos Realm</th><td>${report.kerberos_realm || 'Not configured'}</td></tr>
        <tr><th>Keytab Found</th><td>${yesNo(report.keytab_found)}</td></tr>
        <tr><td colspan="2" class="section-title">Advanced SSSD Configuration</td></tr>
        <tr><th>AD Provider Mode</th><td>${yesNo(report.ad_provider_mode)}</td></tr>
        <tr><th>Enumerate Issue</th><td>${yesNo(report.enumerate_issue)}</td></tr>
        <tr><th>Use FQDN Set</th><td>${yesNo(report.use_fqdn_set)}</td></tr>
        
        <tr><td colspan="2" class="section-title">Configuration Files</td></tr>
        <tr><td colspan="2">
            ${report.sssd_config_snippet ? `
            <details>
                <summary style="cursor: pointer; font-weight: bold; padding: 5px 0;">View sssd.conf</summary>
                <pre class="log-block" style="white-space: pre-wrap; margin-top: 10px;">${report.sssd_config_snippet}</pre>
            </details>` : 'sssd.conf not found / not readable'}
        </td></tr>
    </table>`;

    if (report.problems && report.problems.length > 0) {
        html += `<h2>Critical Problems Detected</h2><ul class="problem-list">${listItems(report.problems)}</ul>`;
    }
    if (report.warnings && report.warnings.length > 0) {
        html += `<h2 class="warn-header">Warnings & Recommendations</h2><ul class="warn-list">${listItems(report.warnings)}</ul>`;
    }
    if (report.sssd_log_errors && report.sssd_log_errors.length > 0) {
        html += `<h2>SSSD Log Errors</h2>`;
        report.sssd_log_errors.forEach(error => {
            html += `<div style="margin-bottom: 15px;"><h3 style="color: #d9534f; font-size: 1.1em; margin-bottom: 8px;">${error.description}</h3>${error.examples && error.examples.length > 0 ? `<div class="log-block">${error.examples.map(ex => ex.replace(/</g, '&lt;').replace(/>/g, '&gt;')).join('<br>')}</div>` : ''}</div>`;
        });
    }
    if (report.mac_denial_examples && report.mac_denial_examples.length > 0) {
        html += `<h2>MAC Security Denials</h2><div class="mac-block">${report.mac_denial_examples.map(ex => ex.replace(/</g, '&lt;').replace(/>/g, '&gt;')).join('<br>')}</div>`;
    }
    if (report.timeline && report.timeline.length > 0) {
        html += `<h2>Event Timeline</h2><div class="timeline">`;
        report.timeline.forEach(event => {
            html += `<div class="timeline-event"><div class="timeline-time">${event.timestamp}</div><div class="timeline-msg">${event.message}</div>${event.raw_log ? `<div class="timeline-raw">${event.raw_log.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</div>` : ''}</div>`;
        });
        html += `</div>`;
    }
    if (report.matched_tids && report.matched_tids.length > 0) {
        html += `<h2>Relevant Knowledge Base Articles</h2>`;
        report.matched_tids.forEach(tid => {
            html += `<div style="margin-bottom: 20px; padding: 15px; border: 1px solid #ddd; border-radius: 4px; background: #f9f9f9;"><h3 style="margin-top: 0; color: #0056b3;"><a href="${tid.url}" target="_blank" style="text-decoration: none; color: inherit;">${tid.title}</a></h3><p style="margin-bottom: 10px; color: #666; font-size: 0.9em;">TID: ${tid.tid_id}</p><p>${tid.description}</p>${tid.evidence && tid.evidence.length > 0 ? `<details><summary>Evidence Found</summary><div class="log-block">${tid.evidence.map(ex => ex.replace(/</g, '&lt;').replace(/>/g, '&gt;')).join('<br>')}</div></details>` : ''}</div>`;
        });
    }
    return html;
}

let currentReport = null;
let currentZoom = 1.0;

document.querySelector('#app').innerHTML = `
    <div id="topBar" style="padding: 20px; background-color: #1e1e2e; color: white; text-align: center; border-bottom: 4px solid #0056b3;">
        <h2 style="margin: 0 0 10px 0; border: none; color: #fff;">SSSD Supportconfig Analyzer</h2>
        <div style="display: flex; justify-content: center; align-items: center; gap: 10px;">
            <input id="filePath" type="text" placeholder="Select or paste path to supportconfig.txz..." style="width: 450px; padding: 10px; font-size: 14px; border-radius: 4px; border: 1px solid #ccc;"/>
            <button id="browseBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 4px; font-weight: bold;">Browse...</button>
            <label style="color: white; margin-left: 5px; margin-right: 5px; display: flex; align-items: center; font-size: 14px; cursor: pointer;" title="Redacts IP Addresses and Domain Names from the report">
                <input type="checkbox" id="anonymizeCheck" style="margin-right: 5px; cursor: pointer; transform: scale(1.2);"> Anonymize PII
            </label>
            <button id="analyzeBtn" style="padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #0056b3; color: white; border: none; border-radius: 4px; font-weight: bold;">Analyze</button>
            <button id="exportPdfBtn" style="display: none; padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #28a745; color: white; border: none; border-radius: 4px; font-weight: bold;">Export PDF</button>
            <button id="exportTxtBtn" style="display: none; padding: 10px 20px; font-size: 14px; cursor: pointer; background-color: #17a2b8; color: white; border: none; border-radius: 4px; font-weight: bold;">Export TXT</button>
            <div id="zoomControls" style="display: none; margin-left: 5px;">
                <button id="zoomOutBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 4px 0 0 4px; border-right: 1px solid #5a6268; font-weight: bold;">A-</button>
                <button id="zoomInBtn" style="padding: 10px 15px; font-size: 14px; cursor: pointer; background-color: #6c757d; color: white; border: none; border-radius: 0 4px 4px 0; font-weight: bold;">A+</button>
            </div>
        </div>
    </div>
    <!-- Progress bar container (hidden by default) -->
    <div id="progressContainer" style="display: none; padding: 20px 40px; background-color: #f8f9fa; border-bottom: 1px solid #dee2e6;">
        <div class="progress-container">
            <div class="progress-status" id="progressStatus">Initializing analysis...</div>
            <div class="progress-bar-wrapper">
                <div class="progress-bar-fill" id="progressFill" style="width: 0%;"></div>
            </div>
            <div class="progress-percentage" id="progressPercentage">0%</div>
        </div>
    </div>
    <div id="pdfContentArea">
        <div id="resultBox" class="report-wrapper" style="min-height: 80vh;">
            <div style="text-align: center; color: #777; margin-top: 50px;">Waiting for supportconfig file...</div>
        </div>
    </div>
`;

const filePathInput = document.querySelector('#filePath');
const browseBtn = document.querySelector('#browseBtn');
const analyzeBtn = document.querySelector('#analyzeBtn');
const anonymizeCheck = document.querySelector('#anonymizeCheck');
const exportPdfBtn = document.querySelector('#exportPdfBtn');
const exportTxtBtn = document.querySelector('#exportTxtBtn');
const zoomInBtn = document.querySelector('#zoomInBtn');
const zoomOutBtn = document.querySelector('#zoomOutBtn');
const resultBox = document.querySelector('#resultBox');

browseBtn.addEventListener('click', async () => {
    try {
        const path = await OpenFileBrowser();
        if (path) filePathInput.value = path;
    } catch (error) {
        alert('Failed to open file browser: ' + error);
    }
});

analyzeBtn.addEventListener('click', async () => {
    const filePath = filePathInput.value.trim();
    if (!filePath) {
        alert('Please select a supportconfig file to analyze');
        return;
    }
    analyzeBtn.disabled = true;
    analyzeBtn.textContent = 'Analyzing...';
    browseBtn.disabled = true;
    filePathInput.disabled = true;

    // Show progress bar and reset to 0%
    const progressContainer = document.querySelector('#progressContainer');
    const progressFill = document.querySelector('#progressFill');
    const progressStatus = document.querySelector('#progressStatus');
    const progressPercentage = document.querySelector('#progressPercentage');
    progressContainer.style.display = 'block';
    progressFill.style.width = '0%';
    progressStatus.textContent = 'Starting analysis...';
    progressPercentage.textContent = '0%';

    try {
        const report = await Analyze(filePath, anonymizeCheck.checked);
        currentReport = report;
        resultBox.innerHTML = renderReportHTML(report);
        currentZoom = 1.0;
        applyZoom();
        exportPdfBtn.style.display = 'inline-block';
        exportTxtBtn.style.display = 'inline-block';
        zoomControls.style.display = 'inline-block';
    } catch (error) {
        alert('Analysis failed: ' + error);
        resultBox.innerHTML = `<div style="text-align: center; color: #d9534f; margin-top: 50px;">Analysis failed: ${error}</div>`;
    } finally {
        // Hide progress bar when done
        progressContainer.style.display = 'none';
        analyzeBtn.disabled = false;
        analyzeBtn.textContent = 'Analyze';
        browseBtn.disabled = false;
        filePathInput.disabled = false;
    }
});

exportPdfBtn.addEventListener('click', () => {
    if (!currentReport) return;
    const now = new Date();
    const timestamp = now.toISOString().replace(/T/, '_').replace(/:/g, '-').split('.')[0];
    const originalTitle = document.title;
    document.title = `SSSD_Analysis_Report_${timestamp}`;
    window.print();
    document.title = originalTitle;
});

exportTxtBtn.addEventListener('click', async () => {
    if (!currentReport) return;
    try {
        const result = await SaveTXT(currentReport);
        if (result !== 'cancelled') alert('Text report saved to: ' + result);
    } catch (error) {
        alert('Failed to export text: ' + error);
    }
});

zoomInBtn.addEventListener('click', () => {
    if (currentZoom < 1.5) { currentZoom += 0.1; applyZoom(); }
});

zoomOutBtn.addEventListener('click', () => {
    if (currentZoom > 0.8) { currentZoom -= 0.1; applyZoom(); }
});

function applyZoom() {
    resultBox.style.transform = `scale(${currentZoom})`;
    resultBox.style.transformOrigin = 'top left';
}

EventsOn('analyze-progress', (message, percentage) => {
    // Update progress bar elements
    const progressFill = document.querySelector('#progressFill');
    const progressStatus = document.querySelector('#progressStatus');
    const progressPercentage = document.querySelector('#progressPercentage');
    
    if (progressFill && progressStatus && progressPercentage) {
        progressFill.style.width = `${percentage}%`;
        progressStatus.textContent = message;
        progressPercentage.textContent = `${percentage}%`;
    }
    
    // Also update button text for redundancy
    if (percentage > 0 && percentage < 100) {
        analyzeBtn.textContent = `${percentage}% - ${message}`;
    } else if (percentage >= 100) {
        analyzeBtn.textContent = 'Complete!';
        progressStatus.textContent = 'Analysis complete!';
        progressPercentage.textContent = '100%';
        progressFill.style.width = '100%';
    }
});

OnFileDrop((x, y, files) => {
    if (files.length > 0) filePathInput.value = files[0];
}, false);

document.addEventListener('keydown', (e) => {
    if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
            case 'o': e.preventDefault(); browseBtn.click(); break;
            case 'Enter': e.preventDefault(); if (filePathInput.value.trim()) analyzeBtn.click(); break;
            case 'p': e.preventDefault(); if (exportPdfBtn.style.display !== 'none') exportPdfBtn.click(); break;
            case 's': e.preventDefault(); if (exportTxtBtn.style.display !== 'none') exportTxtBtn.click(); break;
        }
    }
});

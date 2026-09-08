// SSSD Inspector Frontend - Main Application
// Uses modular UI components, CSS classes, and structured error handling
import './styles/layout.css';
import './styles/components.css';
import './styles/report.css';

import { Analyze, OpenFileBrowser, SaveTXT, SavePDF, SaveJSON } from '../wailsjs/go/main/App';
import { OnFileDrop, EventsOn } from '../wailsjs/runtime/runtime';
import { UI, CSS_CLASSES, COLORS, ZOOM, FILES } from './config/constants.js';
import { FileValidator } from './utils/validators.js';
import { renderCorrelationGraph } from './components/CorrelationGraph.js';

// ==========================================================================
// Helper Functions
// ==========================================================================

const yesNo = (bool) => bool ? '<span class="success">Yes</span>' : '<span class="fail">No</span>';
const listItems = (arr) => arr && arr.length > 0 ? arr.map(i => `<li>${i}</li>`).join('') : '';

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text || '';
    return div.innerHTML;
}

// ==========================================================================
// Report Rendering
// ==========================================================================

function renderReportHTML(report) {
    const s = report.summary || { health_score: 0, critical_count: 0, error_count: 0, warning_count: 0, log_error_count: 0, headline: '' };
    const scoreClass = s.health_score >= 80 ? 'success' : s.health_score >= 50 ? 'warn' : 'fail';
    let html = `<h1>Analysis Report</h1>`;
    html += `
    <div class="exec-summary">
        <div class="exec-score">
            <div class="score-label">Health Score</div>
            <div class="score ${scoreClass}">${s.health_score}</div>
            <div class="score-bar"><div class="score-fill ${scoreClass}" style="width: ${s.health_score}%;"></div></div>
        </div>
        <div class="exec-headline">
            <div class="headline">${escapeHtml(s.headline)}</div>
            <div class="exec-counts">
                <span class="exec-count count-critical">Critical: ${s.critical_count}</span>
                <span class="exec-count count-error">Errors: ${s.error_count}</span>
                <span class="exec-count count-warning">Warnings: ${s.warning_count}</span>
                <span class="exec-count count-info">Log Patterns: ${s.log_error_count}</span>
            </div>
        </div>
    </div>
    <div class="report-meta">
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
        <tr><th>SSSD Packages</th><td>${report.sssd_packages && report.sssd_packages.length > 0 ? `<pre class="pkg-list">${report.sssd_packages.join('\n')}</pre>` : 'None Detected'}</td></tr>
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
  // Lightweight INI syntax highlighter for sssd.conf (no external deps).
  // Escapes HTML first, then wraps tokens in spans: sections, comments,
  // keys and values. Line-based so whitespace/indentation is preserved.
  const highlightIni = (text) => {
    const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    const KEY_RE = /^([A-Za-z_][A-Za-z0-9_.-]*)(\s*=\s*)(.*)$/;
    return String(text).split('\n').map((line) => {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith(';')) {
        return `<span class="ini-comment">${esc(line)}</span>`;
      }
      const sec = trimmed.match(/^\[([^\]]*)\]$/);
      if (sec) {
        // Highlight only the section header part; keep surrounding whitespace verbatim.
        const start = line.indexOf('[');
        const end = line.lastIndexOf(']');
        const before = line.slice(0, start);
        const after = line.slice(end + 1);
        return `${esc(before)}<span class="ini-section">[${esc(sec[1])}]</span>${esc(after)}`;
      }
      const kv = line.match(KEY_RE);
      if (kv) {
        return `<span class="ini-key">${esc(kv[1])}</span><span class="ini-op">${esc(kv[2])}</span><span class="ini-value">${esc(kv[3])}</span>`;
      }
      return esc(line);
    }).join('\n');
  };


        <tr><th>Use FQDN Set</th><td>${yesNo(report.use_fqdn_set)}</td></tr>
        
        <tr><td colspan="2" class="section-title">Configuration Files</td></tr>
        <tr><td colspan="2">
            ${report.sssd_config_snippet ? `
            <details open>
                <summary style="cursor: pointer; font-weight: bold; padding: 5px 0;">View sssd.conf</summary>
                <pre class="log-block ini-snippet" style="white-space: pre-wrap; margin-top: 10px;">${highlightIni(report.sssd_config_snippet)}</pre>
            </details>` : 'sssd.conf not found / not readable'}
        </td></tr>
    </table>`;

    // Collapsible section helper: same pattern as the Correlation Graph
    // section — native <details>/<summary>, zero JS needed.
    const collapsible = (title, cls, body, open = true, count = null) => {
      const badge = count !== null ? ` <span class="section-count">${count}</span>` : '';
      return `<details class="report-section"${open ? ' open' : ''}><summary><h2 class="${cls || ''}">${title}${badge}</h2></summary><div class="section-body">${body}</div></details>`;
    };

    if (report.config_findings && report.config_findings.length > 0) {
        let body = '';
        report.config_findings.forEach(f => {
            const cls = f.severity === 2 ? 'critical' : f.severity === 1 ? 'error' : 'warning';
            const line = f.source_line && f.source_line > 0 ? f.source_line : 'n/a';
            body += `<div class="finding ${cls}"><div class="headline">${escapeHtml(f.message)}</div><div class="finding-meta">Source: ${escapeHtml(f.source_path)} | Key: ${escapeHtml(f.source_key)} | Line: ${line}${f.evidence ? `<br/>Evidence: ${escapeHtml(f.evidence)}` : ''}</div></div>`;
        });
        html += collapsible('Configuration Findings (with provenance)', 'warn-header', body, true, report.config_findings.length);
    }
    if (report.problems && report.problems.length > 0) {
        html += collapsible('Critical Problems Detected', '', `<ul class="problem-list">${listItems(report.problems)}</ul>`, true, report.problems.length);
    }
    if (report.warnings && report.warnings.length > 0) {
        html += collapsible('Warnings & Recommendations', 'warn-header', `<ul class="warn-list">${listItems(report.warnings)}</ul>`, true, report.warnings.length);
    }
    if (report.sssd_log_errors && report.sssd_log_errors.length > 0) {
        let body = '';
        report.sssd_log_errors.forEach(error => {
            body += `<div class="error-item"><h3 class="error-title">${escapeHtml(error.description)}</h3>${error.examples && error.examples.length > 0 ? `<div class="log-block">${error.examples.map(ex => escapeHtml(ex)).join('<br>')}</div>` : ''}</div>`;
        });
        html += collapsible('SSSD Log Errors', '', body, true, report.sssd_log_errors.length);
    }
    if (report.mac_denial_examples && report.mac_denial_examples.length > 0) {
        html += collapsible('MAC Security Denials', '', `<div class="mac-block">${report.mac_denial_examples.map(ex => escapeHtml(ex)).join('<br>')}</div>`, true, report.mac_denial_examples.length);
    }
    if (report.timeline && report.timeline.length > 0) {
        let body = '<div class="timeline">';
        report.timeline.forEach(event => {
            body += `<div class="timeline-event"><div class="timeline-time">${event.timestamp}</div><div class="timeline-msg">${escapeHtml(event.message)}</div>${event.raw_log ? `<div class="timeline-raw">${escapeHtml(event.raw_log)}</div>` : ''}</div>`;
        });
        body += '</div>';
        html += collapsible('Event Timeline', '', body, true, report.timeline.length);
    }
    if (report.matched_tids && report.matched_tids.length > 0) {
        let body = '';
        report.matched_tids.forEach(tid => {
            body += `<div class="kb-article"><h4><a href="${escapeHtml(tid.url)}" target="_blank" class="tid-link">${escapeHtml(tid.title)}</a></h4><p class="tid-id">TID: ${escapeHtml(tid.tid_id)}</p><p class="kb-description">${escapeHtml(tid.description)}</p>${tid.evidence && tid.evidence.length > 0 ? `<details><summary>Evidence Found</summary><div class="kb-evidence">${tid.evidence.map(ex => escapeHtml(ex)).join('<br>')}</div></details>` : ''}</div>`;
        });
        html += collapsible('Relevant Knowledge Base Articles', 'kb-header', body, true, report.matched_tids.length);
    }
    if (report.temporal_clusters && report.temporal_clusters.length > 0) {
        let body = '';
        report.temporal_clusters.forEach(c => {
            body += `<div class="finding warning"><div class="headline">${escapeHtml(c.description)}</div><div class="finding-meta">${c.event_count} occurrences between ${escapeHtml(c.window_start)} and ${escapeHtml(c.window_end)}</div>${c.sample_raw_log ? `<div class="log-block">${escapeHtml(c.sample_raw_log)}</div>` : ''}</div>`;
        });
        html += collapsible('Temporal Clusters (Retry Loops / Flapping)', 'warn-header', body, true, report.temporal_clusters.length);
    }
    if (report.kb_suggestions && report.kb_suggestions.length > 0) {
        let body = `<p class="kb-description">Log lines that no known pattern matched, correlated with similar Knowledge Base articles (TF-IDF similarity):</p>`;
        report.kb_suggestions.forEach(s => {
            body += `<div class="kb-article"><h4><a href="${escapeHtml(s.url)}" target="_blank" class="tid-link">${escapeHtml(s.tid_id)}: ${escapeHtml(s.title)}</a></h4><p class="tid-id">Similarity: ${(s.score * 100).toFixed(0)}%</p>${s.sample_line ? `<div class="log-block">${escapeHtml(s.sample_line)}</div>` : ''}</div>`;
        });
        html += collapsible('KB Suggestions (Fuzzy Match)', 'kb-header', body, true, report.kb_suggestions.length);
    }
    // Correlation graph container (rendered after DOM insertion by renderCorrelationGraph)
    html += `<details class="corr-graph-section report-section"><summary><h2 class="corr-graph-heading">Correlation Graph</h2></summary><div class="corr-graph-host" id="corrGraphHost"></div></details>`;

    return html;
}

// ==========================================================================
// Application State
// ==========================================================================

let currentReport = null;
let currentZoom = 1.0;

// UI Component instances
let progressBar = null;
let statusMessage = null;

// ==========================================================================
// Application Template
// ==========================================================================

document.querySelector('#app').innerHTML = `
    <div class="top-bar">
        <h2>${UI.APP_TITLE}</h2>
        <div class="top-bar-controls">
            <input id="filePath" type="text" class="file-path-input" placeholder="${UI.PLACEHOLDERS.FILE_PATH}" />
            <button id="browseBtn" class="btn btn-secondary">${UI.BUTTONS.BROWSE}</button>
            <label class="anonymize-label" title="${UI.TOOLTIPS.ANONYMIZE}">
                <input type="checkbox" id="anonymizeCheck" class="anonymize-check" /> ${UI.BUTTONS.ANONYMIZE || 'Anonymize PII'}
            </label>
            <button id="analyzeBtn" class="btn btn-primary">${UI.BUTTONS.ANALYZE}</button>
            <button id="exportPdfBtn" class="btn btn-success" style="display:none">${UI.BUTTONS.EXPORT_PDF || 'Export PDF'}</button>
            <button id="exportTxtBtn" class="btn btn-info" style="display:none">${UI.BUTTONS.EXPORT_TXT || 'Export TXT'}</button>
            <button id="exportJsonBtn" class="btn btn-secondary" style="display:none" title="Export structured, machine-readable JSON report">Export JSON</button>
            <div id="zoomControls" class="zoom-controls" style="display:none">
                <button id="zoomOutBtn" class="zoom-btn">${UI.BUTTONS.ZOOM_OUT}</button>
                <button id="zoomInBtn" class="zoom-btn">${UI.BUTTONS.ZOOM_IN}</button>
            </div>
            <button id="themeToggleBtn" class="btn btn-outline" title="Toggle dark/light mode" style="font-size: 18px; padding: 5px 10px;">☀️</button>
        </div>
    </div>
    <!-- Progress bar container (hidden by default) -->
    <div id="progressContainer" class="progress-container-wrapper">
        <div class="progress-container">
            <div class="progress-status" id="progressStatus">Initializing analysis...</div>
            <div class="progress-bar-wrapper">
                <div class="progress-bar-fill" id="progressFill"></div>
            </div>
            <div class="progress-percentage" id="progressPercentage">0%</div>
        </div>
    </div>
    <div id="statusMessageContainer" class="status-message-container"></div>
    <div class="pdf-content-area">
        <div id="resultBox" class="result-box report-wrapper">
            <div class="empty-state">
                <div class="icon">🔍</div>
                <p>Waiting for supportconfig file...</p>
            </div>
        </div>
    </div>
`;

// ==========================================================================
// DOM References
// ==========================================================================

const filePathInput = document.querySelector('#filePath');
const browseBtn = document.querySelector('#browseBtn');
const analyzeBtn = document.querySelector('#analyzeBtn');
const anonymizeCheck = document.querySelector('#anonymizeCheck');
const exportPdfBtn = document.querySelector('#exportPdfBtn');
const exportTxtBtn = document.querySelector('#exportTxtBtn');
const exportJsonBtn = document.querySelector('#exportJsonBtn');
const zoomInBtn = document.querySelector('#zoomInBtn');
const zoomOutBtn = document.querySelector('#zoomOutBtn');
const resultBox = document.querySelector('#resultBox');
const statusContainer = document.querySelector('#statusMessageContainer');
const zoomControls = document.querySelector('#zoomControls');

// Initialize progress bar component
const progressContainer = document.querySelector('#progressContainer');
const progressFill = document.querySelector('#progressFill');
const progressStatus = document.querySelector('#progressStatus');
const progressPercentage = document.querySelector('#progressPercentage');

// ==========================================================================
// Status Message Helper
// ==========================================================================

function showStatus(message, type = 'info') {
    if (statusMessage) {
        statusMessage.destroy();
    }
    statusMessage = new StatusMessage({
        message: message,
        type: type,
        dismissible: true,
        autoHide: type === 'success' ? 5000 : type === 'info' ? 3000 : 0
    });
    statusContainer.appendChild(statusMessage.getElement());
}

// ==========================================================================
// Event Handlers
// ==========================================================================

browseBtn.addEventListener('click', async () => {
    try {
        const path = await OpenFileBrowser();
        if (path) {
            filePathInput.value = path;
            // Validate the selected file
            const validation = FileValidator.validateFileExtension(path);
            if (!validation.isValid) {
                showStatus(validation.message, 'warning');
            }
        }
    } catch (error) {
        showStatus('Failed to open file browser: ' + error, 'error');
    }
});

analyzeBtn.addEventListener('click', async () => {
    const filePath = filePathInput.value.trim();
    
    // Validate file path
    if (!filePath) {
        showStatus('Please select a supportconfig file to analyze', 'warning');
        filePathInput.focus();
        return;
    }

    // Validate file extension
    const validation = FileValidator.validateFileExtension(filePath);
    if (!validation.isValid) {
        showStatus(validation.message, 'warning');
        return;
    }

    // Disable controls during analysis
    analyzeBtn.disabled = true;
    analyzeBtn.textContent = `${UI.BUTTONS.ANALYZE}...`;
    browseBtn.disabled = true;
    filePathInput.disabled = true;

    // Show progress bar and reset
    progressContainer.style.display = 'block';
    progressFill.style.width = '0%';
    progressStatus.textContent = 'Starting analysis...';
    progressPercentage.textContent = '0%';
    
    // Hide placeholder
    resultBox.innerHTML = '';

    try {
        const report = await Analyze(filePath, anonymizeCheck.checked);
        currentReport = report;
        resultBox.innerHTML = renderReportHTML(report);
        const corrHost = document.getElementById('corrGraphHost');
        if (corrHost) renderCorrelationGraph(corrHost, report);
        currentZoom = 1.0;
        applyZoom();
        exportPdfBtn.style.display = 'inline-block';
        exportTxtBtn.style.display = 'inline-block';
        exportJsonBtn.style.display = 'inline-block';
        zoomControls.style.display = 'inline-flex';
        showStatus('Analysis completed successfully', 'success');
    } catch (error) {
        showStatus('Analysis failed: ' + error, 'error');
        resultBox.innerHTML = `<div class="empty-state"><div class="icon">❌</div><p style="color: #d9534f;">Analysis failed: ${escapeHtml(String(error))}</p></div>`;
    } finally {
        progressContainer.style.display = 'none';
        analyzeBtn.disabled = false;
        analyzeBtn.textContent = UI.BUTTONS.ANALYZE;
        browseBtn.disabled = false;
        filePathInput.disabled = false;
    }
});

exportPdfBtn.addEventListener('click', () => {
    if (!currentReport) return;
    showStatus('Preparing PDF export...', 'info');
    const now = new Date();
    const timestamp = now.toISOString().replace(/T/, '_').replace(/:/g, '-').split('.')[0];
    const originalTitle = document.title;
    document.title = `SSSD_Analysis_Report_${timestamp}`;
    window.print();
    document.title = originalTitle;
});

exportTxtBtn.addEventListener('click', async () => {
    if (!currentReport) return;
    showStatus('Exporting text report...', 'info');
    try {
        const result = await SaveTXT(currentReport);
        if (result !== 'cancelled') {
            showStatus('Text report saved to: ' + result, 'success');
        } else {
            showStatus('Export cancelled', 'info');
        }
    } catch (error) {
        showStatus('Failed to export text: ' + error, 'error');
    }
});

exportJsonBtn.addEventListener('click', async () => {
    if (!currentReport) return;
    showStatus('Exporting JSON report...', 'info');
    try {
        const result = await SaveJSON(currentReport);
        if (result !== 'cancelled') {
            showStatus('JSON report saved to: ' + result, 'success');
        } else {
            showStatus('Export cancelled', 'info');
        }
    } catch (error) {
        showStatus('Failed to export JSON: ' + error, 'error');
    }
});

zoomInBtn.addEventListener('click', () => {
    if (currentZoom < ZOOM.MAX) {
        currentZoom = Math.round((currentZoom + ZOOM.STEP) * 10) / 10;
        applyZoom();
    }
});

zoomOutBtn.addEventListener('click', () => {
    if (currentZoom > ZOOM.MIN) {
        currentZoom = Math.round((currentZoom - ZOOM.STEP) * 10) / 10;
        applyZoom();
    }
});

function applyZoom() {
    resultBox.style.transform = `scale(${currentZoom})`;
    resultBox.style.transformOrigin = 'top left';
}

// ==========================================================================
// Wails Events
// ==========================================================================

EventsOn('analyze-progress', (message, percentage) => {
    if (progressFill && progressStatus && progressPercentage) {
        progressFill.style.width = `${percentage}%`;
        progressStatus.textContent = message;
        progressPercentage.textContent = `${percentage}%`;
    }
    
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
    if (files.length > 0) {
        const path = files[0];
        filePathInput.value = path;
        // Validate dropped file
        const validation = FileValidator.validateFileExtension(path);
        if (!validation.isValid) {
            showStatus(validation.message, 'warning');
        }
    }
}, false);

// ==========================================================================
// Keyboard Shortcuts
// ==========================================================================

document.addEventListener('keydown', (e) => {
    if (e.ctrlKey || e.metaKey) {
        switch(e.key) {
            case 'o': e.preventDefault(); browseBtn.click(); break;
            case 'Enter': e.preventDefault(); if (filePathInput.value.trim()) analyzeBtn.click(); break;
            case 'p': e.preventDefault(); if (exportPdfBtn.style.display !== 'none') exportPdfBtn.click(); break;
            case 's': e.preventDefault(); if (exportTxtBtn.style.display !== 'none') exportTxtBtn.click(); break;
            case 'j': e.preventDefault(); if (exportJsonBtn.style.display !== 'none') exportJsonBtn.click(); break;
        }
    }
});

// ==========================================================================
// Theme / Dark Mode Toggle
// ==========================================================================

function getPreferredTheme() {
    const saved = localStorage.getItem('sssd-inspector-theme');
    if (saved) {
        return saved; // 'dark' or 'light'
    }
    // If no saved preference, use system preference
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function applyTheme(theme) {
    const isDark = theme === 'dark';
    document.documentElement.classList.toggle('dark-mode', isDark);
    // Update toggle button icon
    const themeBtn = document.querySelector('#themeToggleBtn');
    if (themeBtn) {
        themeBtn.textContent = isDark ? '🌙' : '☀️';
        themeBtn.title = isDark ? 'Switch to light mode' : 'Switch to dark mode';
    }
}

function toggleTheme() {
    const currentTheme = document.documentElement.classList.contains('dark-mode') ? 'dark' : 'light';
    const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
    localStorage.setItem('sssd-inspector-theme', newTheme);
    applyTheme(newTheme);
}

// Initialize theme
const initialTheme = getPreferredTheme();
applyTheme(initialTheme);

// Theme toggle button click handler
document.querySelector('#themeToggleBtn').addEventListener('click', toggleTheme);

// Listen for system theme changes (only if user hasn't set a manual preference)
window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    if (!localStorage.getItem('sssd-inspector-theme')) {
        applyTheme(e.matches ? 'dark' : 'light');
    }
});

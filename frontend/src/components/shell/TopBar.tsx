// TopBar — application title, view switch, analysis controls and export/zoom
// actions. Markup and CSS classes are those of the legacy shell, plus the new
// .tab-bar used to reach the Definitions Studio.
import { UI, ZOOM } from '../../config/ui';
import type { Theme } from '../../hooks/useTheme';

export type TabKey = 'analysis' | 'definitions';

export interface TopBarProps {
  tab: TabKey;
  onTabChange: (tab: TabKey) => void;
  path: string;
  onPathChange: (value: string) => void;
  anonymize: boolean;
  onAnonymizeChange: (value: boolean) => void;
  busy: boolean;
  hasReport: boolean;
  analyzeLabel: string;
  onBrowse: () => void;
  onAnalyze: () => void;
  onExportPdf: () => void;
  onExportTxt: () => void;
  onExportJson: () => void;
  zoom: number;
  onZoomIn: () => void;
  onZoomOut: () => void;
  theme: Theme;
  onToggleTheme: () => void;
}

const TABS: Array<{ key: TabKey; label: string; hint: string }> = [
  { key: 'analysis', label: 'Analysis', hint: 'Analyze a supportconfig and read the report (Ctrl+1)' },
  { key: 'definitions', label: 'Definitions Studio', hint: 'Inspect, edit, validate and dry-run the YAML rules (Ctrl+2)' },
];

export function TopBar(props: TopBarProps) {
  const {
    tab, onTabChange, path, onPathChange, anonymize, onAnonymizeChange, busy,
    hasReport, analyzeLabel, onBrowse, onAnalyze, onExportPdf, onExportTxt,
    onExportJson, zoom, onZoomIn, onZoomOut, theme, onToggleTheme,
  } = props;

  return (
    <header className="top-bar">
      <h2>{UI.APP_TITLE}</h2>
      <nav className="tab-bar" role="tablist" aria-label="Views">
        {TABS.map((entry) => (
          <button
            key={entry.key}
            type="button"
            role="tab"
            id={`tab-${entry.key}`}
            aria-selected={tab === entry.key}
            aria-controls={`panel-${entry.key}`}
            className={`tab-btn${tab === entry.key ? ' tab-btn-active' : ''}`}
            title={entry.hint}
            onClick={() => onTabChange(entry.key)}
          >
            {entry.label}
          </button>
        ))}
      </nav>
      <div className="top-bar-controls">
        <input
          id="filePath"
          type="text"
          className="file-path-input"
          placeholder={UI.PLACEHOLDERS.FILE_PATH}
          value={path}
          disabled={busy}
          aria-label="Supportconfig path"
          onChange={(event) => onPathChange(event.target.value)}
        />
        <button id="browseBtn" type="button" className="btn btn-secondary" disabled={busy} onClick={onBrowse}>
          {UI.BUTTONS.BROWSE}
        </button>
        <label className="anonymize-label" title={UI.TOOLTIPS.ANONYMIZE}>
          <input
            type="checkbox"
            id="anonymizeCheck"
            className="anonymize-check"
            checked={anonymize}
            disabled={busy}
            onChange={(event) => onAnonymizeChange(event.target.checked)}
          />{' '}
          Anonymize PII
        </label>
        <button id="analyzeBtn" type="button" className="btn btn-primary" disabled={busy} onClick={onAnalyze}>
          {analyzeLabel}
        </button>
        {hasReport && (
          <>
            <button id="exportPdfBtn" type="button" className="btn btn-success" onClick={onExportPdf}>
              {UI.BUTTONS.EXPORT_PDF}
            </button>
            <button id="exportTxtBtn" type="button" className="btn btn-info" onClick={onExportTxt}>
              {UI.BUTTONS.EXPORT_TXT}
            </button>
            <button
              id="exportJsonBtn"
              type="button"
              className="btn btn-secondary"
              title="Export structured, machine-readable JSON report"
              onClick={onExportJson}
            >
              Export JSON
            </button>
            <div id="zoomControls" className="zoom-controls">
              <button
                id="zoomOutBtn"
                type="button"
                className="zoom-btn"
                disabled={zoom <= ZOOM.MIN}
                onClick={onZoomOut}
              >
                {UI.BUTTONS.ZOOM_OUT}
              </button>
              <button
                id="zoomInBtn"
                type="button"
                className="zoom-btn"
                disabled={zoom >= ZOOM.MAX}
                onClick={onZoomIn}
              >
                {UI.BUTTONS.ZOOM_IN}
              </button>
            </div>
          </>
        )}
        <button
          id="themeToggleBtn"
          type="button"
          className="btn btn-outline theme-toggle"
          title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          onClick={onToggleTheme}
        >
          {theme === 'dark' ? '🌙' : '☀️'}
        </button>
      </div>
    </header>
  );
}

export default TopBar;

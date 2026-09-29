// App — the application shell.
//
// Layout contract: the children rendered here are the direct children of #app
// (layout.css styles #app as a flex column), which is why the tab switch happens
// inside a Fragment instead of a wrapper <div>.
import { useCallback, useEffect, useRef, useState } from 'react';
import * as backend from './api/backend';
import { ProgressPanel } from './components/shell/ProgressPanel';
import { StatusBanner } from './components/shell/StatusBanner';
import { TopBar } from './components/shell/TopBar';
import type { TabKey } from './components/shell/TopBar';
import { ReportView } from './components/report/ReportView';
import { DefinitionsStudio } from './components/definitions/DefinitionsStudio';
import { useAnalysis } from './hooks/useAnalysis';
import { useStatus } from './hooks/useStatus';
import { useTheme } from './hooks/useTheme';

export interface AppProps {
  initialTab?: TabKey;
}

export function App({ initialTab = 'analysis' }: AppProps) {
  const status = useStatus();
  const analysis = useAnalysis(status);
  const { theme, toggle } = useTheme();
  const [tab, setTab] = useState<TabKey>(initialTab);

  // Drag & drop: OnFileDrop has no per-callback unsubscribe, so it is
  // registered once and the current handler is read from a ref.
  const dropHandler = useRef(analysis.acceptDroppedPath);
  dropHandler.current = analysis.acceptDroppedPath;
  useEffect(() => {
    backend.onFileDrop((paths) => {
      if (paths.length > 0) {
        dropHandler.current(paths[0]);
      }
    });
  }, []);

  // Keyboard shortcuts (parity with the legacy shell, plus view switching).
  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }
      switch (event.key) {
        case '1':
          event.preventDefault();
          setTab('analysis');
          break;
        case '2':
          event.preventDefault();
          setTab('definitions');
          break;
        case 'o':
          event.preventDefault();
          analysis.browse();
          break;
        case 'Enter':
          event.preventDefault();
          if (tab === 'analysis') {
            analysis.run();
          }
          break;
        case 'p':
          event.preventDefault();
          if (analysis.report) {
            analysis.exportPdf();
          }
          break;
        case 's':
          event.preventDefault();
          if (analysis.report) {
            analysis.exportTxt();
          }
          break;
        case 'j':
          event.preventDefault();
          if (analysis.report) {
            analysis.exportJson();
          }
          break;
        default:
          break;
      }
    };
    document.addEventListener('keydown', onKeyDown);
    return () => document.removeEventListener('keydown', onKeyDown);
  }, [analysis]);

  const errorText = analysis.error;

  const resultBody = useCallback(() => {
    if (errorText !== null) {
      return (
        <div className="empty-state">
          <div className="icon">❌</div>
          <p style={{ color: '#d9534f' }}>Analysis failed: {errorText}</p>
        </div>
      );
    }
    if (analysis.report) {
      return <ReportView report={analysis.report} />;
    }
    return (
      <div className="empty-state">
        <div className="icon">🔍</div>
        <p>Waiting for supportconfig file...</p>
      </div>
    );
  }, [analysis.report, errorText]);

  return (
    <>
      <TopBar
        tab={tab}
        onTabChange={setTab}
        path={analysis.path}
        onPathChange={analysis.setPath}
        anonymize={analysis.anonymize}
        onAnonymizeChange={analysis.setAnonymize}
        busy={analysis.busy}
        hasReport={analysis.report !== null}
        analyzeLabel={analysis.analyzeLabel}
        onBrowse={analysis.browse}
        onAnalyze={analysis.run}
        onExportPdf={analysis.exportPdf}
        onExportTxt={analysis.exportTxt}
        onExportJson={analysis.exportJson}
        zoom={analysis.zoom}
        onZoomIn={analysis.zoomIn}
        onZoomOut={analysis.zoomOut}
        theme={theme}
        onToggleTheme={toggle}
      />

      {analysis.progress !== null && (
        <ProgressPanel message={analysis.progress.message} percentage={analysis.progress.percentage} />
      )}

      <StatusBanner status={status.status} onDismiss={status.dismiss} />

      <div className="pdf-content-area">
        {tab === 'analysis' ? (
          <div
            className="result-box report-wrapper"
            id="resultBox"
            style={{ transform: `scale(${analysis.zoom})`, transformOrigin: 'top left' }}
          >
            {resultBody()}
          </div>
        ) : (
          <div className="result-box report-wrapper definitions-area" id="definitionsBox">
            <DefinitionsStudio status={status} goToAnalysis={() => setTab('analysis')} path={analysis.path} />
          </div>
        )}
      </div>
    </>
  );
}

export default App;

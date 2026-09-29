// useAnalysis — the analysis flow: path input, browsing, run, exports, zoom and
// progress tracking. Keeps App.tsx free of imperative Wails plumbing.
import { useCallback, useEffect, useState } from 'react';
import { UI, ZOOM } from '../config/constants.js';
import { validateArchivePath } from '../utils/fileValidation.js';
import * as backend from '../api/backend';
import type { ReportData } from '../api/backend';
import type { StatusApi } from './useStatus';

export interface ProgressState {
  message: string;
  percentage: number;
}

export interface AnalysisApi {
  path: string;
  setPath: (value: string) => void;
  anonymize: boolean;
  setAnonymize: (value: boolean) => void;
  busy: boolean;
  progress: ProgressState | null;
  report: ReportData | null;
  error: string | null;
  zoom: number;
  analyzeLabel: string;
  browse: () => void;
  run: () => void;
  acceptDroppedPath: (path: string) => void;
  zoomIn: () => void;
  zoomOut: () => void;
  exportPdf: () => void;
  exportTxt: () => void;
  exportJson: () => void;
}

const INITIAL_PROGRESS: ProgressState = { message: 'Starting analysis...', percentage: 0 };

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function useAnalysis(status: StatusApi): AnalysisApi {
  const { show } = status;
  const [path, setPath] = useState('');
  const [anonymize, setAnonymize] = useState(false);
  const [busy, setBusy] = useState(false);
  const [progress, setProgress] = useState<ProgressState | null>(null);
  const [report, setReport] = useState<ReportData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [zoom, setZoom] = useState(1.0);

  // Progress events: subscribed once, so a re-render can never double-subscribe.
  useEffect(() => backend.onAnalyzeProgress((message, percentage) => {
    setProgress({ message, percentage });
  }), []);

  // Definitions skipped while loading are reported by the backend; without this
  // banner a dropped rule file would be invisible (M0 gate).
  useEffect(() => backend.onDefinitionsWarning((diagnostics) => {
    if (Array.isArray(diagnostics) && diagnostics.length > 0) {
      show(
        `${diagnostics.length} definition problem(s) — some rules/articles were skipped. See the report for details.`,
        'warning',
      );
    }
  }), [show]);

  const warnAboutExtension = useCallback((candidate: string) => {
    const validation = validateArchivePath(candidate);
    if (!validation.isValid) {
      show(validation.message, 'warning');
    }
  }, [show]);

  const browse = useCallback(() => {
    void (async () => {
      try {
        const selected = await backend.openFileBrowser();
        if (selected) {
          setPath(selected);
          warnAboutExtension(selected);
        }
      } catch (err) {
        show(`Failed to open file browser: ${errorText(err)}`, 'error');
      }
    })();
  }, [show, warnAboutExtension]);

  const acceptDroppedPath = useCallback((dropped: string) => {
    setPath(dropped);
    warnAboutExtension(dropped);
  }, [warnAboutExtension]);

  const run = useCallback(() => {
    void (async () => {
      const target = path.trim();
      if (!target) {
        show('Please select a supportconfig file to analyze', 'warning');
        return;
      }
      const validation = validateArchivePath(target);
      if (!validation.isValid) {
        show(validation.message, 'warning');
        return;
      }

      setBusy(true);
      setProgress(INITIAL_PROGRESS);
      setReport(null);
      setError(null);
      try {
        const result = await backend.analyze(target, anonymize);
        setReport(result);
        setZoom(1.0);
        show('Analysis completed successfully', 'success');
      } catch (err) {
        const message = errorText(err);
        setError(message);
        show(`Analysis failed: ${message}`, 'error');
      } finally {
        setProgress(null);
        setBusy(false);
      }
    })();
  }, [anonymize, path, show]);

  const zoomIn = useCallback(() => {
    setZoom((current) => Math.min(ZOOM.MAX, Math.round((current + ZOOM.STEP) * 10) / 10));
  }, []);

  const zoomOut = useCallback(() => {
    setZoom((current) => Math.max(ZOOM.MIN, Math.round((current - ZOOM.STEP) * 10) / 10));
  }, []);

  // PDF export prints the report through the webview: the Go SavePDF binding
  // expects base64 data from a frontend PDF generator, so the print path
  // (window.print, as in M0) keeps the feature without extra dependencies.
  const exportPdf = useCallback(() => {
    if (!report) {
      return;
    }
    show('Preparing PDF export...', 'info');
    const stamp = new Date().toISOString().replace(/T/, '_').replace(/:/g, '-').split('.')[0];
    const originalTitle = document.title;
    document.title = `SSSD_Analysis_Report_${stamp}`;
    window.print();
    document.title = originalTitle;
  }, [report, show]);

  const saveWith = useCallback(
    (label: string, save: (data: ReportData) => Promise<string>) => {
      void (async () => {
        if (!report) {
          return;
        }
        show(`Exporting ${label} report...`, 'info');
        try {
          const saved = await save(report);
          if (saved === backend.CANCELLED) {
            show('Export cancelled', 'info');
          } else {
            show(`${label} report saved to: ${saved}`, 'success');
          }
        } catch (err) {
          show(`Failed to export ${label.toLowerCase()}: ${errorText(err)}`, 'error');
        }
      })();
    },
    [report, show],
  );

  const exportTxt = useCallback(() => saveWith('Text', backend.saveTxt), [saveWith]);
  const exportJson = useCallback(() => saveWith('JSON', backend.saveJson), [saveWith]);

  const analyzeLabel = busy
    ? progress && progress.percentage > 0 && progress.percentage < 100
      ? `${progress.percentage}% - ${progress.message}`
      : `${UI.BUTTONS.ANALYZE}...`
    : UI.BUTTONS.ANALYZE;

  return {
    path, setPath, anonymize, setAnonymize, busy, progress, report, error, zoom,
    analyzeLabel, browse, run, acceptDroppedPath, zoomIn, zoomOut, exportPdf, exportTxt, exportJson,
  };
}

// DiagnosticsList — the definitions that were skipped while loading.
//
// This is the M0 visibility guarantee: anything the fail-safe loaders dropped
// (malformed YAML, unreadable external KB article, duplicated rule name) is
// listed with its file and line instead of silently disappearing.
import type { Diagnostic } from '../../api/backend';
import { listOf } from '../../utils/payload';

export interface DiagnosticsListProps {
  diagnostics: Diagnostic[];
}

export function diagnosticsLocation(diag: Diagnostic): string {
  const line = typeof diag.line === 'number' ? diag.line : 0;
  return line > 0 ? `${diag.file}:${line}` : diag.file;
}

export function DiagnosticsList({ diagnostics }: DiagnosticsListProps) {
  // Guarded even though every caller passes listOf(...): a null array here
  // would take down the whole report, and the cost of the check is nil.
  return (
    <>
      {listOf(diagnostics).map((diag, index) => (
        <div className="finding warning" key={`${diag.file}-${diag.line ?? 0}-${index}`}>
          <div className="headline">{diag.message}</div>
          <div className="finding-meta">Source: {diagnosticsLocation(diag)}</div>
        </div>
      ))}
    </>
  );
}

export default DiagnosticsList;

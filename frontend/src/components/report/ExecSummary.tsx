// ExecSummary — health score, headline and severity counters.
import type { ExecutiveSummary } from '../../api/backend';

export interface ExecSummaryProps {
  /** Partial because the UI must survive a report produced by an older bridge. */
  summary: Partial<ExecutiveSummary>;
}

export function ExecSummary({ summary }: ExecSummaryProps) {
  const score = typeof summary.health_score === 'number' ? summary.health_score : 0;
  const scoreClass = score >= 80 ? 'success' : score >= 50 ? 'warn' : 'fail';

  return (
    <div className="exec-summary">
      <div className="exec-score">
        <div className="score-label">Health Score</div>
        <div className={`score ${scoreClass}`}>{score}</div>
        <div className="score-bar">
          <div className={`score-fill ${scoreClass}`} style={{ width: `${score}%` }} />
        </div>
      </div>
      <div className="exec-headline">
        <div className="headline">{summary.headline ?? ''}</div>
        <div className="exec-counts">
          <span className="exec-count count-critical">Critical: {summary.critical_count ?? 0}</span>
          <span className="exec-count count-error">Errors: {summary.error_count ?? 0}</span>
          <span className="exec-count count-warning">Warnings: {summary.warning_count ?? 0}</span>
          <span className="exec-count count-info">Log Patterns: {summary.log_error_count ?? 0}</span>
        </div>
      </div>
    </div>
  );
}

export default ExecSummary;

// RuleDryRunPanel — dry-run the loaded rules against a supportconfig.
//
// Firing rules come first (with the evidence line that made them fire);
// non-firing rules stay listed so "my rule did nothing" is answerable.
import type { RuleTestResult } from '../../api/backend';
import type { RuleDryRunApi } from '../../hooks/useRuleDryRun';
import { DiagnosticsList } from '../report/DiagnosticsList';

export interface RuleDryRunPanelProps {
  dryRun: RuleDryRunApi;
}

function DryRunResult({ result }: { result: RuleTestResult }) {
  const fired = result.outcomes.filter((outcome) => outcome.matched);
  const silent = result.outcomes.filter((outcome) => !outcome.matched);
  const diagnostics = result.diagnostics ?? [];

  const outcomeRow = (outcome: RuleTestResult['outcomes'][number], key: string) => (
    <div
      className={`studio-outcome${outcome.matched ? ' studio-outcome-fired' : ''}`}
      key={key}
      data-testid={outcome.matched ? 'dry-run-fired' : 'dry-run-silent'}
    >
      <span className="rule-name">{outcome.rule.name}</span>
      <span className="rule-meta">
        {' '}
        — {outcome.rule.severity || 'warning'}, {outcome.matched ? 'fires' : 'does not fire'} on {outcome.file}
        {outcome.line ? `:${outcome.line}` : ''} ({(outcome.rule.patterns || []).join(', ')})
      </span>
      {outcome.message ? <div>{outcome.message}</div> : null}
      {outcome.evidence ? <div className="studio-evidence">{outcome.evidence}</div> : null}
    </div>
  );

  return (
    <>
      <div className="studio-badges">
        <span className="studio-badge"><strong>{result.matched}</strong> of {result.total} rules fire</span>
        <span className="studio-badge">target: <code className="studio-path">{result.target_path}</code></span>
      </div>

      {diagnostics.length > 0 ? (
        <>
          <h3>Skipped definitions during this run</h3>
          <DiagnosticsList diagnostics={diagnostics} />
        </>
      ) : null}

      {result.outcomes.length === 0 ? (
        <p className="studio-empty">No rules are loaded, so nothing can fire.</p>
      ) : (
        <>
          <h3>Firing rules</h3>
          {fired.length === 0
            ? <p className="studio-empty">No loaded rule matches this supportconfig.</p>
            : fired.map((outcome, index) => outcomeRow(outcome, `fired-${index}`))}

          <h3>Rules that did not fire</h3>
          {silent.length === 0
            ? <p className="studio-empty">Every loaded rule fired.</p>
            : silent.map((outcome, index) => outcomeRow(outcome, `silent-${index}`))}
        </>
      )}
    </>
  );
}

export function RuleDryRunPanel({ dryRun }: RuleDryRunPanelProps) {
  const { target, setTarget, browse, run, running, result, clear } = dryRun;

  return (
    <section className="studio-panel" aria-label="Rule dry-run">
      <h2>Dry-run against a supportconfig</h2>
      <p className="panel-note">
        Runs the very matcher the analysis uses on the loaded rules. Nothing is written and no
        report is produced.
      </p>

      <div className="studio-field">
        <label htmlFor="dryRunTarget">Supportconfig archive or directory</label>
        <input
          id="dryRunTarget"
          type="text"
          value={target}
          placeholder="~/Downloads/supportconfig.txz"
          onChange={(event) => setTarget(event.target.value)}
        />
      </div>

      <div className="studio-toolbar">
        <button type="button" className="btn btn-secondary" onClick={browse} disabled={running}>
          Browse…
        </button>
        <button type="button" className="btn btn-primary" onClick={run} disabled={running}>
          {running ? 'Running…' : 'Run dry-run'}
        </button>
        {result !== null && (
          <button type="button" className="btn btn-outline" onClick={clear}>
            Clear
          </button>
        )}
      </div>

      {result !== null ? <DryRunResult result={result} /> : <p className="studio-empty">No dry-run yet.</p>}
    </section>
  );
}

export default RuleDryRunPanel;

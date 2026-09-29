// RuleEditorPanel — edit, validate and save one rules document.
//
// Every action is delegated to the Go service, so what the editor says about a
// document is what the loader (and therefore the analysis) will do with it.
import type { RuleValidationResult } from '../../api/backend';
import { SCOPE_SYSTEM, SCOPE_USER } from '../../api/backend';
import { listOf } from '../../utils/payload';
import type { RuleEditorApi } from '../../hooks/useRuleEditor';
import { SCHEMA_HELP } from './rulesTemplate';

export interface RuleEditorPanelProps {
  editor: RuleEditorApi;
}

function ValidationVerdict({ verdict }: { verdict: RuleValidationResult }) {
  const diagnostics = listOf(verdict.diagnostics);
  return (
    <div
      className={`studio-verdict ${verdict.valid ? 'studio-verdict-valid' : 'studio-verdict-invalid'}`}
      data-testid="rule-verdict"
    >
      <strong>{verdict.valid ? 'Valid' : 'Not valid'}</strong> — {verdict.label},{' '}
      {verdict.rule_count} rule(s) would load.
      {diagnostics.length > 0 ? (
        <ul className="studio-diagnostics">
          {diagnostics.map((diag, index) => (
            <li key={`${diag.file}-${diag.line ?? 0}-${index}`}>
              <div>{diag.message}</div>
              <div className="diag-location">{diag.line ? `${diag.file}:${diag.line}` : diag.file}</div>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

export function RuleEditorPanel({ editor }: RuleEditorPanelProps) {
  const {
    content, setContent, scope, setScope, docPath, docExists, docLoading, dirty,
    loadDocument, insertTemplate, validation, validating, validate, saveResult, saving, save,
  } = editor;

  const verdict = validation ?? saveResult?.validation ?? null;

  return (
    <section className="studio-panel" aria-label="Rule editor">
      <h2>Rule editor</h2>
      <p className="panel-note">
        {docLoading
          ? 'Loading the document…'
          : docExists
            ? `Editing ${docPath}${dirty ? ' (unsaved changes)' : ''}`
            : `No document yet in the ${scope} scope — a save creates ${docPath}`}
      </p>

      <div className="studio-toolbar">
        <button type="button" className="btn btn-secondary" onClick={loadDocument} disabled={docLoading}>
          Reload from disk
        </button>
        <button type="button" className="btn btn-secondary" onClick={insertTemplate}>
          Insert template
        </button>
        <button type="button" className="btn btn-primary" onClick={validate} disabled={validating}>
          {validating ? 'Validating…' : 'Validate'}
        </button>
        <button type="button" className="btn btn-success" onClick={save} disabled={saving}>
          {saving ? 'Saving…' : 'Save'}
        </button>
      </div>

      <div className="studio-scope" role="radiogroup" aria-label="Definition scope">
        <label>
          <input
            type="radio"
            name="definition-scope"
            value={SCOPE_USER}
            checked={scope === SCOPE_USER}
            onChange={() => setScope(SCOPE_USER)}
          />{' '}
          User ($HOME/.sssd-inspector)
        </label>
        <label>
          <input
            type="radio"
            name="definition-scope"
            value={SCOPE_SYSTEM}
            checked={scope === SCOPE_SYSTEM}
            onChange={() => setScope(SCOPE_SYSTEM)}
          />{' '}
          System (needs root)
        </label>
      </div>

      <div className="studio-field">
        <label htmlFor="rulesEditor">rules.yaml</label>
        <textarea
          id="rulesEditor"
          className="studio-editor"
          spellCheck={false}
          value={content}
          placeholder={'rules:\n  - name: "my-rule"\n    severity: warning\n    patterns: ["…"]\n    message: "…"\n'}
          onChange={(event) => setContent(event.target.value)}
        />
      </div>

      {verdict !== null ? <ValidationVerdict verdict={verdict} /> : null}

      {saveResult?.saved ? (
        <div className="studio-verdict studio-verdict-valid" data-testid="save-result">
          Saved to {saveResult.path} ({saveResult.bytes} bytes).
          {saveResult.backup ? ` Previous content kept as ${saveResult.backup}.` : ''}
        </div>
      ) : null}

      <details>
        <summary>Schema reference</summary>
        <div className="studio-schema">
          <table className="studio-table">
            <thead>
              <tr>
                <th>Field</th>
                <th>Required</th>
                <th>Meaning</th>
              </tr>
            </thead>
            <tbody>
              {SCHEMA_HELP.map((entry) => (
                <tr key={entry.field}>
                  <td><code>{entry.field}</code></td>
                  <td>{entry.required ? 'yes' : 'no'}</td>
                  <td>{entry.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>
    </section>
  );
}

export default RuleEditorPanel;

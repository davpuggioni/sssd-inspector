// DefinitionInventoryPanel — "where does the inspector look, and what is
// actually loaded right now?".
//
// Every row is one discovery location in loader order, with the counts the
// loader's own parsers produced and a real write probe, not permission bits.
import type { DefinitionsInventory, DefinitionScope } from '../../api/backend';
import { SCOPE_SYSTEM, SCOPE_USER } from '../../api/backend';
import { DiagnosticsList } from '../report/DiagnosticsList';

export interface DefinitionInventoryPanelProps {
  inventory: DefinitionsInventory | null;
  loading: boolean;
  loadError: string | null;
  onReload: () => void;
  onOpenFolder: (scope: DefinitionScope) => void;
}

function sizeLabel(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KiB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MiB`;
}

function fileState(info: DefinitionsInventory['files'][number]): string {
  if (!info.exists) {
    return 'missing';
  }
  if (info.is_dir) {
    return 'directory';
  }
  return info.writable ? 'present, writable' : 'present, read-only';
}

export function DefinitionInventoryPanel(props: DefinitionInventoryPanelProps) {
  const { inventory, loading, loadError, onReload, onOpenFolder } = props;
  // The last column counts options for a catalog override and bytes for
  // everything else, so its header follows the rows instead of lying.
  const hasCatalog = (inventory?.files ?? []).some((info) => info.kind === 'catalog');

  return (
    <section className="studio-panel" aria-label="Definition inventory">
      <h2>Discovery inventory</h2>
      <div className="studio-toolbar">
        <button type="button" className="btn btn-secondary" onClick={onReload} disabled={loading}>
          {loading ? 'Reloading…' : 'Reload'}
        </button>
        <button type="button" className="btn btn-outline" onClick={() => onOpenFolder(SCOPE_USER)}>
          Open user folder
        </button>
        <button type="button" className="btn btn-outline" onClick={() => onOpenFolder(SCOPE_SYSTEM)}>
          Open system folder
        </button>
      </div>

      {loadError !== null && <p className="studio-verdict studio-verdict-invalid">{loadError}</p>}

      {inventory === null ? (
        <p className="studio-empty">Loading inventory…</p>
      ) : (
        <>
          <div className="studio-badges">
            <span className="studio-badge"><strong>{inventory.rule_count}</strong> rules loaded</span>
            <span className="studio-badge"><strong>{inventory.article_count}</strong> KB articles</span>
            <span className="studio-badge">user root: <code className="studio-path">{inventory.user_root || 'n/a'}</code></span>
            <span className="studio-badge">system root: <code className="studio-path">{inventory.system_root || 'n/a'}</code></span>
          </div>

          <table className="studio-table">
            <thead>
              <tr>
                <th>Scope</th>
                <th>Kind</th>
                <th>Path</th>
                <th>State</th>
                <th className="num">Rules</th>
                <th className="num">Articles</th>
                <th className="num">{hasCatalog ? 'Options' : 'Size'}</th>
              </tr>
            </thead>
            <tbody>
              {inventory.files.map((info) => (
                <tr key={info.path}>
                  <td>{info.scope}</td>
                  <td>{info.kind}</td>
                  <td className="studio-path">{info.path}</td>
                  <td>{fileState(info)}</td>
                  <td className="num">{info.kind === 'catalog' ? '—' : info.rule_count}</td>
                  <td className="num">{info.kind === 'catalog' ? '—' : info.article_count}</td>
                  <td className="num">
                    {info.kind === 'catalog'
                      ? (info.option_count || (info.exists ? 'unusable' : '—'))
                      : (info.size_bytes ? sizeLabel(info.size_bytes) : '—')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {inventory.rules.length > 0 && (
            <>
              <h3>Loaded rules</h3>
              <table className="studio-table">
                <thead>
                  <tr>
                    <th>Rule</th>
                    <th>Severity</th>
                    <th>Category</th>
                    <th>Source</th>
                  </tr>
                </thead>
                <tbody>
                  {inventory.rules.map((info) => (
                    <tr key={`${info.file}-${info.line}-${info.rule.name}`}>
                      <td className="rule-name">{info.rule.name}</td>
                      <td>{info.rule.severity}</td>
                      <td>{info.rule.category || '—'}</td>
                      <td className="studio-path">
                        {info.file}{info.line ? `:${info.line}` : ''} ({info.scope})
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </>
          )}

          {inventory.diagnostics && inventory.diagnostics.length > 0 && (
            <>
              <h3>Skipped input</h3>
              <DiagnosticsList diagnostics={inventory.diagnostics} />
            </>
          )}
        </>
      )}
    </section>
  );
}

export default DefinitionInventoryPanel;

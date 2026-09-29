// CatalogPanel — the SSSD option catalog in effect, and the way to replace it.
//
// Read-only about the catalog itself: the panel's job is to state which file
// produced the verdict (embedded, or an override in a definitions root) and
// where an override may come from. Installing one is a service operation
// (InstallCatalog), validated with the loader's own decoder, so an unusable
// catalog is refused instead of silently doing nothing.
import type { CatalogInfo, DefinitionSaveResult } from '../../api/backend';
import { DiagnosticsList } from '../report/DiagnosticsList';

export interface CatalogPanelProps {
  catalog: CatalogInfo | null;
  loading: boolean;
  installing: boolean;
  lastInstall: DefinitionSaveResult | null;
  onInstall: (scope: 'user' | 'system') => void;
}

export function CatalogPanel({ catalog, loading, installing, lastInstall, onInstall }: CatalogPanelProps) {
  if (loading && catalog === null) {
    return <p className="studio-empty">Loading catalog information…</p>;
  }
  if (catalog === null) {
    return <p className="studio-empty">Catalog information unavailable.</p>;
  }

  const diagnostics = catalog.diagnostics ?? [];
  const overridePaths = catalog.override_paths ?? [];

  return (
    <>
      <table className="studio-table">
        <tbody>
          <tr>
            <th>In effect</th>
            <td>
              {catalog.available ? (
                <>
                  {catalog.using_override ? 'override' : 'embedded in the binary'}{' '}
                  {catalog.effective !== 'embedded' && (
                    <code className="studio-path">{catalog.effective}</code>
                  )}
                </>
              ) : (
                'unavailable'
              )}
            </td>
          </tr>
          <tr>
            <th>Upstream release</th>
            <td>{catalog.version || 'unknown'}</td>
          </tr>
          <tr>
            <th>Generated</th>
            <td>{catalog.generated || 'unknown'}</td>
          </tr>
          <tr>
            <th>Contents</th>
            <td>{catalog.option_count} options in {catalog.section_count} sections</td>
          </tr>
          {catalog.sources && catalog.sources.length > 0 ? (
            <tr>
              <th>Documentation sources</th>
              <td>{catalog.sources.length} ({catalog.sources.slice(0, 3).join(', ')}{catalog.sources.length > 3 ? ', …' : ''})</td>
            </tr>
          ) : null}
          {overridePaths.length > 0 ? (
            <tr>
              <th>Override locations</th>
              <td>
                {overridePaths.map((path) => (
                  <div className="studio-path" key={path}>{path}</div>
                ))}
                <div className="studio-empty">
                  Drop a generated <code>catalog.json</code> in one of these to validate against a
                  newer SSSD release; the first one present wins.
                </div>
              </td>
            </tr>
          ) : null}
          {catalog.error ? (
            <tr>
              <th>Error</th>
              <td>{catalog.error}</td>
            </tr>
          ) : null}
        </tbody>

      </table>

      <div className="studio-toolbar">
        <button
          type="button"
          className="btn btn-primary"
          disabled={installing}
          onClick={() => onInstall('user')}
          title="Choose a generated catalog.json and install it for your user"
        >
          {installing ? 'Installing…' : 'Install catalog (user)'}
        </button>
        <button
          type="button"
          className="btn btn-secondary"
          disabled={installing}
          onClick={() => onInstall('system')}
          title="Install for the whole system (needs root)"
        >
          Install for system
        </button>
      </div>

      {lastInstall !== null ? (
        <div
          className={`studio-verdict ${lastInstall.saved ? 'studio-verdict-valid' : 'studio-verdict-invalid'}`}
          data-testid="catalog-install-result"
        >
          {lastInstall.saved
            ? `Installed to ${lastInstall.path} (${lastInstall.bytes} bytes).`
            : `Not installed: ${lastInstall.path}`}
          {lastInstall.backup ? ` Previous catalog kept as ${lastInstall.backup}.` : ''}
        </div>
      ) : null}

      {diagnostics.length > 0 ? (
        <>
          <h3>Skipped catalog override</h3>
          <DiagnosticsList diagnostics={diagnostics} />
        </>
      ) : null}
    </>
  );
}

export default CatalogPanel;

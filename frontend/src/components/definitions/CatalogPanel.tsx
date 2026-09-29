// CatalogPanel — the SSSD option catalog in effect.
//
// Read-only by design: the catalog decides what "valid option" means for every
// configuration finding, so the panel's job is to state which file produced
// that verdict (embedded, or an override in a definitions root) and where an
// override may be dropped.
import type { CatalogInfo } from '../../api/backend';
import { DiagnosticsList } from '../report/DiagnosticsList';

export interface CatalogPanelProps {
  catalog: CatalogInfo | null;
  loading: boolean;
}

export function CatalogPanel({ catalog, loading }: CatalogPanelProps) {
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

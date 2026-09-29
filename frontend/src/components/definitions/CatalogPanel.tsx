// CatalogPanel — the embedded offline SSSD option catalog.
//
// Read-only (GetCatalogInfo): the catalog ships with the binary, so the Studio
// can only show which release its configuration claims are based on.
import type { CatalogInfo } from '../../api/backend';

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

  return (
    <table className="studio-table">
      <tbody>
        <tr>
          <th>Source</th>
          <td>
            {catalog.available ? 'available' : 'unavailable'}
            {catalog.source ? ` — ${catalog.source}` : ''}
          </td>
        </tr>
        <tr><th>Upstream release</th><td>{catalog.version || 'unknown'}</td></tr>
        <tr><th>Generated</th><td>{catalog.generated || 'unknown'}</td></tr>
        <tr>
          <th>Contents</th>
          <td>
            {catalog.option_count} options in {catalog.section_count} sections
          </td>
        </tr>
        {catalog.sources && catalog.sources.length > 0 ? (
          <tr>
            <th>Upstream sources</th>
            <td>{catalog.sources.join(', ')}</td>
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
  );
}

export default CatalogPanel;

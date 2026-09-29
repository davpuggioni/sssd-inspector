// useDefinitionsInventory — loads the discovery inventory and the catalog
// description, and owns the "open the definitions folder" affordance.
import { useCallback, useEffect, useState } from 'react';
import * as backend from '../api/backend';
import type { CatalogInfo, DefinitionsInventory, DefinitionScope } from '../api/backend';
import type { StatusApi } from './useStatus';

export interface DefinitionsInventoryApi {
  inventory: DefinitionsInventory | null;
  catalog: CatalogInfo | null;
  loading: boolean;
  loadError: string | null;
  reload: () => void;
  openFolder: (scope: DefinitionScope) => void;
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function useDefinitionsInventory(status: StatusApi): DefinitionsInventoryApi {
  const { show } = status;
  const [inventory, setInventory] = useState<DefinitionsInventory | null>(null);
  const [catalog, setCatalog] = useState<CatalogInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const reload = useCallback(() => {
    void (async () => {
      setLoading(true);
      setLoadError(null);
      try {
        const [inv, cat] = await Promise.all([backend.listDefinitions(), backend.getCatalogInfo()]);
        setInventory(inv);
        setCatalog(cat);
      } catch (err) {
        const message = errorText(err);
        setLoadError(message);
        show(`Cannot list definitions: ${message}`, 'error');
      } finally {
        setLoading(false);
      }
    })();
  }, [show]);

  useEffect(reload, [reload]);

  const openFolder = useCallback((scope: DefinitionScope) => {
    void (async () => {
      try {
        await backend.openDefinitionsRoot(scope);
      } catch (err) {
        show(`Cannot open the ${scope} definitions folder: ${errorText(err)}`, 'error');
      }
    })();
  }, [show]);

  return { inventory, catalog, loading, loadError, reload, openFolder };
}

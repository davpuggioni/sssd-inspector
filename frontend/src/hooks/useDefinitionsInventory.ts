// useDefinitionsInventory — loads the discovery inventory and the catalog
// description, owns the "open the definitions folder" affordance, and installs
// a catalog override.
import { useCallback, useEffect, useState } from 'react';
import * as backend from '../api/backend';
import type {
  CatalogInfo,
  DefinitionSaveResult,
  DefinitionsInventory,
  DefinitionScope,
} from '../api/backend';
import type { StatusApi } from './useStatus';

export interface DefinitionsInventoryApi {
  inventory: DefinitionsInventory | null;
  catalog: CatalogInfo | null;
  loading: boolean;
  loadError: string | null;
  reload: () => void;
  openFolder: (scope: DefinitionScope) => void;
  /** Install a generated catalog.json as the override for a scope. */
  installCatalog: (scope: DefinitionScope) => Promise<void>;
  installing: boolean;
  /** Result of the last install, for the panel to render. */
  lastInstall: DefinitionSaveResult | null;
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
  const [installing, setInstalling] = useState(false);
  const [lastInstall, setLastInstall] = useState<DefinitionSaveResult | null>(null);

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

  const installCatalog = useCallback(async (scope: DefinitionScope) => {
    setInstalling(true);
    setLastInstall(null);
    try {
      const selected = await backend.openCatalogFile();
      if (!selected) {
        return; // the user closed the chooser
      }
      const result = await backend.installCatalog(selected, scope);
      setLastInstall(result);
      show(
        result.saved
          ? `Catalog installed to ${result.path} (${result.bytes} bytes)`
          : `Catalog not installed: ${result.path}`,
        result.saved ? 'success' : 'warning',
      );
      // Show the new state: the panel must not keep claiming the old release.
      reload();
    } catch (err) {
      show(`Cannot install the catalog: ${errorText(err)}`, 'error');
    } finally {
      setInstalling(false);
    }
  }, [reload, show]);

  return {
    inventory,
    catalog,
    loading,
    loadError,
    reload,
    openFolder,
    installCatalog,
    installing,
    lastInstall,
  };
}

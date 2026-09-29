// DefinitionsStudio — where a support engineer goes to answer three questions
// about custom definitions: WHERE are they looked up, IS my file valid, DID
// my rule fire. All three are answered by definitions_service.go, the same code
// the analysis and the -definitions-info / -validate-rules / -rules-test flags
// use, so the Studio can never disagree with a real run.
import { useMemo } from 'react';
import { CatalogPanel } from './CatalogPanel';
import { DefinitionInventoryPanel } from './DefinitionInventoryPanel';
import { RuleDryRunPanel } from './RuleDryRunPanel';
import { RuleEditorPanel } from './RuleEditorPanel';
import { useDefinitionsInventory } from '../../hooks/useDefinitionsInventory';
import { useRuleDryRun } from '../../hooks/useRuleDryRun';
import { useRuleEditor } from '../../hooks/useRuleEditor';
import type { StatusApi } from '../../hooks/useStatus';

export interface DefinitionsStudioProps {
  /** Shared shell status banner. */
  status: StatusApi;
  /** Switches the shell back to the analysis view. */
  goToAnalysis: () => void;
  /** Path selected in the analysis view, prefilled in the dry-run. */
  path: string;
}

export function DefinitionsStudio({ status, goToAnalysis, path }: DefinitionsStudioProps) {
  const inventory = useDefinitionsInventory(status);
  // A save must be visible in the inventory without an extra manual reload.
  const editorOptions = useMemo(() => ({ onSaved: inventory.reload }), [inventory.reload]);
  const editor = useRuleEditor(status, editorOptions);
  const dryRun = useRuleDryRun(status, path);

  return (
    <div className="definitions-root">
      <div className="studio-header">
        <div>
          <h1>Definitions Studio</h1>
          <p>Custom YAML rules and knowledge base articles used by the analysis.</p>
        </div>
        <div className="studio-toolbar">
          {path ? (
            <button type="button" className="btn btn-secondary" onClick={goToAnalysis}>
              Back to analysis ({path})
            </button>
          ) : null}
        </div>
      </div>

      <div className="studio-grid">
        <DefinitionInventoryPanel
          inventory={inventory.inventory}
          loading={inventory.loading}
          loadError={inventory.loadError}
          onReload={inventory.reload}
          onOpenFolder={inventory.openFolder}
        />

        <RuleEditorPanel editor={editor} />

        <RuleDryRunPanel dryRun={dryRun} />

        <section className="studio-panel" aria-label="Option catalog">
          <h2>SSSD option catalog</h2>
          <p className="panel-note">
            The offline reference the analysis uses to validate sssd.conf against a specific
            upstream release.
          </p>
          <CatalogPanel catalog={inventory.catalog} loading={inventory.loading} />
        </section>
      </div>
    </div>
  );
}

export default DefinitionsStudio;

// useRuleDryRun — "which of my rules would fire on this supportconfig?".
//
// The dry-run reuses applyAnalysisRules on a throwaway report (TestRulesAgainst
// in Go), so an outcome here is exactly a finding the analysis would produce.
import { useCallback, useEffect, useState } from 'react';
import * as backend from '../api/backend';
import type { RuleTestResult } from '../api/backend';
import type { StatusApi } from './useStatus';

export interface RuleDryRunApi {
  target: string;
  setTarget: (value: string) => void;
  browse: () => void;
  run: () => void;
  running: boolean;
  result: RuleTestResult | null;
  clear: () => void;
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function useRuleDryRun(status: StatusApi, initialTarget: string): RuleDryRunApi {
  const { show } = status;
  const [target, setTarget] = useState(initialTarget);
  const [result, setResult] = useState<RuleTestResult | null>(null);
  const [running, setRunning] = useState(false);

  // Pre-fill with the path the analysis view is holding, so the common flow
  // (analyze a supportconfig, then tune the rules it triggered) is one click.
  useEffect(() => {
    if (initialTarget) {
      setTarget(initialTarget);
    }
  }, [initialTarget]);

  const browse = useCallback(() => {
    void (async () => {
      try {
        const selected = await backend.openFileBrowser();
        if (selected) {
          setTarget(selected);
        }
      } catch (err) {
        show(`Cannot open the file browser: ${errorText(err)}`, 'error');
      }
    })();
  }, [show]);

  const run = useCallback(() => {
    void (async () => {
      const value = target.trim();
      if (!value) {
        show('Select a supportconfig file or directory to dry-run the rules against.', 'warning');
        return;
      }
      setRunning(true);
      try {
        const outcome = await backend.testRulesAgainst(value);
        setResult(outcome);
        show(
          outcome.matched > 0
            ? `${outcome.matched} of ${outcome.total} rule(s) would fire.`
            : `None of the ${outcome.total} loaded rule(s) would fire.`,
          outcome.matched > 0 ? 'success' : 'info',
        );
      } catch (err) {
        setResult(null);
        show(`Dry-run failed: ${errorText(err)}`, 'error');
      } finally {
        setRunning(false);
      }
    })();
  }, [show, target]);

  const clear = useCallback(() => setResult(null), []);

  return { target, setTarget, browse, run, running, result, clear };
}

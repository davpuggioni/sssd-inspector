// useRuleEditor — the load → edit → validate → save loop of a rules document.
//
// The verdict always comes from the Go service (validateRuleYAML is the
// loader's own parser), so "valid" in the Studio means exactly "the analysis
// would load this document" — SaveRuleYAML refuses anything else anyway.
import { useCallback, useEffect, useRef, useState } from 'react';
import * as backend from '../api/backend';
import type { DefinitionSaveResult, DefinitionScope, RuleValidationResult } from '../api/backend';
import { SCOPE_USER } from '../api/backend';
import { STARTER_RULES } from '../components/definitions/rulesTemplate';
import type { StatusApi } from './useStatus';

export interface RuleEditorApi {
  content: string;
  setContent: (value: string) => void;
  scope: DefinitionScope;
  setScope: (scope: DefinitionScope) => void;
  docPath: string;
  docExists: boolean;
  docLoading: boolean;
  dirty: boolean;
  loadDocument: () => void;
  insertTemplate: () => void;
  validation: RuleValidationResult | null;
  validating: boolean;
  validate: () => void;
  saveResult: DefinitionSaveResult | null;
  saving: boolean;
  save: () => void;
}

export interface RuleEditorOptions {
  /** Called after a successful save so the inventory can be refreshed. */
  onSaved: () => void;
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function useRuleEditor(status: StatusApi, options: RuleEditorOptions): RuleEditorApi {
  const { show } = status;
  const [content, setContent] = useState('');
  const [scope, setScope] = useState<DefinitionScope>(SCOPE_USER);
  const [docPath, setDocPath] = useState('');
  const [docExists, setDocExists] = useState(false);
  const [docLoading, setDocLoading] = useState(false);
  const [validation, setValidation] = useState<RuleValidationResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [saveResult, setSaveResult] = useState<DefinitionSaveResult | null>(null);
  const [saving, setSaving] = useState(false);

  // Content as last read from (or written to) disk, for the dirty flag.
  const loadedRef = useRef('');

  const loadDocument = useCallback(() => {
    void (async () => {
      setDocLoading(true);
      setValidation(null);
      setSaveResult(null);
      try {
        const doc = await backend.readRuleYaml(scope);
        setDocPath(doc.path);
        setDocExists(doc.exists);
        const next = doc.exists ? doc.content : '';
        loadedRef.current = next;
        setContent(next);
        if (!doc.exists) {
          show(`No rules.yaml in the ${scope} scope yet — starting from an empty document.`, 'info');
        }
      } catch (err) {
        show(`Cannot read the ${scope} rules.yaml: ${errorText(err)}`, 'error');
      } finally {
        setDocLoading(false);
      }
    })();
  }, [scope, show]);

  // Reload when the scope changes: the editor must never show one scope's
  // document while a Save would write it into the other.
  useEffect(loadDocument, [loadDocument]);

  const insertTemplate = useCallback(() => {
    setContent((current) => (current.trim() === '' ? STARTER_RULES : `${current}\n${STARTER_RULES}`));
    setValidation(null);
  }, []);

  const validate = useCallback(() => {
    void (async () => {
      setValidating(true);
      try {
        const verdict = await backend.validateRuleYaml(content);
        setValidation(verdict);
        show(
          verdict.valid
            ? `Document is valid: ${verdict.rule_count} rule(s) would load.`
            : `Document is not valid: ${verdict.diagnostics?.length ?? 0} problem(s) — see the verdict.`,
          verdict.valid ? 'success' : 'warning',
        );
      } catch (err) {
        show(`Validation failed: ${errorText(err)}`, 'error');
      } finally {
        setValidating(false);
      }
    })();
  }, [content, show]);

  const save = useCallback(() => {
    void (async () => {
      setSaving(true);
      try {
        const result = await backend.saveRuleYaml(content, scope);
        setSaveResult(result);
        if (result.saved) {
          loadedRef.current = content;
          setDocExists(true);
          show(`Saved ${result.path} (${result.bytes} bytes)`, 'success');
          options.onSaved();
        } else {
          setValidation(result.validation);
          show('Save refused: the document is not valid — see the diagnostics.', 'warning');
        }
      } catch (err) {
        show(`Cannot save: ${errorText(err)}`, 'error');
      } finally {
        setSaving(false);
      }
    })();
  }, [content, options, scope, show]);

  return {
    content,
    setContent,
    scope,
    setScope,
    docPath,
    docExists,
    docLoading,
    dirty: content !== loadedRef.current,
    loadDocument,
    insertTemplate,
    validation,
    validating,
    validate,
    saveResult,
    saving,
    save,
  };
}

// DefinitionsStudio.test.tsx — the Studio must only ever show what the Go
// service returned, and must never report success for a refused save.
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useStatus } from '../hooks/useStatus';
import { mockBackend, makeCatalog, makeDryRun, makeInventory, makeValidation } from './helpers';

const backend = mockBackend();
vi.mock('../api/backend', () => backend);

const { DefinitionsStudio } = await import('../components/definitions/DefinitionsStudio');
const { StatusBanner } = await import('../components/shell/StatusBanner');

/**
 * Minimal harness: the Studio shares the shell's status API, so the banner is
 * rendered here too — a message the user must see is part of what is tested.
 */
function Studio() {
  const status = useStatus();
  return (
    <>
      <StatusBanner status={status.status} onDismiss={status.dismiss} />
      <DefinitionsStudio status={status} goToAnalysis={() => undefined} path="" />
    </>
  );
}

beforeEach(() => {
  // Re-seed every mock: a mockResolvedValue set by a previous test would
  // otherwise leak into the next one.
  Object.assign(backend, mockBackend());
  backend.listDefinitions.mockResolvedValue(makeInventory());
  backend.getCatalogInfo.mockResolvedValue(makeCatalog());
  backend.readRuleYaml.mockResolvedValue({ path: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', exists: true, bytes: 10, content: 'rules:\n  - name: "x"\n' });
});

describe('DefinitionsStudio — inventory', () => {
  it('shows every discovery location with its state and counts', async () => {
    render(<Studio />);
    await waitFor(() => expect(screen.getByText('/etc/sssd-inspector/rules.yaml')).toBeInTheDocument());
    expect(screen.getByText('missing')).toBeInTheDocument();
    expect(screen.getByText('present, writable')).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument(); // rule_count badge
  });

  it('surfaces skipped definition inputs from the inventory', async () => {
    backend.listDefinitions.mockResolvedValue(makeInventory({
      diagnostics: [{ file: '/etc/sssd-inspector/rules.yaml', line: 2, message: 'invalid rules file: bad indent', severity: 0 }],
    }));
    render(<Studio />);
    await waitFor(() => expect(screen.getByText(/Skipped input/)).toBeInTheDocument());
    expect(screen.getByText(/invalid rules file: bad indent/)).toBeInTheDocument();
  });

  it('shows the catalog override in effect, not just the embedded release', async () => {
    backend.getCatalogInfo.mockResolvedValue(makeCatalog({
      version: '2.16.0', generated: '2026-10-01', option_count: 1, section_count: 1,
      effective: '/home/u/.sssd-inspector/catalog.json', using_override: true,
    }));
    render(<Studio />);
    await waitFor(() => expect(screen.getByText('2.16.0')).toBeInTheDocument());
    expect(screen.getByText('override')).toBeInTheDocument();
    expect(screen.getAllByText('/home/u/.sssd-inspector/catalog.json').length).toBeGreaterThan(0);
    expect(screen.getByText(/Drop a generated/)).toBeInTheDocument();
  });

  it('lists a skipped catalog override instead of hiding it', async () => {
    backend.getCatalogInfo.mockResolvedValue(makeCatalog({
      diagnostics: [{ file: '/home/u/.sssd-inspector/catalog.json', line: 0, message: 'option catalog override skipped: malformed — falling back to the catalog embedded in the binary', severity: 0 }],
    }));
    render(<Studio />);
    await waitFor(() => expect(screen.getByText(/option catalog override skipped/)).toBeInTheDocument());
    expect(screen.getByText('Skipped catalog override')).toBeInTheDocument();
  });


  it('survives a payload with null arrays, as the Go bridge sends it', async () => {
    // encoding/json emits null for a nil slice, so an installation with NO
    // custom rules delivers "rules": null even though the generated model says
    // rules: RuleInfo[]. The Studio must render that, not crash on .length.
    const { makeInventory } = await import('./helpers');
    const payload = { ...makeInventory(), rules: null, diagnostics: null };
    backend.listDefinitions.mockResolvedValue(payload as unknown as ReturnType<typeof makeInventory>);

    render(<Studio />);

    expect(await screen.findByText('Discovery inventory')).toBeInTheDocument();
    // The count badge still renders, and no "Loaded rules" table is invented
    // out of a null array.
    expect(screen.getByText(/rules loaded/)).toBeInTheDocument();
    expect(screen.queryByText('Loaded rules')).toBeNull();
    expect(screen.getByRole('button', { name: 'Install catalog (user)' })).toBeInTheDocument();
  });

  it('survives a completely empty inventory payload', async () => {
    const { makeInventory } = await import('./helpers');
    const payload = {
      user_root: '/home/u/.sssd-inspector',
      system_root: '/etc/sssd-inspector',
      files: null, rules: null, rule_count: 0, article_count: 0, diagnostics: null,
    };
    backend.listDefinitions.mockResolvedValue(payload as unknown as ReturnType<typeof makeInventory>);

    render(<Studio />);
    expect(await screen.findByText('Discovery inventory')).toBeInTheDocument();
    expect(screen.getByText('Rule editor')).toBeInTheDocument();
    expect(screen.getByText(/KB articles/)).toBeInTheDocument();
  });

  it('installs a catalog override for the user scope and refreshes', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    await waitFor(() => expect(backend.getCatalogInfo).toHaveBeenCalled());
    const listCalls = backend.listDefinitions.mock.calls.length;

    await user.click(screen.getByRole('button', { name: 'Install catalog (user)' }));

    await waitFor(() => expect(backend.openCatalogFile).toHaveBeenCalled());
    expect(backend.installCatalog).toHaveBeenCalledWith('/tmp/catalog.json', 'user');
    const result = await screen.findByTestId('catalog-install-result');
    expect(result).toHaveTextContent('Installed to /home/u/.sssd-inspector/catalog.json');
    expect(result).toHaveTextContent('catalog.json.bak');
    // The inventory must be reloaded, or the panel would keep showing the
    // catalog that was in effect before the install.
    await waitFor(() => expect(backend.listDefinitions.mock.calls.length).toBeGreaterThan(listCalls));
  });

  it('installs for the system scope when asked', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Install for system' }));
    expect(backend.installCatalog).toHaveBeenCalledWith('/tmp/catalog.json', 'system');
  });

  it('does nothing when the file chooser is cancelled', async () => {
    const user = userEvent.setup();
    backend.openCatalogFile.mockResolvedValue('');
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Install catalog (user)' }));
    await waitFor(() => expect(backend.openCatalogFile).toHaveBeenCalled());
    expect(backend.installCatalog).not.toHaveBeenCalled();
    expect(screen.queryByTestId('catalog-install-result')).toBeNull();
  });

  it('reports a refused install instead of claiming success', async () => {
    const user = userEvent.setup();
    backend.installCatalog.mockResolvedValue({ path: '/home/u/.sssd-inspector/catalog.json', scope: 'user', saved: false, bytes: 0, backup: '', validation: { label: 'invalid', valid: false, rule_count: 0, rules: [], diagnostics: [] } });
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Install catalog (user)' }));
    const result = await screen.findByTestId('catalog-install-result');
    expect(result).toHaveTextContent('Not installed');
    expect(result).toHaveClass('studio-verdict-invalid');
  });

  it('surfaces an install failure from the service', async () => {
    const user = userEvent.setup();
    backend.installCatalog.mockRejectedValue(new Error('/tmp/catalog.json cannot be used as a catalog override: contains no options'));
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Install catalog (user)' }));
    await waitFor(() => expect(screen.getByTestId('status-message')).toHaveTextContent('contains no options'));
    expect(screen.queryByTestId('catalog-install-result')).toBeNull();
  });

});
  it('opens the definitions folder of the chosen scope', async () => {
    const user = userEvent.setup();

    render(<Studio />);
    await user.click(screen.getByRole('button', { name: 'Open user folder' }));
    expect(backend.openDefinitionsRoot).toHaveBeenCalledWith('user');
    await user.click(screen.getByRole('button', { name: 'Open system folder' }));
    expect(backend.openDefinitionsRoot).toHaveBeenCalledWith('system');
  });

describe('DefinitionsStudio — rule editor', () => {
  it('loads the current document of the selected scope', async () => {
    render(<Studio />);
    await waitFor(() => expect(backend.readRuleYaml).toHaveBeenCalledWith('user'));
    // getByDisplayValue does not match a <textarea> whose value was set by
    // state after mount, so read the value off the element directly.
    const editor = await screen.findByLabelText('rules.yaml') as HTMLTextAreaElement;
    expect(editor.value).toBe('rules:\n  - name: "x"\n');
  });

  it('reloads when the scope changes, so the editor never mixes scopes', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    await waitFor(() => expect(backend.readRuleYaml).toHaveBeenCalledWith('user'));
    await user.click(screen.getByRole('radio', { name: /System/ }));
    await waitFor(() => expect(backend.readRuleYaml).toHaveBeenCalledWith('system'));
  });

  it('renders the verdict returned by the Go validator', async () => {
    const user = userEvent.setup();
    backend.validateRuleYaml.mockResolvedValue(makeValidation(false));
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Validate' }));
    await waitFor(() => expect(screen.getByTestId('rule-verdict')).toHaveTextContent('Not valid'));
    expect(screen.getByText(/missing patterns or message/)).toBeInTheDocument();
  });

  it('reports a refused save as refused, with the diagnostics', async () => {
    const user = userEvent.setup();
    backend.saveRuleYaml.mockResolvedValue({ path: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', saved: false, bytes: 0, validation: makeValidation(false) });
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Save' }));
    await waitFor(() => expect(screen.getByTestId('rule-verdict')).toHaveTextContent('Not valid'));
    expect(screen.queryByTestId('save-result')).toBeNull();
  });

  it('confirms a successful save and refreshes the inventory', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    const callsBefore = backend.listDefinitions.mock.calls.length;
    await user.click(await screen.findByRole('button', { name: 'Save' }));
    await waitFor(() => expect(screen.getByTestId('save-result')).toHaveTextContent('Saved to'));
    await waitFor(() => expect(backend.listDefinitions.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('inserts a valid starter document the loader would accept', async () => {
    const user = userEvent.setup();
    backend.readRuleYaml.mockResolvedValue({ path: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', exists: false, bytes: 0, content: '' });
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Insert template' }));
    const editor = screen.getByLabelText('rules.yaml') as HTMLTextAreaElement;
    expect(editor.value).toContain('rules:');
    expect(editor.value).toContain('severity: warning');
  });
});

describe('DefinitionsStudio — dry run', () => {
  it('refuses to run without a target', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    await user.click(await screen.findByRole('button', { name: 'Run dry-run' }));
    expect(backend.testRulesAgainst).not.toHaveBeenCalled();
    expect(screen.getByText(/Select a supportconfig file or directory/)).toBeInTheDocument();
  });

  it('shows which rules fire and which do not, with the evidence line', async () => {
    const user = userEvent.setup();
    backend.testRulesAgainst.mockResolvedValue(makeDryRun());
    render(<Studio />);
    const input = screen.getByLabelText('Supportconfig archive or directory');
    await user.type(input, '/tmp/sc.txz');
    await user.click(screen.getByRole('button', { name: 'Run dry-run' }));
    await waitFor(() => expect(screen.getByTestId('dry-run-fired')).toHaveTextContent('fires'));
    expect(screen.getByTestId('dry-run-silent')).toHaveTextContent('does not fire');
    expect(screen.getByText('ldap_default_authtok = rc4-hmac')).toBeInTheDocument();
  });
});


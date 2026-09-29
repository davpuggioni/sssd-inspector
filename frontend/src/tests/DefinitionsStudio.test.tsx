// DefinitionsStudio.test.tsx — the Studio must only ever show what the Go
// service returned, and must never report success for a refused save.
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useStatus } from '../hooks/useStatus';
import { mockBackend, makeDryRun, makeInventory, makeValidation } from './helpers';

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
  backend.getCatalogInfo.mockResolvedValue({ source: 'embedded', version: '2.9', generated: '2026-01-01', option_count: 120, section_count: 8, available: true });
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

  it('opens the definitions folder of the chosen scope', async () => {
    const user = userEvent.setup();
    render(<Studio />);
    await user.click(screen.getByRole('button', { name: 'Open user folder' }));
    expect(backend.openDefinitionsRoot).toHaveBeenCalledWith('user');
    await user.click(screen.getByRole('button', { name: 'Open system folder' }));
    expect(backend.openDefinitionsRoot).toHaveBeenCalledWith('system');
  });
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


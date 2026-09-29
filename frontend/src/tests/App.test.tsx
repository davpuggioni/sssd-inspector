// App.test.tsx — the shell end to end with a stubbed backend.
//
// This is the closest thing to launching the GUI headlessly: mount the real
// App, switch views, run an analysis and check that the report reaches the
// screen. A regression in the wiring (state, effects, events) fails here
// rather than in a window nobody can inspect.
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockBackend, makeReport } from './helpers';
import type { ReportData } from '../api/backend';

const backend = mockBackend();
vi.mock('../api/backend', () => backend);

const { App } = await import('../App');

beforeEach(() => {
  Object.assign(backend, mockBackend());
});

describe('App — shell', () => {
  it('starts on the analysis view with an empty state', () => {
    render(<App />);
    expect(screen.getByText('SSSD Supportconfig Analyzer')).toBeInTheDocument();
    expect(screen.getByText('Waiting for supportconfig file...')).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Analysis' })).toHaveAttribute('aria-selected', 'true');
  });

  it('hides the export and zoom actions until there is a report', () => {
    render(<App />);
    expect(screen.queryByText('Export JSON')).toBeNull();
    expect(screen.queryByText('A+')).toBeNull();
  });

  it('switches between analysis and the Definitions Studio', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(screen.getByRole('tab', { name: /Definitions Studio/ }));
    expect(screen.queryByText('Waiting for supportconfig file...')).toBeNull();
    expect(await screen.findByText('Discovery inventory')).toBeInTheDocument();
    expect(screen.getByText('Rule editor')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: 'Analysis' }));
    expect(screen.getByText('Waiting for supportconfig file...')).toBeInTheDocument();
  });

  it('switches views with Ctrl+1 and Ctrl+2', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard('{Control>}2{/Control}');
    expect(await screen.findByText('Dry-run against a supportconfig')).toBeInTheDocument();
    await user.keyboard('{Control>}1{/Control}');
    expect(screen.getByText('Waiting for supportconfig file...')).toBeInTheDocument();
  });
});

describe('App — analysis flow', () => {
  it('refuses to analyze without a path', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    expect(backend.analyze).not.toHaveBeenCalled();
    expect(screen.getByTestId('status-message')).toHaveTextContent('Please select a supportconfig file');
  });

  it('refuses a path that is not a supportconfig archive', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/notes.txt');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    expect(backend.analyze).not.toHaveBeenCalled();
  });

  it('analyzes the selected archive and renders the report', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));

    expect(backend.analyze).toHaveBeenCalledWith('/tmp/supportconfig.txz', false);
    expect(await screen.findByRole('heading', { name: 'Analysis Report' })).toBeInTheDocument();
    expect(screen.getByText('RC4 enctype in sssd.conf')).toBeInTheDocument();
    // Export actions become available only once there is a report.
    expect(screen.getByText('Export JSON')).toBeInTheDocument();
    expect(screen.getByTestId('status-message')).toHaveTextContent('Analysis completed successfully');
  });

  it('passes the anonymize flag through and shows the failure state on error', async () => {
    const user = userEvent.setup();
    backend.analyze.mockRejectedValue(new Error('archive is corrupt'));
    render(<App />);
    await user.click(screen.getByRole('checkbox', { name: /Anonymize PII/ }));
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));

    // The message appears twice on purpose: once in the status banner, once in
    // the result area — the user must not be left with a stale report.
    await waitFor(() => expect(screen.getAllByText(/Analysis failed: archive is corrupt/).length).toBeGreaterThan(0));
    expect(backend.analyze).toHaveBeenCalledWith('/tmp/supportconfig.txz', true);
    expect(screen.queryByRole('heading', { name: 'Analysis Report' })).toBeNull();
  });

  it('exports the current report through the backend', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    await screen.findByRole('heading', { name: 'Analysis Report' });

    await user.click(screen.getByRole('button', { name: /Export JSON/ }));
    await waitFor(() => expect(backend.saveJson).toHaveBeenCalledTimes(1));
    expect(backend.saveJson).toHaveBeenCalledWith(expect.objectContaining({ sles_release: 'SLES 15 SP6' }));
    expect(await screen.findByText(/JSON report saved to: report\.json/)).toBeInTheDocument();
  });

  it('drops the previous report while a new analysis is running', async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    await screen.findByRole('heading', { name: 'Analysis Report' });

    backend.analyze.mockReturnValue(new Promise(() => undefined));
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    expect(screen.getByRole('button', { name: 'Analyze...' })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Analysis Report' })).toBeNull();
  });
});


// App — Wails events and OS drag & drop.
//
// These replace the legacy browser test-runner suites, which exercised the
// same behaviours by hand-rolling fake runtime objects. Here the shell runs
// for real and the registered callbacks are fired directly.
describe('App — Wails events', () => {
  it('shows analysis progress driven by the backend', async () => {
    const user = userEvent.setup();
    // A pending analysis keeps the shell in the "running" state, which is when
    // the progress events are supposed to be visible.
    backend.analyze.mockReturnValue(new Promise(() => undefined));
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));

    act(() => backend.events.progress[0]('Scanning SSSD logs…', 42));
    expect(await screen.findByText('Scanning SSSD logs…')).toBeInTheDocument();
    expect(screen.getByText('42%')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '42% - Scanning SSSD logs…' })).toBeInTheDocument();
  });

  it('completes the progress readout when the analysis returns', async () => {
    const user = userEvent.setup();
    let settle: ((report: ReportData) => void) | undefined;
    backend.analyze.mockReturnValue(new Promise<ReportData>((resolve) => { settle = resolve; }));
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));

    act(() => backend.events.progress[0]('Streaming and Parsing SSSD Logs...', 85));
    expect(await screen.findByText('85%')).toBeInTheDocument();

    await act(async () => { settle?.(makeReport()); });
    expect(await screen.findByRole('heading', { name: 'Analysis Report' })).toBeInTheDocument();
    expect(screen.queryByText('85%')).toBeNull();
  });

  it('ignores a progress event that arrives after the run has finished', async () => {
    const user = userEvent.setup();
    let settle: ((report: ReportData) => void) | undefined;
    backend.analyze.mockReturnValue(new Promise<ReportData>((resolve) => { settle = resolve; }));
    render(<App />);
    await user.type(screen.getByLabelText('Supportconfig path'), '/tmp/supportconfig.txz');
    await user.click(screen.getByRole('button', { name: 'Analyze' }));
    await act(async () => { settle?.(makeReport()); });
    expect(await screen.findByRole('heading', { name: 'Analysis Report' })).toBeInTheDocument();

    // Wails can deliver the final tick after the promise resolved: a stuck
    // progress bar over a finished report would look like a hung application.
    act(() => backend.events.progress[0]('Analysis complete!', 100));
    expect(screen.queryByText('100%')).toBeNull();
    expect(screen.getByRole('heading', { name: 'Analysis Report' })).toBeInTheDocument();
  });

  it('warns when the backend skipped definition files', async () => {
    render(<App />);
    act(() => backend.events.definitionsWarning[0]([
      { file: '/home/u/.sssd-inspector/rules.yaml', line: 4, message: 'rule "x": missing patterns or message, skipped', severity: 0 },
    ]));
    const banner = await screen.findByTestId('status-message');
    expect(banner).toHaveTextContent('1 definition problem(s)');
    expect(banner).toHaveClass('status-warning');
  });

  it('fills the path field from an OS file drop', async () => {
    render(<App />);
    act(() => backend.events.fileDrop[0](['/tmp/dropped-supportconfig.txz']));
    await waitFor(() => expect(screen.getByLabelText('Supportconfig path')).toHaveValue('/tmp/dropped-supportconfig.txz'));
  });

  it('warns about a dropped file the analyser cannot read', async () => {
    render(<App />);
    act(() => backend.events.fileDrop[0](['/tmp/notes.txt']));
    const banner = await screen.findByTestId('status-message');
    expect(banner).toHaveTextContent('Unsupported file format');
    // The path is still kept: the user may have renamed by mistake.
    expect(screen.getByLabelText('Supportconfig path')).toHaveValue('/tmp/notes.txt');
  });

  it('ignores an empty drop', async () => {
    render(<App />);
    act(() => backend.events.fileDrop[0]([]));
    expect(screen.getByLabelText('Supportconfig path')).toHaveValue('');
    expect(screen.queryByTestId('status-message')).toBeNull();
  });
});


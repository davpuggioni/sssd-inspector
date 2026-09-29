// ReportView.test.tsx — the "Risultati" view must show everything the Go
// analysis produced, including the skipped-definition diagnostics that the
// M0 gate depends on.
import { describe, expect, it } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import { ReportView } from '../components/report/ReportView';
import { makeReport } from './helpers';

describe('ReportView', () => {
  it('renders the executive summary from the report summary', () => {
    render(<ReportView report={makeReport()} />);
    expect(screen.getByText('72')).toBeInTheDocument();
    expect(screen.getByText('RC4 enctype in sssd.conf')).toBeInTheDocument();
    expect(screen.getByText(/Critical: 1/)).toBeInTheDocument();
    expect(screen.getByText(/Log Patterns: 4/)).toBeInTheDocument();
  });

  it('lists skipped definition inputs with their file and line', () => {
    render(<ReportView report={makeReport()} />);
    const section = screen.getByText(/Definition Problems/).closest('details');
    expect(section).not.toBeNull();
    expect(within(section as HTMLElement).getByText(/missing patterns or message/)).toBeInTheDocument();
    expect(within(section as HTMLElement).getByText(/\/home\/u\/\.sssd-inspector\/rules\.yaml:4/)).toBeInTheDocument();
  });

  it('shows configuration findings with provenance', () => {
    render(<ReportView report={makeReport()} />);
    expect(screen.getByText(/Configuration Findings/)).toBeInTheDocument();
    expect(screen.getByText(/Source: sssd\.conf \| Key: ad_config \| Line: 12/)).toBeInTheDocument();
    expect(screen.getByText(/Evidence: ldap_default_authtok/)).toBeInTheDocument();
  });

  it('renders the sssd.conf snippet with INI highlighting', () => {
    const { container } = render(<ReportView report={makeReport()} />);
    expect(container.querySelector('.ini-section')?.textContent).toBe('[sssd]');
    expect(container.querySelector('.ini-key')?.textContent).toBe('services');
  });

  it('omits empty sections instead of rendering empty headers', () => {
    const report = makeReport({ problems: [], warnings: [], mac_denial_examples: [], timeline: [], diagnostics: undefined });
    render(<ReportView report={report} />);
    expect(screen.queryByText(/Critical Problems Detected/)).toBeNull();
    expect(screen.queryByText(/Warnings & Recommendations/)).toBeNull();
    expect(screen.queryByText(/Event Timeline/)).toBeNull();
    expect(screen.queryByText(/Definition Problems/)).toBeNull();
  });

  it('reports a missing summary without crashing (older bridge payload)', () => {
    const report = makeReport();
    // A report from a bridge that predates the summary field.
    delete (report as unknown as { summary?: unknown }).summary;
    render(<ReportView report={report} />);
    expect(screen.getByText('0')).toBeInTheDocument();
  });

  it('links knowledge base articles and shows the dry-run style evidence blocks', () => {
    render(<ReportView report={makeReport()} />);
    const link = screen.getByRole('link', { name: 'SSSD fails with RC4' });
    expect(link).toHaveAttribute('href', 'https://example.com/tid/1234');
    expect(screen.getByText(/Temporal Clusters/)).toBeInTheDocument();
    expect(screen.getByText(/12 occurrences between/)).toBeInTheDocument();
  });
});

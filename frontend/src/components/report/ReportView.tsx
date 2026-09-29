// ReportView — the "Risultati" view: everything the Go analysis produced,
// in the same order and with the same CSS classes as the legacy renderer, but
// built from typed React components instead of one big HTML template string.
import type { ReportData } from '../../api/backend';
import { Section } from '../common/Section';
import { ConfigFindingsList } from './ConfigFindingsList';
import { CorrelationGraphSection } from './CorrelationGraphSection';
import { DiagnosticsList } from './DiagnosticsList';
import { ExecSummary } from './ExecSummary';
import { KbArticlesList } from './KbArticlesList';
import { KbSuggestionsList } from './KbSuggestionsList';
import { LogErrorsList } from './LogErrorsList';
import { SystemInfoTable } from './SystemInfoTable';
import { TemporalClustersList } from './TemporalClustersList';
import { TimelineList } from './TimelineList';

export interface ReportViewProps {
  report: ReportData;
}

function LinesList({ items, listClass }: { items?: string[]; listClass: string }) {
  if (!items || items.length === 0) {
    return null;
  }
  return (
    <ul className={listClass}>
      {items.map((item, index) => (
        <li key={`${index}-${item}`}>{item}</li>
      ))}
    </ul>
  );
}

export function ReportView({ report }: ReportViewProps) {
  const diagnostics = report.diagnostics ?? [];
  const temporalClusters = report.temporal_clusters ?? [];
  const kbSuggestions = report.kb_suggestions ?? [];

  return (
    <>
      <h1>Analysis Report</h1>

      <ExecSummary summary={report.summary ?? {}} />

      <div className="report-meta">
        <strong>Analysis Date:</strong> {report.timestamp}
        <br />
        {report.support_case_id ? (
          <>
            <strong>Support Case (SR#):</strong> {report.support_case_id}
          </>
        ) : null}
      </div>

      <SystemInfoTable report={report} />

      {diagnostics.length > 0 && (
        <Section
          title="Definition Problems (Skipped Input)"
          headingClass="warn-header"
          count={diagnostics.length}
        >
          <DiagnosticsList diagnostics={diagnostics} />
        </Section>
      )}

      {report.config_findings && report.config_findings.length > 0 && (
        <Section
          title="Configuration Findings (with provenance)"
          headingClass="warn-header"
          count={report.config_findings.length}
        >
          <ConfigFindingsList findings={report.config_findings} />
        </Section>
      )}

      {report.problems && report.problems.length > 0 && (
        <Section title="Critical Problems Detected" count={report.problems.length}>
          <LinesList items={report.problems} listClass="problem-list" />
        </Section>
      )}

      {report.warnings && report.warnings.length > 0 && (
        <Section title={'Warnings & Recommendations'} headingClass="warn-header" count={report.warnings.length}>
          <LinesList items={report.warnings} listClass="warn-list" />
        </Section>
      )}

      {report.sssd_log_errors && report.sssd_log_errors.length > 0 && (
        <Section title="SSSD Log Errors" count={report.sssd_log_errors.length}>
          <LogErrorsList errors={report.sssd_log_errors} />
        </Section>
      )}

      {report.mac_denial_examples && report.mac_denial_examples.length > 0 && (
        <Section title="MAC Security Denials" count={report.mac_denial_examples.length}>
          <div className="mac-block">
            {report.mac_denial_examples.map((example, index) => (
              <span key={`${index}-${example}`}>
                {example}
                {index < report.mac_denial_examples.length - 1 ? <br /> : null}
              </span>
            ))}
          </div>
        </Section>
      )}

      {report.timeline && report.timeline.length > 0 && (
        <Section title="Event Timeline" count={report.timeline.length}>
          <TimelineList events={report.timeline} />
        </Section>
      )}

      {report.matched_tids && report.matched_tids.length > 0 && (
        <Section
          title="Relevant Knowledge Base Articles"
          headingClass="kb-header"
          count={report.matched_tids.length}
        >
          <KbArticlesList articles={report.matched_tids} />
        </Section>
      )}

      {temporalClusters.length > 0 && (
        <Section
          title="Temporal Clusters (Retry Loops / Flapping)"
          headingClass="warn-header"
          count={temporalClusters.length}
        >
          <TemporalClustersList clusters={temporalClusters} />
        </Section>
      )}

      {kbSuggestions.length > 0 && (
        <Section title="KB Suggestions (Fuzzy Match)" headingClass="kb-header" count={kbSuggestions.length}>
          <KbSuggestionsList suggestions={kbSuggestions} />
        </Section>
      )}

      <CorrelationGraphSection report={report} />
    </>
  );
}

export default ReportView;

// ReportView — the "Risultati" view: everything the Go analysis produced,
// in the same order and with the same CSS classes as the legacy renderer, but
// built from typed React components instead of one big HTML template string.
import type { ReportData } from '../../api/backend';
import { listOf } from '../../utils/payload';
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
  // Every array field of a Go payload goes through listOf: a nil slice is
  // serialised as null, and a report with no findings is the normal case.
  const diagnostics = listOf(report.diagnostics);
  const temporalClusters = listOf(report.temporal_clusters);
  const kbSuggestions = listOf(report.kb_suggestions);
  const configFindings = listOf(report.config_findings);
  const problems = listOf(report.problems);
  const warnings = listOf(report.warnings);
  const logErrors = listOf(report.sssd_log_errors);
  const macDenials = listOf(report.mac_denial_examples);
  const timeline = listOf(report.timeline);
  const matchedTids = listOf(report.matched_tids);

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

      {configFindings.length > 0 && (
        <Section
          title="Configuration Findings (with provenance)"
          headingClass="warn-header"
          count={configFindings.length}
        >
          <ConfigFindingsList findings={configFindings} />
        </Section>
      )}

      {problems.length > 0 && (
        <Section title="Critical Problems Detected" count={problems.length}>
          <LinesList items={problems} listClass="problem-list" />
        </Section>
      )}

      {warnings.length > 0 && (
        <Section title={'Warnings & Recommendations'} headingClass="warn-header" count={warnings.length}>
          <LinesList items={warnings} listClass="warn-list" />
        </Section>
      )}

      {logErrors.length > 0 && (
        <Section title="SSSD Log Errors" count={logErrors.length}>
          <LogErrorsList errors={logErrors} />
        </Section>
      )}

      {macDenials.length > 0 && (
        <Section title="MAC Security Denials" count={macDenials.length}>
          <div className="mac-block">
            {macDenials.map((example, index) => (
              <span key={`${index}-${example}`}>
                {example}
                {index < macDenials.length - 1 ? <br /> : null}
              </span>
            ))}
          </div>
        </Section>
      )}

      {timeline.length > 0 && (
        <Section title="Event Timeline" count={timeline.length}>
          <TimelineList events={timeline} />
        </Section>
      )}

      {matchedTids.length > 0 && (
        <Section
          title="Relevant Knowledge Base Articles"
          headingClass="kb-header"
          count={matchedTids.length}
        >
          <KbArticlesList articles={matchedTids} />
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

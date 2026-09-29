// CorrelationGraphSection — hosts the interactive correlation graph.
//
// The graph itself is a React component (CorrelationGraph.tsx) with the layout
// in correlationGraph.ts; this section only decides when it is visible and
// renders the "nothing to show" case.
import type { ReportData } from '../../api/backend';
import { CorrelationGraph } from './CorrelationGraph';
import { Section } from '../common/Section';

export interface CorrelationGraphSectionProps {
  report: ReportData;
}

export function CorrelationGraphSection({ report }: CorrelationGraphSectionProps) {
  const graph = report.graph;
  const nodeCount = (graph?.entities?.length ?? 0) + (graph?.findings?.length ?? 0) + (graph?.sources?.length ?? 0);

  return (
    <Section title="Correlation Graph" className="corr-graph-section" defaultOpen={false}>
      {nodeCount === 0 ? (
        <p className="studio-empty">
          No entities, findings or evidence to correlate in this report.
        </p>
      ) : (
        <CorrelationGraph graph={graph} />
      )}
    </Section>
  );
}

export default CorrelationGraphSection;

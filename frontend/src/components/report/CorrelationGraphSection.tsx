// CorrelationGraphSection — hosts the dependency-free force-directed SVG graph.
//
// CorrelationGraph.js is a self-contained imperative renderer (no dependencies,
// built for the legacy DOM frontend); React mounts it into a host element and
// clears it on unmount instead of rewriting 346 lines of layout math.
import { useEffect, useRef } from 'react';
import type { ReportData } from '../../api/backend';
import { renderCorrelationGraph } from '../CorrelationGraph.js';
import { Section } from '../common/Section';

export interface CorrelationGraphSectionProps {
  report: ReportData;
}

export function CorrelationGraphSection({ report }: CorrelationGraphSectionProps) {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) {
      return undefined;
    }
    renderCorrelationGraph(host, report);
    return () => {
      host.innerHTML = '';
    };
  }, [report]);

  return (
    <Section title="Correlation Graph" className="corr-graph-section" defaultOpen={false}>
      <div className="corr-graph-host" ref={hostRef} data-testid="correlation-graph" />
    </Section>
  );
}

export default CorrelationGraphSection;

// ConfigFindingsList — structured, provenance-aware configuration findings.
//
// Each entry keeps the evidence trail the Go validator produced: source file,
// key and line, so a reviewer can jump to the exact sssd.conf line.
import type { ConfigFinding } from '../../api/backend';
import { severityLabel } from '../../api/backend';

export interface ConfigFindingsListProps {
  findings: ConfigFinding[];
}

export function ConfigFindingsList({ findings }: ConfigFindingsListProps) {
  return (
    <>
      {findings.map((finding, index) => {
        const cls = severityLabel(finding.severity);
        const line = finding.source_line && finding.source_line > 0 ? finding.source_line : 'n/a';
        return (
          <div className={`finding ${cls}`} key={`${finding.source_path}-${finding.source_key}-${index}`}>
            <div className="headline">{finding.message}</div>
            <div className="finding-meta">
              Source: {finding.source_path} | Key: {finding.source_key} | Line: {line}
              {finding.evidence ? (<><br />Evidence: {finding.evidence}</>) : null}
            </div>
          </div>
        );
      })}
    </>
  );
}

export default ConfigFindingsList;

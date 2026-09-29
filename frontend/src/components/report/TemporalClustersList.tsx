// TemporalClustersList — retry loops / flapping detected by the temporal
// correlation phase (event count inside a time window).
import type { TemporalCluster } from '../../api/backend';

export interface TemporalClustersListProps {
  clusters: TemporalCluster[];
}

export function TemporalClustersList({ clusters }: TemporalClustersListProps) {
  return (
    <>
      {clusters.map((cluster, index) => (
        <div className="finding warning" key={`${cluster.window_start}-${index}`}>
          <div className="headline">{cluster.description}</div>
          <div className="finding-meta">
            {cluster.event_count} occurrences between {cluster.window_start} and {cluster.window_end}
          </div>
          {cluster.sample_raw_log ? <div className="log-block">{cluster.sample_raw_log}</div> : null}
        </div>
      ))}
    </>
  );
}

export default TemporalClustersList;

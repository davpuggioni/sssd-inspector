// ProgressPanel — analysis progress readout driven by the 'analyze-progress'
// Wails event. Markup mirrors the legacy progress block so components.css and
// the print stylesheet keep applying unchanged.
export interface ProgressPanelProps {
  message: string;
  percentage: number;
}

export function ProgressPanel({ message, percentage }: ProgressPanelProps) {
  const clamped = Math.max(0, Math.min(100, percentage));
  return (
    <div className="progress-container-wrapper">
      <div className="progress-container">
        <div className="progress-status">{message}</div>
        <div className="progress-bar-wrapper">
          <div className="progress-bar-fill" style={{ width: `${clamped}%` }} />
        </div>
        <div className="progress-percentage">{clamped}%</div>
      </div>
    </div>
  );
}

export default ProgressPanel;

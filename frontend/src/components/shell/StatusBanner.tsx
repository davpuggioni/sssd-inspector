// StatusBanner — renders the single shell status message.
//
// The legacy UIComponents.js StatusMessage is an imperative DOM widget; this
// is its React replacement, using the same .status-message / .status-<type>
// classes that components.css already styles.
import type { StatusApi } from '../../hooks/useStatus';

export interface StatusBannerProps {
  status: StatusApi['status'];
  onDismiss: () => void;
}

export function StatusBanner({ status, onDismiss }: StatusBannerProps) {
  return (
    <div className="status-message-container">
      {status !== null && (
        <div
          className={`status-message status-${status.type}`}
          data-testid="status-message"
          role={status.type === 'error' ? 'alert' : 'status'}
        >
          <div className="status-content">{status.message}</div>
          <button type="button" className="status-dismiss" onClick={onDismiss} title="Dismiss">
            ×
          </button>
        </div>
      )}
    </div>
  );
}

export default StatusBanner;

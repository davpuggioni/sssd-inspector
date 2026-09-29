// useStatus — single transient status banner for the shell.
//
// Replaces the imperative UIComponents.js StatusMessage: one message at a
// time, auto-dismissed after a timeout unless it is an error (an error the
// user missed is an error that never happened).
import { useCallback, useEffect, useRef, useState } from 'react';

export type StatusType = 'info' | 'success' | 'warning' | 'error';

export interface StatusMessageState {
  message: string;
  type: StatusType;
}

export interface StatusApi {
  status: StatusMessageState | null;
  show: (message: string, type?: StatusType) => void;
  dismiss: () => void;
}

const AUTO_DISMISS_MS = 8000;

export function useStatus(): StatusApi {
  const [status, setStatus] = useState<StatusMessageState | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearTimer = useCallback(() => {
    if (timer.current !== null) {
      clearTimeout(timer.current);
      timer.current = null;
    }
  }, []);

  const dismiss = useCallback(() => {
    clearTimer();
    setStatus(null);
  }, [clearTimer]);

  const show = useCallback(
    (message: string, type: StatusType = 'info') => {
      clearTimer();
      setStatus({ message, type });
      if (type !== 'error') {
        timer.current = setTimeout(() => setStatus(null), AUTO_DISMISS_MS);
      }
    },
    [clearTimer],
  );

  useEffect(() => clearTimer, [clearTimer]);

  return { status, show, dismiss };
}

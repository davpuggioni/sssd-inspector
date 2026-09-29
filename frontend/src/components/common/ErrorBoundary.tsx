// ErrorBoundary — last-resort guard around the whole UI.
//
// A rendering bug in one report section must not leave the user with a blank
// window and a dead Analyze button, so the failure is shown with its message.
import { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  override componentDidCatch(error: Error, info: ErrorInfo): void {
    // eslint-disable-next-line no-console
    console.error('sssd-inspector UI error', error, info.componentStack);
  }

  override render(): ReactNode {
    const { error } = this.state;
    if (error === null) {
      return this.props.children;
    }
    return (
      <div className="result-box report-wrapper">
        <h1>Something went wrong</h1>
        <p>The user interface hit an unexpected error; the analysis engine is unaffected.</p>
        <pre className="log-block">{error.message}</pre>
        <button type="button" className="btn btn-primary" onClick={() => this.setState({ error: null })}>
          Try again
        </button>
      </div>
    );
  }
}

export default ErrorBoundary;

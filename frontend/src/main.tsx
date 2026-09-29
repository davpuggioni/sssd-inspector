// main.tsx — React entry point.
//
// Styles are imported in the same order as the legacy entry point (layout →
// components → report) so the existing cascade keeps applying unchanged;
// studio.css only adds the classes introduced by the React migration.
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import './styles/layout.css';
import './styles/components.css';
import './styles/report.css';
import './styles/studio.css';

import { App } from './App';
import { ErrorBoundary } from './components/common/ErrorBoundary';
import { backendAvailable } from './api/backend';
import { bootstrapTheme } from './hooks/useTheme';

// Apply the persisted/system theme before the first paint (no light flash).
bootstrapTheme();

const container = document.getElementById('app');
if (container === null) {
  throw new Error('mount point #app is missing from index.html');
}

if (!backendAvailable()) {
  // `npm run dev` in a plain browser: the UI renders, backend calls explain
  // themselves. This is also the case in the unit tests, which stub api/backend.
  console.warn('sssd-inspector: no Go bridge detected — running UI-only (backend calls will fail).');
}

createRoot(container).render(
  <StrictMode>
    <ErrorBoundary>
      <App />
    </ErrorBoundary>
  </StrictMode>,
);

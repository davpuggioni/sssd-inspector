// ui.ts — the user-visible strings and the zoom limits.
//
// This replaces config/constants.js (320 lines of which the UI used nine
// entries) and config/FrontendConfig.js, neither of which the React app needs:
// nothing here is dynamic, and a value that is only read in one component
// belongs next to that component. Keep it typed and small — if a constant is
// used once, inline it and delete the entry.
export const UI = {
  APP_TITLE: 'SSSD Supportconfig Analyzer',
  BUTTONS: {
    BROWSE: 'Browse...',
    ANALYZE: 'Analyze',
    EXPORT_PDF: '📄 Export PDF',
    EXPORT_TXT: '📝 Export TXT',
    ZOOM_IN: 'A+',
    ZOOM_OUT: 'A-',
  },
  PLACEHOLDERS: {
    FILE_PATH: 'Select or paste path to supportconfig.txz...',
  },
  TOOLTIPS: {
    ANONYMIZE: 'Redacts IP Addresses and Domain Names from the report',
  },
} as const;

/** Zoom bounds for the report view. */
export const ZOOM = {
  DEFAULT: 1.0,
  STEP: 0.1,
  MIN: 0.8,
  MAX: 1.5,
} as const;

/**
 * Archive formats the analyser accepts. Must stay in step with
 * constants.PatternSupportconfig in Go ("*.txz;*.tar.xz"), which is what the
 * open dialog offers and what extractIfArchive handles.
 */
export const SUPPORTED_EXTENSIONS = ['.txz', '.tar.xz'] as const;

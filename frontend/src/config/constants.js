/**
 * Frontend Constants for SSSD Inspector
 * Centralized configuration values for the frontend application
 * @version 2.0.0
 */

/**
 * UI Constants
 */
export const UI = {
  /** Application title and branding */
  APP_TITLE: 'SSSD Supportconfig Analyzer',
  APP_VERSION: '2.0.0',
  
  /** Button labels */
  BUTTONS: {
    BROWSE: 'Browse...',
    ANALYZE: 'Analyze',
    EXPORT_PDF: '📄 Export PDF',
    EXPORT_TXT: '📝 Export TXT',
    ZOOM_IN: 'A+',
    ZOOM_OUT: 'A-',
  },
  
  /** Input placeholders */
  PLACEHOLDERS: {
    FILE_PATH: 'Select or paste path to supportconfig.txz...',
  },
  
  /** Status messages */
  MESSAGES: {
    WAITING: 'Waiting for supportconfig file...',
    ANALYZING: 'Analyzing...',
    ANALYSIS_COMPLETE: 'Analysis Complete!',
    ERROR: 'Error occurred during analysis',
  },
  
  /** Tooltip text */
  TOOLTIPS: {
    ANONYMIZE: 'Redacts IP Addresses and Domain Names from the report',
  },
};

/**
 * CSS Class Names
 */
export const CSS_CLASSES = {
  /** Layout classes */
  LAYOUT: {
    TOP_BAR: 'top-bar',
    RESULT_BOX: 'result-box',
    REPORT_WRAPPER: 'report-wrapper',
    PDF_CONTENT_AREA: 'pdf-content-area',
  },
  
  /** Component classes */
  COMPONENTS: {
    BUTTON: 'btn',
    INPUT: 'input',
    INPUT_BOX: 'input-box',
    RESULT: 'result',
    ZOOM_CONTROLS: 'zoom-controls',
  },
  
  /** State classes */
  STATES: {
    SUCCESS: 'success',
    FAIL: 'fail',
    WARN: 'warn',
    SUCCESS_TEXT: 'success-text',
    DISABLED: 'disabled',
    HIDDEN: 'hidden',
  },
  
  /** Typography classes */
  TYPOGRAPHY: {
    HEADING_1: 'h1',
    HEADING_2: 'h2',
    SECTION_TITLE: 'section-title',
  },
  
  /** Table classes */
  TABLES: {
    INFO_TABLE: 'info-table',
  },
  
  /** List classes */
  LISTS: {
    PROBLEM_LIST: 'problem-list',
    WARN_LIST: 'warn-list',
  },
  
  /** Content classes */
  CONTENT: {
    LOG_BLOCK: 'log-block',
    MAC_BLOCK: 'mac-block',
    TIMELINE: 'timeline',
    TIMELINE_EVENT: 'timeline-event',
    TIMELINE_TIME: 'timeline-time',
    TIMELINE_MSG: 'timeline-msg',
    TIMELINE_RAW: 'timeline-raw',
  },
};

/**
 * Color Scheme
 */
export const COLORS = {
  /** Primary colors */
  PRIMARY: {
    MAIN: '#0056b3',
    LIGHT: '#e7f3ff',
    DARK: '#004085',
  },
  
  /** Secondary colors */
  SECONDARY: {
    MAIN: '#6c757d',
    LIGHT: '#f8f9fa',
    DARK: '#5a6268',
  },
  
  /** Status colors */
  STATUS: {
    SUCCESS: '#28a745',
    WARNING: '#ffc107',
    ERROR: '#dc3545',
    INFO: '#17a2b8',
  },
  
  /** Background colors */
  BACKGROUND: {
    MAIN: '#1e1e2e',
    LIGHT: '#f4f4f9',
    WHITE: '#ffffff',
  },
  
  /** Text colors */
  TEXT: {
    PRIMARY: '#333333',
    SECONDARY: '#666666',
    MUTED: '#777777',
    WHITE: '#ffffff',
  },
};

/**
 * Animation and Timing
 */
export const ANIMATION = {
  /** Durations in milliseconds */
  DURATIONS: {
    FAST: 200,
    NORMAL: 300,
    SLOW: 500,
  },
  
  /** Easing functions */
  EASING: {
    EASE: 'ease',
    EASE_IN: 'ease-in',
    EASE_OUT: 'ease-out',
    EASE_IN_OUT: 'ease-in-out',
  },
};

/**
 * File and Export Configuration
 */
export const FILES = {
  /** Supported file extensions */
  SUPPORTED_EXTENSIONS: ['.txz', '.tar.xz'],
  
  /** Export file names */
  EXPORT_NAMES: {
    PDF: 'SSSD_Analysis_Report.pdf',
    TXT: 'SSSD_Analysis_Report.txt',
  },
  
  /** File size limits */
  LIMITS: {
    MAX_FILE_SIZE: 100 * 1024 * 1024, // 100MB
    MAX_LINE_LENGTH: 1024 * 1024,    // 1MB
  },
};

/**
 * Zoom Configuration
 */
export const ZOOM = {
  /** Zoom levels */
  LEVELS: [0.8, 0.9, 1.0, 1.1, 1.2, 1.3, 1.4, 1.5],
  
  /** Default zoom level */
  DEFAULT: 1.0,
  
  /** Zoom step increment */
  STEP: 0.1,
  
  /** Minimum and maximum zoom */
  MIN: 0.8,
  MAX: 1.5,
};

/**
 * Progress Configuration
 */
export const PROGRESS = {
  /** Progress update interval in milliseconds */
  UPDATE_INTERVAL: 100,
  
  /** Progress steps with messages */
  STEPS: [
    { percentage: 0, message: 'Initializing streaming engine...' },
    { percentage: 10, message: 'Scanning Hardware & OS Data...' },
    { percentage: 25, message: 'Analyzing Network & Kerberos state...' },
    { percentage: 40, message: 'Evaluating SSSD Configurations...' },
    { percentage: 60, message: 'Streaming and Parsing SSSD Logs...' },
    { percentage: 85, message: 'Matching Knowledge Base Articles...' },
    { percentage: 95, message: 'Sanitizing PII data...' },
    { percentage: 100, message: 'Analysis Complete!' },
  ],
};

/**
 * Error Messages
 */
export const ERRORS = {
  /** File-related errors */
  FILE: {
    NOT_SELECTED: 'Please select a supportconfig file to analyze',
    INVALID_FORMAT: 'Invalid file format. Please select a .txz or .tar.xz file',
    TOO_LARGE: 'File size exceeds the maximum limit (100MB)',
    NOT_FOUND: 'The specified file was not found',
  },
  
  /** Analysis errors */
  ANALYSIS: {
    FAILED: 'Analysis failed. Please check the file and try again',
    TIMEOUT: 'Analysis timed out. Please try with a smaller file',
    CANCELLED: 'Analysis was cancelled',
  },
  
  /** Export errors */
  EXPORT: {
    FAILED: 'Failed to export report',
    NO_DATA: 'No analysis data available to export',
    CANCELLED: 'Export was cancelled',
  },
};

/**
 * Success Messages
 */
export const SUCCESS = {
  /** Analysis success */
  ANALYSIS: 'Analysis completed successfully',
  
  /** Export success */
  EXPORT: {
    PDF: 'PDF report exported successfully',
    TXT: 'Text report exported successfully',
  },
};

/**
 * Validation Rules
 */
export const VALIDATION = {
  /** File path validation */
  FILE_PATH: {
    MIN_LENGTH: 1,
    MAX_LENGTH: 1000,
    PATTERN: /^[a-zA-Z0-9.:_/\\ -]+$/,
  },
  
  /** Input validation */
  INPUT: {
    DEBOUNCE_DELAY: 300,
  },
};

/**
 * Event Names
 */
export const EVENTS = {
  /** Backend events */
  BACKEND: {
    ANALYZE_PROGRESS: 'analyze-progress',
  },
  
  /** Frontend events */
  FRONTEND: {
    FILE_SELECTED: 'file-selected',
    ANALYSIS_STARTED: 'analysis-started',
    ANALYSIS_COMPLETED: 'analysis-completed',
    ANALYSIS_FAILED: 'analysis-failed',
    EXPORT_STARTED: 'export-started',
    EXPORT_COMPLETED: 'export-completed',
  },
};

/**
 * Debug Configuration
 */
export const DEBUG = {
  /** Enable debug mode */
  ENABLED: false,
  
  /** Log levels */
  LOG_LEVELS: {
    ERROR: 0,
    WARN: 1,
    INFO: 2,
    DEBUG: 3,
  },
  
  /** Current log level */
  CURRENT_LOG_LEVEL: 2, // INFO
};

# Frontend Architecture Documentation

## Overview

This document provides a comprehensive overview of the frontend architecture, components, and development guidelines.

## Architecture Principles

### 1. Modular Design
- **Separation of Concerns**: Each module has a single responsibility
- **Loose Coupling**: Modules communicate through well-defined interfaces
- **High Cohesion**: Related functionality is grouped together
- **Reusability**: Components are designed for reuse across the application

### 2. Clean Code Principles
- **Descriptive Naming**: Functions, variables, and classes have clear, meaningful names
- **Single Responsibility**: Each function does one thing well
- **DRY (Don't Repeat Yourself)**: Common functionality is abstracted into utilities
- **Consistent Style**: Code follows consistent formatting and conventions

### 3. Modern JavaScript Standards
- **ES6+ Features**: Classes, modules, arrow functions, destructuring
- **Async/Await**: Clean handling of asynchronous operations
- **Module System**: ES6 imports/exports for dependency management
- **JSDoc Documentation**: Comprehensive documentation for all public APIs

## Directory Structure

```
frontend/
├── src/
│   ├── main.tsx                  # React entry point (mounts #app)
│   ├── App.tsx                   # Shell: top bar, view switch, status, shortcuts
│   ├── api/
│   │   └── backend.ts            # Typed façade over the generated Wails bindings
│   ├── hooks/
│   │   ├── useStatus.ts          # Single transient status message
│   │   ├── useTheme.ts           # Dark/light mode with persistence
│   │   ├── useAnalysis.ts        # Analyze, export, zoom, progress events
│   │   ├── useDefinitionsInventory.ts
│   │   ├── useRuleEditor.ts      # Load → edit → validate → save
│   │   └── useRuleDryRun.ts      # Dry-run against a supportconfig
│   ├── components/
│   │   ├── common/               # Section, ErrorBoundary
│   │   ├── shell/                # TopBar, ProgressPanel, StatusBanner
│   │   ├── report/               # ReportView and one component per section
│   │   └── definitions/          # Definitions Studio panels
│   ├── config/                   # Legacy JS constants/UI components
│   ├── utils/                    # Legacy JS validators + typed wrappers
│   ├── styles/                   # layout.css, components.css, report.css, studio.css
│   └── tests/                    # vitest suites (backend, ReportView, Studio)
├── tests/                        # Legacy browser test-runner (JS)
├── wailsjs/                      # Generated Wails bindings — DO NOT EDIT
├── tsconfig.json
├── vite.config.ts
└── package.json
```

`wailsjs/` is regenerated with `wails generate module` after any Go signature
change; `api/backend.ts` is the only place that imports it, so the rest of the
UI depends on its types and never on the raw bindings.

## Core Modules

### Configuration System

#### constants.js
Centralized configuration values for the frontend application:

```javascript
export const UI = {
  APP_TITLE: 'SSSD Supportconfig Analyzer',
  BUTTONS: {
    BROWSE: 'Browse...',
    ANALYZE: 'Analyze',
    EXPORT_PDF: '📄 Export PDF',
    EXPORT_TXT: '📝 Export TXT',
  },
  // ... more constants
};
```

#### FrontendConfig.js
Dynamic configuration management with validation and persistence:

```javascript
export class FrontendConfig {
  constructor(options = {}) {
    this.config = this.mergeWithDefaults(options);
    this.validators = this.createValidators();
    this.observers = new Set();
  }
  
  get(path, defaultValue = undefined) { /* ... */ }
  set(path, value, validate = true) { /* ... */ }
  update(updates, validate = true) { /* ... */ }
}
```

**Features:**
- Configuration validation with custom validators
- Observer pattern for change notifications
- LocalStorage persistence
- System preference detection (dark mode, reduced motion)
- Deep merge for nested configuration

### UI Components

#### Button Component
Reusable button with consistent styling and behavior:

```javascript
export class Button {
  constructor(options = {}) {
    this.id = options.id || `btn-${Date.now()}`;
    this.text = options.text || '';
    this.onClick = options.onClick || null;
    // ...
  }
  
  setText(text) { /* ... */ }
  setDisabled(disabled) { /* ... */ }
  destroy() { /* ... */ }
}
```

#### ProgressBar Component
Visual progress indicator with percentage and status:

```javascript
export class ProgressBar {
  constructor(options = {}) {
    this.min = options.min || 0;
    this.max = options.max || 100;
    this.showPercentage = options.showPercentage !== false;
    // ...
  }
  
  setValue(value, status = '') { /* ... */ }
  reset() { /* ... */ }
}
```

#### StatusMessage Component
Reusable notification component for user feedback:

```javascript
export class StatusMessage {
  constructor(options = {}) {
    this.type = options.type || 'info'; // info, success, warning, error
    this.dismissible = options.dismissible !== false;
    this.autoHide = options.autoHide || 0;
    // ...
  }
  
  update(message, type = this.type) { /* ... */ }
  show() { /* ... */ }
  hide() { /* ... */ }
}
```

#### FileInput Component
Advanced file input with drag-and-drop support:

```javascript
export class FileInput {
  constructor(options = {}) {
    this.accept = options.accept || '.txz,.tar.xz';
    this.enableDragDrop = options.enableDragDrop !== false;
    this.onFileSelect = options.onFileSelect || null;
    // ...
  }
  
  validateFile(file) { /* ... */ }
  clear() { /* ... */ }
  setDisabled(disabled) { /* ... */ }
}
```

### Utility Modules

#### validators.js
Comprehensive input validation system:

```javascript
export class FileValidator {
  static validateFilePath(filePath) { /* ... */ }
  static validateFileExtension(fileName) { /* ... */ }
  static validateFileSize(fileSize) { /* ... */ }
  static validateFile({ path, name, size }) { /* ... */ }
}

export class FormValidator {
  static validateAnalysisForm(form) { /* ... */ }
  static displayErrors(container, messages) { /* ... */ }
  static clearErrors(container) { /* ... */ }
}
```

#### helpers.js
General utility functions:

```javascript
export class FormatHelper {
  static formatBoolean(value, options = {}) { /* ... */ }
  static formatArray(array, options = {}) { /* ... */ }
  static formatTimestamp(timestamp, options = {}) { /* ... */ }
  static escapeHtml(text) { /* ... */ }
}

export class DOMHelper {
  static createElement(tagName, options = {}) { /* ... */ }
  static findElement(selector, parent = document) { /* ... */ }
  static addEventListener(element, event, handler, options = {}) { /* ... */ }
}

export class StorageHelper {
  static saveToLocalStorage(key, value) { /* ... */ }
  static getFromLocalStorage(key, defaultValue = null) { /* ... */ }
}
```

### Main Application

#### App.tsx
`App.tsx` is the shell: it composes the hooks, the top bar, the progress panel
and the status banner, and switches between the two views (Analysis and
Definitions Studio). It renders a Fragment on purpose — `layout.css` styles
`#app` as a flex column, so the top bar, progress, banner and content area must
stay its direct children.

```tsx
export function App({ initialTab = 'analysis' }: AppProps) {
  const status = useStatus();           // one transient message
  const analysis = useAnalysis(status); // path, run, exports, zoom
  const { theme, toggle } = useTheme();
  const [tab, setTab] = useState<TabKey>(initialTab);
  // ...
  return (
    <>
      <TopBar ... />
      {analysis.progress !== null && <ProgressPanel ... />}
      <StatusBanner status={status.status} onDismiss={status.dismiss} />
      <div className="pdf-content-area">
        {tab === 'analysis' ? <ReportView ... /> : <DefinitionsStudio ... />}
      </div>
    </>
  );
}
```

**Key Features:**
- State lives in hooks (`useStatus`, `useAnalysis`, `useTheme`, …), so the
  components stay declarative
- Backend access only through `api/backend.ts` (typed, bridge-guarded)
- An error boundary around the whole UI: a rendering bug must not leave a dead
  window
- Keyboard shortcuts: `Ctrl+O` browse, `Ctrl+Enter` analyze, `Ctrl+P/S/J`
  export, `Ctrl+1/2` switch view
- Wails events (`analyze-progress`, `definitions-warning`) and OS drag & drop
  are subscribed in effects, with cleanup

## CSS Architecture

### Modular CSS Structure

#### layout.css
- Base layout and grid system
- Responsive design breakpoints
- Typography and spacing utilities
- Print styles

#### components.css
- Component-specific styles
- Button variants and states
- Progress bar animations
- Status message types
- File input drag-and-drop styles

#### report.css
- Report display styling
- Table formatting
- Timeline visualization
- Knowledge base article layout
- Print-optimized styles

### CSS Best Practices
- **BEM-like naming**: Component-based class names
- **CSS Custom Properties**: Theme variables for consistency
- **Mobile-first**: Responsive design with progressive enhancement
- **Accessibility**: High contrast, reduced motion, screen reader support
- **Performance**: Optimized animations and transitions

## State Management

### Application State
```javascript
this.state = {
  currentReport: null,      // Analysis results
  currentZoom: 1.0,        // Report zoom level
  isAnalyzing: false,      // Analysis in progress
  selectedFile: null,      // Selected file information
  progress: {              // Progress tracking
    percentage: 0,
    message: ''
  }
};
```

### Configuration State
```javascript
this.config = {
  app: {                   // Application settings
    name: 'SSSD Supportconfig Analyzer',
    version: '2.0.0',
    debug: false,
    logLevel: 2
  },
  ui: {                    // UI preferences
    theme: 'light',
    language: 'en',
    fontSize: 'medium',
    animations: true
  },
  analysis: {             // Analysis settings
    autoStart: false,
    showProgress: true,
    timeout: 300000
  }
  // ... more configuration sections
};
```

### State Persistence
- **LocalStorage**: User preferences and application state
- **Session Storage**: Temporary analysis data
- **Configuration Export/Import**: Backup and restore settings

## Event System

### Frontend Events
```javascript
export const EVENTS = {
  FRONTEND: {
    FILE_SELECTED: 'file-selected',
    ANALYSIS_STARTED: 'analysis-started',
    ANALYSIS_COMPLETED: 'analysis-completed',
    ANALYSIS_FAILED: 'analysis-failed',
    EXPORT_STARTED: 'export-started',
    EXPORT_COMPLETED: 'export-completed'
  }
};
```

### Backend Integration
```javascript
// Progress events from Go backend
EventsOn('analyze-progress', (message, percentage) => {
  this.onAnalysisProgress(message, percentage);
});

// File drop support
OnFileDrop((files, x, y) => {
  this.onFileDrop(files);
});
```

## Testing Framework

### Test Structure
```
tests/
├── validators.test.js        # Validator unit tests
├── components.test.js        # Component unit tests
└── test-runner.html          # Browser-based test runner
```

### Test Categories

#### Unit Tests
- **Validator Tests**: Input validation logic
- **Component Tests**: UI component behavior
- **Helper Tests**: Utility function correctness

#### Integration Tests
- **Component Integration**: Component interaction
- **Backend Integration**: Wails bridge functionality
- **End-to-End**: Complete user workflows

### Test Runner Features
- **Browser-based execution**: Real DOM testing
- **Visual feedback**: Progress indicators and results
- **Console capture**: Complete test output logging
- **Result export**: JSON export for CI/CD integration

## Performance Optimization

### Code Splitting
- **Dynamic imports**: Load modules on demand
- **Component lazy loading**: Reduce initial bundle size
- **Route-based splitting**: Separate analysis and report views

### Memory Management
- **Event listener cleanup**: Prevent memory leaks
- **Component destruction**: Proper resource cleanup
- **State pruning**: Remove unnecessary data

### Rendering Optimization
- **Virtual DOM**: Efficient updates (if needed)
- **Debounced inputs**: Reduce validation frequency
- **Progressive loading**: Load large reports incrementally

## Accessibility

### WCAG 2.1 Compliance
- **Keyboard navigation**: Full keyboard support
- **Screen reader support**: Semantic HTML and ARIA labels
- **High contrast mode**: Enhanced visibility options
- **Reduced motion**: Respect user preferences

### Accessibility Features
- **Focus management**: Logical tab order and focus indicators
- **Alternative text**: Meaningful descriptions for images
- **Color contrast**: Sufficient contrast ratios
- **Resizable text**: Support for text scaling

## Security Considerations

### Input Validation
- **File type validation**: Only accept supported formats
- **Path traversal prevention**: Validate file paths
- **XSS prevention**: HTML escaping for user content

### Data Protection
- **PII anonymization**: Optional data redaction
- **Local storage encryption**: Sensitive data protection
- **Secure communication**: HTTPS for external resources

## Development Guidelines

### Code Style
- **ESLint configuration**: Consistent code formatting
- **Prettier integration**: Automatic code formatting
- **JSDoc documentation**: Complete API documentation
- **Type checking**: JSDoc types for better IDE support

### Git Workflow
- **Feature branches**: Isolate development work
- **Pull requests**: Code review process
- **Automated testing**: Pre-commit hooks
- **Documentation updates**: Keep docs in sync

### Build Process
- **Vite build tool**: Fast development and optimized builds
- **Asset optimization**: Minification and compression
- **Bundle analysis**: Monitor bundle size
- **Environment configuration**: Build-time variables

## Migration Guide

### From Legacy Code
1. **Replace inline styles**: Move to modular CSS files
2. **Extract constants**: Use centralized configuration
3. **Componentize UI**: Replace DOM manipulation with components
4. **Add validation**: Implement input validation
5. **Update event handling**: Use modern event patterns

### Breaking Changes
- **CSS class names**: Updated to follow BEM conventions
- **Component APIs**: New component-based architecture
- **Configuration format**: YAML-based configuration system
- **Testing approach**: New testing framework

## Future Enhancements

### Planned Features
- **Web Components**: Framework-agnostic components
- **PWA support**: Offline functionality
- **Internationalization**: Multi-language support

> **TypeScript migration**: done (M2, 2026-09-29) — the shell, the report view
> and the Definitions Studio are React 18 + TypeScript in `strict` mode, typed
> against the generated Wails models. `npm run build` type-checks first, and
> `npm test` runs the vitest suites. The remaining legacy JavaScript
> (`config/`, `utils/validators.js`, `components/UIComponents.js`,
> `components/CorrelationGraph.js`) is still exercised by the browser
> test-runner; porting or dropping it is a separate cleanup.

### Technical Debt
- **Legacy code removal**: Clean up deprecated files
- **Performance monitoring**: Add performance metrics
- **Error tracking**: Implement error reporting
- **Automated testing**: Expand test coverage

## Conclusion

The refactored SSSD Inspector frontend provides a solid foundation for future development with improved maintainability, testability, and user experience. The modular architecture allows for easy extension and modification while following modern web development best practices.

The comprehensive testing framework ensures code quality and reliability, while the documentation system provides clear guidance for developers working on the project.

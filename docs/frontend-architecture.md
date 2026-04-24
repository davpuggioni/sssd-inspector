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
│   ├── config/
│   │   ├── constants.js          # Centralized constants
│   │   └── FrontendConfig.js     # Dynamic configuration manager
│   ├── components/
│   │   └── UIComponents.js       # Reusable UI components
│   ├── utils/
│   │   ├── validators.js         # Input validation utilities
│   │   └── helpers.js            # General helper functions
│   ├── styles/
│   │   ├── layout.css            # Layout and responsive design
│   │   ├── components.css        # Component-specific styles
│   │   └── report.css            # Report display styles
│   ├── main.js                   # Main application entry point
│   ├── style.css                 # Legacy styles (deprecated)
│   └── app.css                   # Legacy styles (deprecated)
├── tests/
│   ├── validators.test.js        # Validator unit tests
│   ├── components.test.js        # Component unit tests
│   └── test-runner.html          # Test runner interface
└── package.json                  # Dependencies and build scripts
```

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

#### main.js
The main application class orchestrates all components:

```javascript
class SSSDInspectorApp {
  constructor() {
    this.config = getConfig();
    this.state = {
      currentReport: null,
      currentZoom: ZOOM.DEFAULT,
      isAnalyzing: false,
      selectedFile: null,
    };
    this.components = {};
    this.init();
  }
  
  async init() {
    this.setupConfiguration();
    this.createComponents();
    this.render();
    this.setupEventListeners();
    this.setupBackendEvents();
    this.setupKeyboardShortcuts();
    this.loadSavedState();
  }
}
```

**Key Features:**
- Component lifecycle management
- Event-driven architecture
- State management with persistence
- Keyboard shortcuts support
- Responsive design adaptation
- Error handling and user feedback

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
- **TypeScript migration**: Add static type checking
- **Web Components**: Framework-agnostic components
- **PWA support**: Offline functionality
- **Internationalization**: Multi-language support

### Technical Debt
- **Legacy code removal**: Clean up deprecated files
- **Performance monitoring**: Add performance metrics
- **Error tracking**: Implement error reporting
- **Automated testing**: Expand test coverage

## Conclusion

The refactored SSSD Inspector frontend provides a solid foundation for future development with improved maintainability, testability, and user experience. The modular architecture allows for easy extension and modification while following modern web development best practices.

The comprehensive testing framework ensures code quality and reliability, while the documentation system provides clear guidance for developers working on the project.

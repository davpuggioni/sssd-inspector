# Frontend Development Guide

## Getting Started

This guide provides comprehensive instructions for developing and maintaining the SSSD Inspector frontend application.

## Prerequisites

### Required Tools
- **Node.js** (v16 or higher)
- **npm** or **yarn** package manager
- **Modern web browser** (Chrome, Firefox, Safari, Edge)
- **Code editor** with JavaScript/HTML/CSS support

### Recommended Tools
- **VS Code** with extensions:
  - ES6 String HTML
  - Prettier - Code formatter
  - ESLint
  - Live Server
  - GitLens
  - JSDoc Generator

## Project Setup

### 1. Clone the Repository
```bash
git clone <repository-url>
cd sssd-inspector
```

### 2. Install Dependencies
```bash
cd frontend
npm install
# or
yarn install
```

### 3. Development Server
```bash
npm run dev
# or
yarn dev
```

This will start the Vite development server at `http://localhost:5173`

### 4. Build for Production
```bash
npm run build
# or
yarn build
```

The built files will be in the `dist/` directory.

## Development Workflow

### 1. Feature Development
```bash
# Create feature branch
git checkout -b feature/new-feature-name

# Make changes
# ... develop your feature ...

# Run tests
npm run test

# Run linting
npm run lint

# Commit changes
git add .
git commit -m "feat: add new feature"

# Push and create PR
git push origin feature/new-feature-name
```

### 2. Code Review Process
- Create pull request with descriptive title and description
- Ensure all tests pass
- Request code review from team members
- Address feedback and update as needed
- Merge after approval

### 3. Testing Strategy
```bash
# Run all tests
npm run test

# Run specific test file
npm run test:validators
npm run test:components

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

## Architecture Overview

### Module System
The frontend uses ES6 modules for code organization:

```javascript
// Import statements
import { Button, ProgressBar } from './components/UIComponents.js';
import { FileValidator } from './utils/validators.js';
import { UI, COLORS } from './config/constants.js';

// Export statements
export class MyComponent {
  // Component implementation
}

export default MyComponent;
```

### Component Structure
Each component follows a consistent structure:

```javascript
/**
 * Component Description
 * @version 2.0.0
 */

export class ComponentName {
  /**
   * Creates a new component instance
   * @param {Object} options - Component configuration
   */
  constructor(options = {}) {
    // Initialize properties
    this.options = options;
    this.element = this.createElement();
    this.setupEventListeners();
  }

  /**
   * Creates the DOM element
   * @returns {HTMLElement} Component element
   */
  createElement() {
    // Element creation logic
  }

  /**
   * Sets up event listeners
   */
  setupEventListeners() {
    // Event listener setup
  }

  /**
   * Public method
   */
  publicMethod() {
    // Method implementation
  }

  /**
   * Destroys the component
   */
  destroy() {
    // Cleanup logic
  }
}
```

## Coding Standards

### 1. JavaScript Standards

#### Naming Conventions
```javascript
// Constants: UPPER_SNAKE_CASE
export const MAX_FILE_SIZE = 100 * 1024 * 1024;

// Classes: PascalCase
export class FileValidator {
  // ...
}

// Functions and variables: camelCase
const selectedFile = null;
function validateFileInput() {
  // ...
}

// Private methods: prefix with underscore
class MyClass {
  _privateMethod() {
    // ...
  }
}
```

#### Function Documentation
```javascript
/**
 * Validates a file path and returns validation result
 * @param {string} filePath - The file path to validate
 * @param {Object} options - Validation options
 * @param {boolean} options.allowEmpty - Whether to allow empty paths
 * @returns {Object} Validation result with isValid and message properties
 * @throws {Error} If validation fails catastrophically
 * @example
 * const result = validateFilePath('/path/to/file.txt', { allowEmpty: false });
 * if (result.isValid) {
 *   console.log('Valid path');
 * }
 */
function validateFilePath(filePath, options = {}) {
  // Implementation
}
```

#### Error Handling
```javascript
// Use try-catch for async operations
try {
  const result = await someAsyncOperation();
  return result;
} catch (error) {
  console.error('Operation failed:', error);
  throw new Error(`Failed to complete operation: ${error.message}`);
}

// Validate inputs early
function processFile(filePath) {
  if (!filePath || typeof filePath !== 'string') {
    throw new Error('File path must be a non-empty string');
  }
  // Continue processing
}
```

### 2. CSS Standards

#### Class Naming (BEM-like)
```css
/* Block */
.button {
  /* Block styles */
}

/* Element */
.button__icon {
  /* Element styles */
}

/* Modifier */
.button--primary {
  /* Modifier styles */
}

.button--disabled {
  /* Modifier styles */
}
```

#### CSS Organization
```css
/* ==========================================================================
   Component Name
   ========================================================================== */

/* Base styles */
.component {
  /* Base component styles */
}

/* Variants */
.component--variant {
  /* Variant-specific styles */
}

/* States */
.component.is-active {
  /* Active state styles */
}

.component.is-disabled {
  /* Disabled state styles */
}

/* Responsive */
@media (max-width: 768px) {
  .component {
    /* Mobile styles */
  }
}
```

### 3. HTML Standards

#### Semantic HTML
```html
<!-- Use semantic elements -->
<main class="app">
  <header class="app__header">
    <h1 class="app__title">Application Title</h1>
  </header>
  
  <section class="app__content">
    <form class="file-form">
      <label for="file-input" class="file-form__label">Select File</label>
      <input id="file-input" class="file-form__input" type="file">
    </form>
  </section>
</main>
```

#### Accessibility
```html
<!-- Include ARIA labels -->
<button aria-label="Close dialog" class="dialog__close">×</button>

<!-- Use proper form labels -->
<label for="email">Email Address</label>
<input id="email" type="email" required aria-describedby="email-help">
<small id="email-help" class="form__help">Enter your email address</small>

<!-- Keyboard navigation -->
<div tabindex="0" role="button" class="custom-button">
  Custom Button
</div>
```

## Component Development

### 1. Creating a New Component

#### Step 1: Define Component Class
```javascript
// src/components/NewComponent.js
export class NewComponent {
  constructor(options = {}) {
    this.id = options.id || `component-${Date.now()}`;
    this.className = options.className || 'new-component';
    this.onClick = options.onClick || null;
    
    this.element = this.createElement();
    this.setupEventListeners();
  }

  createElement() {
    return DOMHelper.createElement('div', {
      attributes: { id: this.id },
      className: this.className,
      textContent: 'New Component'
    });
  }

  setupEventListeners() {
    if (this.onClick) {
      this.cleanup = DOMHelper.addEventListener(
        this.element,
        'click',
        this.onClick
      );
    }
  }

  destroy() {
    if (this.cleanup) {
      this.cleanup();
    }
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}
```

#### Step 2: Add Styles
```css
/* src/styles/components.css */
.new-component {
  padding: 12px;
  border: 1px solid #dee2e6;
  border-radius: 4px;
  background-color: #ffffff;
}

.new-component:hover {
  border-color: #0056b3;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
```

#### Step 3: Write Tests
```javascript
// tests/components/new-component.test.js
class NewComponentTests {
  static runAll() {
    this.testComponentCreation();
    this.testComponentEvents();
    this.testComponentDestruction();
  }

  static testComponentCreation() {
    const component = new NewComponent({ id: 'test-component' });
    const element = component.getElement();
    
    TestUtils.assert(element !== null, 'Component element should be created');
    TestUtils.assertEqual(element.id, 'test-component', 'Component should have correct ID');
    
    component.destroy();
  }

  // ... more tests
}
```

#### Step 4: Update Documentation
```markdown
## NewComponent

A reusable component for [purpose].

### Usage
```javascript
const component = new NewComponent({
  id: 'my-component',
  onClick: () => console.log('Clicked!')
});
```

### Options
- `id` (string): Component ID
- `className` (string): Additional CSS classes
- `onClick` (function): Click event handler

### Methods
- `getElement()`: Returns the DOM element
- `destroy()`: Cleans up the component
```

### 2. Component Integration

#### Add to Main Application
```javascript
// In main.js
import { NewComponent } from './components/NewComponent.js';

class SSSDInspectorApp {
  createComponents() {
    // ... existing components ...
    
    this.components.newComponent = new NewComponent({
      id: 'new-component',
      onClick: () => this.onNewComponentClick()
    });
  }

  mountComponents() {
    // ... existing mounting ...
    
    const container = DOMHelper.findElement('#new-component-container');
    if (container) {
      container.appendChild(this.components.newComponent.getElement());
    }
  }

  onNewComponentClick() {
    // Handle component click
  }
}
```

## Testing Guidelines

### 1. Unit Testing

#### Test Structure
```javascript
class MyComponentTests {
  static runAll() {
    this.testComponentCreation();
    this.testComponentMethods();
    this.testComponentEvents();
    this.testComponentDestruction();
  }

  static testComponentCreation() {
    console.log('Testing component creation...');
    
    // Arrange
    const options = { id: 'test-component', text: 'Test' };
    
    // Act
    const component = new MyComponent(options);
    
    // Assert
    TestUtils.assert(component !== null, 'Component should be created');
    TestUtils.assertEqual(component.getElement().id, 'test-component', 'ID should match');
    
    // Cleanup
    component.destroy();
    
    console.log('Component creation tests passed');
  }
}
```

#### Test Utilities
```javascript
class TestUtils {
  static assert(condition, message) {
    if (!condition) {
      throw new Error(`Assertion failed: ${message}`);
    }
  }

  static assertEqual(actual, expected, message) {
    if (actual !== expected) {
      throw new Error(`Assertion failed: ${message}. Expected: ${expected}, Actual: ${actual}`);
    }
  }

  static createMockElement(tag = 'div', attributes = {}) {
    const element = document.createElement(tag);
    Object.entries(attributes).forEach(([key, value]) => {
      element.setAttribute(key, value);
    });
    return element;
  }

  static triggerEvent(element, eventType, eventData = {}) {
    const event = new Event(eventType, { bubbles: true, ...eventData });
    element.dispatchEvent(event);
  }
}
```

### 2. Integration Testing

#### Backend Integration Tests
```javascript
class BackendIntegrationTests {
  static async runAll() {
    await this.testAnalyzeFunction();
    await this.testFileBrowser();
    await this.testExportFunction();
  }

  static async testAnalyzeFunction() {
    console.log('Testing analyze function...');
    
    try {
      // Mock the Go function for testing
      const mockReport = {
        timestamp: new Date().toISOString(),
        sles_release: 'Test OS',
        problems: [],
        warnings: []
      };
      
      // Test successful analysis
      const result = await mockAnalyze('/path/to/test.txz', false);
      TestUtils.assert(result !== null, 'Analysis should return result');
      
      console.log('Analyze function tests passed');
    } catch (error) {
      console.error('Analyze function tests failed:', error);
      throw error;
    }
  }
}
```

### 3. End-to-End Testing

#### User Workflow Tests
```javascript
class E2ETests {
  static async runAll() {
    await this.testCompleteAnalysisWorkflow();
    await this.testExportWorkflow();
    await this.testKeyboardShortcuts();
  }

  static async testCompleteAnalysisWorkflow() {
    console.log('Testing complete analysis workflow...');
    
    try {
      // Initialize app
      const app = new SSSDInspectorApp();
      
      // Select file
      const fileInput = app.components.fileInput;
      fileInput.setValue('/path/to/test.txz');
      
      // Start analysis
      await app.startAnalysis();
      
      // Verify results
      TestUtils.assert(app.state.currentReport !== null, 'Report should be generated');
      
      // Test export
      await app.exportTXT();
      
      // Cleanup
      app.destroy();
      
      console.log('Complete workflow tests passed');
    } catch (error) {
      console.error('Complete workflow tests failed:', error);
      throw error;
    }
  }
}
```

## Debugging

### 1. Browser Developer Tools

#### Console Debugging
```javascript
// Add debug logging
console.log('Component initialized:', this.id);
console.debug('State updated:', this.state);
console.warn('Deprecated method used:', methodName);
console.error('Error occurred:', error);

// Use console.group for related logs
console.group('File Validation');
console.log('File path:', filePath);
console.log('File size:', fileSize);
console.log('Validation result:', result);
console.groupEnd();
```

#### Breakpoint Debugging
```javascript
// Add debugger statements for step-through debugging
function validateFile(file) {
  debugger; // Execution will pause here
  const validation = FileValidator.validateFile(file);
  return validation;
}
```

### 2. Component Debugging

#### Component State Inspection
```javascript
class DebuggableComponent {
  constructor(options = {}) {
    this.debug = options.debug || false;
    // ... rest of constructor
  }

  log(message, data = null) {
    if (this.debug) {
      console.log(`[${this.constructor.name}] ${message}`, data);
    }
  }

  updateState(newState) {
    this.log('State update', { old: this.state, new: newState });
    this.state = { ...this.state, ...newState };
  }
}
```

### 3. Performance Debugging

#### Performance Monitoring
```javascript
class PerformanceMonitor {
  static startTimer(name) {
    console.time(name);
  }

  static endTimer(name) {
    console.timeEnd(name);
  }

  static measureFunction(name, fn) {
    console.time(name);
    const result = fn();
    console.timeEnd(name);
    return result;
  }

  static async measureAsyncFunction(name, fn) {
    console.time(name);
    const result = await fn();
    console.timeEnd(name);
    return result;
  }
}

// Usage
PerformanceMonitor.startTimer('component-render');
this.render();
PerformanceMonitor.endTimer('component-render');

const result = PerformanceMonitor.measureFunction('validation', () => {
  return FileValidator.validateFile(file);
});
```

## Performance Optimization

### 1. Code Optimization

#### Lazy Loading
```javascript
// Dynamic imports for heavy components
async loadHeavyComponent() {
  if (!this.heavyComponent) {
    const module = await import('./components/HeavyComponent.js');
    this.heavyComponent = new module.HeavyComponent();
  }
  return this.heavyComponent;
}
```

#### Debouncing
```javascript
// Debounce expensive operations
class Debouncer {
  static create(func, delay = 300) {
    let timeoutId;
    return function(...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
  }
}

// Usage
const debouncedValidation = Debouncer.create(this.validateInput, 300);
input.addEventListener('input', debouncedValidation);
```

### 2. Memory Management

#### Event Listener Cleanup
```javascript
class ComponentWithEvents {
  setupEventListeners() {
    this.eventCleanup = [];
    
    const clickCleanup = DOMHelper.addEventListener(
      this.element,
      'click',
      this.handleClick.bind(this)
    );
    this.eventCleanup.push(clickCleanup);
  }

  destroy() {
    // Clean up all event listeners
    this.eventCleanup.forEach(cleanup => cleanup());
    this.eventCleanup = [];
    
    // Remove DOM element
    if (this.element && this.element.parentNode) {
      this.element.parentNode.removeChild(this.element);
    }
  }
}
```

#### Object Reference Management
```javascript
class MemoryManagedComponent {
  constructor() {
    this.cache = new Map();
    this.maxCacheSize = 50;
  }

  addToCache(key, value) {
    if (this.cache.size >= this.maxCacheSize) {
      // Remove oldest entry
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }
    this.cache.set(key, value);
  }

  clearCache() {
    this.cache.clear();
  }
}
```

## Deployment

### 1. Build Process

#### Production Build
```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

#### Build Configuration
```javascript
// vite.config.js
export default {
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: true,
    minify: 'terser',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['src/components/UIComponents.js'],
          utils: ['src/utils/validators.js', 'src/utils/helpers.js']
        }
      }
    }
  }
};
```

### 2. Environment Configuration

#### Environment Variables
```javascript
// src/config/environment.js
export const ENV = {
  development: import.meta.env.DEV,
  production: import.meta.env.PROD,
  version: import.meta.env.VITE_APP_VERSION,
  apiUrl: import.meta.env.VITE_API_URL
};
```

#### Feature Flags
```javascript
// src/config/features.js
export const FEATURES = {
  advancedValidation: import.meta.env.VITE_FEATURE_ADVANCED_VALIDATION === 'true',
  debugMode: import.meta.env.VITE_DEBUG_MODE === 'true',
  experimentalUI: import.meta.env.VITE_FEATURE_EXPERIMENTAL_UI === 'true'
};
```

## Troubleshooting

### Common Issues

#### 1. Module Import Errors
```javascript
// Problem: Cannot find module
// Solution: Check file paths and extensions
import { Component } from './components/Component.js'; // Add .js extension

// Problem: Circular dependencies
// Solution: Refactor to remove circular imports
// Move shared utilities to separate module
```

#### 2. CSS Not Applying
```javascript
// Problem: Styles not loading
// Solution: Check CSS import order
import './styles/layout.css';      // Base styles first
import './styles/components.css';  // Component styles
import './styles/report.css';      // Specific styles last
```

#### 3. Event Listeners Not Working
```javascript
// Problem: Event listeners not firing
// Solution: Ensure elements exist when adding listeners
document.addEventListener('DOMContentLoaded', () => {
  this.setupEventListeners();
});

// Or use event delegation
document.addEventListener('click', (e) => {
  if (e.target.matches('.button')) {
    this.handleButtonClick(e);
  }
});
```

### Debug Checklist

1. **Console Errors**: Check for JavaScript errors
2. **Network Issues**: Verify API calls are working
3. **CSS Issues**: Use browser dev tools to inspect styles
4. **Performance**: Use performance tab to identify bottlenecks
5. **Memory**: Check for memory leaks in heap snapshots

## Contributing

### 1. Code Review Checklist

- [ ] Code follows style guidelines
- [ ] Functions have JSDoc documentation
- [ ] Tests are included and passing
- [ ] No console.log statements left in production code
- [ ] Error handling is implemented
- [ ] Accessibility considerations are addressed
- [ ] Performance impact is considered

### 2. Release Process

1. **Update version** in package.json
2. **Update CHANGELOG.md** with new features
3. **Run full test suite**
4. **Create release tag**
5. **Deploy to production**
6. **Monitor for issues**

This development guide provides comprehensive information for working with the SSSD Inspector frontend. Follow these guidelines to ensure consistent, high-quality code contributions.

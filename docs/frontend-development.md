# Frontend Development Guide

## Getting Started

This guide provides comprehensive instructions for developing and maintaining the SSSD Inspector frontend application.

## Prerequisites

### Required Tools
- **Node.js** (v18 or higher; v24 recommended)
- **npm**
- **Wails CLI** (only to rebuild the Go binary or regenerate `wailsjs/`)

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
```

### 3. Development Server
```bash
npm run dev
```

This will start the Vite development server at `http://localhost:5173`

### 4. Build for Production
```bash
npm run build     # tsc --noEmit && vite build
```

`npm run build` type-checks first, so a type error fails the build instead of
shipping. The built files land in `dist/` and are embedded into the Go binary by
`//go:embed all:frontend/dist` — do not change the output layout
(`index.html` + `assets/`), which `integration_test.go` relies on.

### 5. Test
```bash
npm test          # vitest run (jsdom + Testing Library)
```

`src/tests/` holds the vitest suites: the bridge guard (`backend.test.ts`), the
report view (`ReportView.test.tsx`), the Definitions Studio
(`DefinitionsStudio.test.tsx`), the whole shell including Wails events and OS
drag & drop (`App.test.tsx`) and the archive check (`fileValidation.test.ts`).
Backend calls are mocked through `api/backend.ts` (see `src/tests/helpers.ts`),
so the components are tested against the generated Wails payloads, and the stub
records the event callbacks so a test can fire `analyze-progress`,
`definitions-warning` or a file drop.

### 6. GUI smoke (no desktop session required)
```bash
npm run build
# Produce a REAL report with the engine, then render the bundle in a browser:
sssd-inspector -tags cli ... # or the CLI binary
./sssd-inspector-cli -anonymize -json /path/to/supportconfig /tmp/guismoke
cd frontend && npm run smoke:gui -- --report /tmp/guismoke/supportconfig_report.json
```

`scripts/gui-smoke.mjs` loads the **built** `dist/` bundle in headless Chromium
with a stubbed Wails bridge (`window.go` / `window.runtime`, exactly the surface
`frontend/wailsjs` uses), drives the UI — fill the path, check *Anonymize PII*,
click **Analyze**, switch to the **Definitions Studio**, **Validate**, **Run
dry-run** — then asserts the markers a user would look for and writes a
screenshot per view to `/tmp/sssd-gui-smoke`. It fails on a render crash (the
error boundary text) or on a missing bridge.

It drives three scenes, not one: `analysis` and `studio`, plus `studio-empty`,
which feeds the Studio an inventory whose `files`/`rules`/`diagnostics` are all
`null` — what Go actually sends on a machine with no custom rules, and the state
that crashed the GUI before `listOf()` existed. The render-crash detector is what
turns that scene into a hard failure, so a null array can never come back.

What it covers: the shipped React code, the generated Wails client, the CSS, the
layout, and a report the engine actually produced. What it cannot cover: the
Go ↔ WebKitGTK boundary, which needs a desktop session (`wails build` then run
the binary). Do that pass by hand on a workstation before a release; on a
headless agent WebKitGTK is not even installed.

## Development Workflow

### 1. Feature Development
```bash
# Create feature branch
git checkout -b feature/new-feature-name

# Make changes
# ... develop your feature ...

# Run tests
npm test

# Typecheck (strict); there is no separate linter configured
npm run typecheck

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
npm test              # vitest run (jsdom + Testing Library)
npm run typecheck     # tsc --noEmit, strict
npm run build         # typecheck, then vite build
npm run smoke:gui     # headless GUI smoke on the built bundle (see below)
```

These are the only scripts; there is no coverage, no per-area and no watch
script. For a single suite, pass the path to Vitest
(`npx vitest run src/tests/DefinitionsStudio.test.tsx`) or add `-t '<name>'` to
run one test.

## Architecture Overview

### Module System
The frontend is a React 18 + TypeScript application (`strict` mode):

```typescript
// Components never import the generated bindings directly.
import { analyze, listDefinitions } from './api/backend';
import type { ReportData } from './api/backend';

// State lives in hooks; components stay declarative.
const { report, run } = useAnalysis(status);
return <ReportView report={report} />;
```

Everything is TypeScript: there is no `.js` or `.jsx` file left under `src/`.
The graph renderer that used to be a legacy SVG class is now
`components/report/CorrelationGraph.tsx`, a normal React component rendered by
`CorrelationGraphSection.tsx`; its layout maths was extracted to the pure
`correlationGraph.ts` so it can be unit-tested without a DOM. Types come from
the generated Wails models.

### Component Structure
A component is a function that returns JSX. State, effects and subscriptions go
in a hook or React state, not in a `destroy()` method — there is no manual
teardown, because React owns the lifecycle.

```tsx
// src/components/definitions/DefinitionInventoryPanel.tsx
export interface DefinitionInventoryPanelProps {
  inventory: DefinitionsInventory | null;
  loading: boolean;
  onReload: () => void;
}

export function DefinitionInventoryPanel({ inventory, loading, onReload }: DefinitionInventoryPanelProps) {
  const rules = listOf(inventory?.rules);
  return <section className="studio-panel">{/* ... */}</section>;
}

export default DefinitionInventoryPanel;
```

Conventions worth keeping:

- Named export for the component, default export as well, so both import styles
  resolve.
- Props are a plain interface; behaviour arrives as callbacks (`onReload`),
  never as an import of the backend module. The component must stay testable
  without the Wails bridge.
- Render `null` fields through `listOf()` from `utils/payload`. Go can send
  `null` for a slice, and `listOf` is what stopped the GUI from crashing on a
  machine with no custom rules.

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

#### Step 1: Write the component
```tsx
// src/components/NewComponent.tsx
export interface NewComponentProps {
  title: string;
  onSelect?: () => void;
}

export function NewComponent({ title, onSelect }: NewComponentProps) {
  return (
    <div className="new-component" onClick={onSelect}>
      {title}
    </div>
  );
}

export default NewComponent;
```

No class, no constructor, no `destroy()`: React mounts and unmounts it. Put
anything stateful in `useState`/`useEffect`, and keep backend calls out of the
component — pass a callback in and let the caller decide.

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
```tsx
// src/tests/NewComponent.test.tsx
import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { NewComponent } from '../components/NewComponent';

describe('NewComponent', () => {
  it('shows its title', () => {
    render(<NewComponent title="Definitions" />);
    expect(screen.getByText('Definitions')).toBeInTheDocument();
  });

  it('reports a selection', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<NewComponent title="Definitions" onSelect={onSelect} />);
    await user.click(screen.getByText('Definitions'));
    expect(onSelect).toHaveBeenCalled();
  });
});
```

Add the suite to the gate: a component without a test is unfinished, and a
test that has never been seen to fail has not been shown to work.

#### Step 4: Update Documentation
```markdown
## NewComponent

Renders [purpose].

### Usage
```tsx
<NewComponent title="Definitions" onSelect={handleSelect} />
```

### Props
- `title` (string): what to show
- `onSelect` (function, optional): called on click
```

### 2. Component Integration

#### Add to the Shell
```tsx
// A component is just a function: render it where it belongs and give it the
// state it needs through props or a hook. No registration step exists.
import { NewComponent } from './components/NewComponent';

function AnalysisView() {
  const { report } = useAnalysis(status);
  return (
    <div className="pdf-content-area">
      {report ? <ReportView report={report} /> : <NewComponent onReady={() => undefined} />}
    </div>
  );
}
```

There is no legacy imperative layer left: `UIComponents.js` and `DOMHelper.js`
were deleted, and every component is mounted by rendering it. See
`CorrelationGraphSection.tsx` for the pattern.

## Testing Guidelines

Tests are **Vitest** with **React Testing Library**, in `src/tests/`, run by
`npm test`. There is no browser test-runner and no `TestUtils` helper: RTL
queries the DOM the way a user reads it, so a test that cannot find a control
by role is telling you the control is unreachable, not that the query is wrong.

### 1. Unit and Component Testing

#### Test structure

One suite per unit, in `src/tests/`, named after the unit under test:

```tsx
// src/tests/DefinitionsStudio.test.tsx
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockBackend, makeInventory } from './helpers';

const backend = mockBackend();
vi.mock('../api/backend', () => backend);

const { DefinitionsStudio } = await import('../components/definitions/DefinitionsStudio');

beforeEach(() => {
  // Re-seed every mock: a mockResolvedValue set by a previous test would
  // otherwise leak into the next one.
  Object.assign(backend, mockBackend());
});

describe('DefinitionsStudio — inventory', () => {
  it('re-reads the inventory on demand', async () => {
    const user = userEvent.setup();
    render(<DefinitionsStudio status={status} />);
    await user.click(screen.getByRole('button', { name: 'Reload' }));
    await waitFor(() => expect(backend.listDefinitions).toHaveBeenCalledTimes(2));
  });
});
```

Rules that keep the suite honest:

- **Query by role, label or text**, never by test id, unless the element has no
  accessible identity. `getByRole('button', { name: 'Reload' })` also asserts
  the control is reachable and named, which a `getElementById` never did.
- **Drive it like a user**: `userEvent.click`, then `await` the outcome. Never
  assert synchronously on a promise the component has not resolved yet.
- **Assert what the user must not be misled about.** A refused save has to read
  as refused; a reload that fails must not blank the panel, because "no rules
  loaded" would otherwise be read as fact.
- **Keep each `it` inside its `describe`.** A stray `it` after a closing `});`
  still runs, but it escapes the group and hides which area regressed.
- **Prove the test bites.** A green test that passes against broken code is
  worse than no test: flip the behaviour, watch it go red, flip it back.

#### Fixtures

`src/tests/helpers.ts` holds the builders (`makeInventory`, `makeCatalog`,
`makeValidation`, `makeDryRun`, `makeReport`, `mockBackend`) so no suite
hand-rolls a payload. Two things it provides matter:

- `mockBackend()` returns benign defaults for every call, so a test only
  overrides the one method it is about.
- `mockBackend().events` records the Wails callbacks the shell registered, so a
  test can fire `analyze-progress`, `definitions-warning` or an OS file drop
  without a real backend.

### 2. Backend Integration Tests

`api/backend.ts` is mocked wholesale, so the suites assert against the
**generated Wails payloads**, not against a hand-written imitation of them.
A payload shape change breaks the suites loudly, which is the point.

Anything that only the Go side can decide is covered by Go tests next to the
code it owns: `openDefinitionsRoot` has its own in `app_definitions_test.go`
(including the case where the folder does not exist yet and must be created).

### 3. End-to-End Testing

Covered by `npm run smoke:gui` (`scripts/gui-smoke.mjs`), which loads the
**built** `dist/` bundle in headless Chromium with a stubbed Wails bridge and
drives the real UI. What it cannot cover is the Go ↔ WebKitGTK boundary,
which needs a desktop session; run that pass by hand before a release. See
"GUI smoke" above for the exact commands.

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
React owns the lifecycle: an effect that subscribes returns its own cleanup,
and React calls it on unmount. There is no `destroy()` to remember to invoke.

```tsx
useEffect(() => {
  const onProgress = (message: string, percentage: number) => setProgress({ message, percentage });
  return backend.onAnalyzeProgress(onProgress);
}, []);
```

For a window or document listener added outside React, clean it up in the same
effect:

```tsx
useEffect(() => {
  const onDrop = (e: DragEvent) => { e.preventDefault(); void onPaths([...e.dataTransfer.files].map((f) => f.name)); };
  window.addEventListener('drop', onDrop);
  return () => window.removeEventListener('drop', onDrop);
}, [onPaths]);
```

Worth stating plainly: a missing cleanup leaks a listener, and a leaked
listener keeps a stale closure alive, so the bug shows up later and somewhere
else. That is why `App.test.tsx` asserts the shell unsubscribes.

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
```typescript
// vite.config.ts
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',      // embedded by //go:embed all:frontend/dist
    assetsDir: 'assets', // integration_test.go looks for dist/assets/*.js
    sourcemap: false,    // do not ship source maps in the embedded binary
  },
});
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

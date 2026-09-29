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
│   │   ├── definitions/          # Definitions Studio panels
│   │   └── CorrelationGraph.js   # the one imperative renderer, mounted via ref
│   ├── config/                   # ui.ts — strings, zoom bounds, archive formats
│   ├── utils/                    # fileValidation.ts (the one input check)
│   ├── styles/                   # layout.css, components.css, report.css, studio.css
│   └── tests/                    # vitest suites (see Testing Framework)
├── wailsjs/                      # Generated Wails bindings — DO NOT EDIT
├── tsconfig.json
├── vite.config.ts
└── package.json
```

`wailsjs/` is regenerated with `wails generate module` after any Go signature
change; `api/backend.ts` is the only place that imports it, so the rest of the
UI depends on its types and never on the raw bindings.

## Core Modules

### UI Constants

#### config/ui.ts
The user-visible strings, the zoom bounds and the accepted archive formats.
Deliberately tiny: the legacy `config/constants.js` held 320 lines of which the
UI used nine, plus a `FrontendConfig` class that nothing read. A value used once
belongs next to its component.

```typescript
export const UI = {
  APP_TITLE: 'SSSD Supportconfig Analyzer',
  BUTTONS: { BROWSE: 'Browse...', ANALYZE: 'Analyze', /* ... */ },
  PLACEHOLDERS: { FILE_PATH: 'Select or paste path to supportconfig.txz...' },
  TOOLTIPS: { ANONYMIZE: 'Redacts IP Addresses and Domain Names from the report' },
} as const;

export const ZOOM = { DEFAULT: 1.0, STEP: 0.1, MIN: 0.8, MAX: 1.5 } as const;

// Must stay in step with constants.PatternSupportconfig in Go.
export const SUPPORTED_EXTENSIONS = ['.txz', '.tar.xz'] as const;
```

### Component Tree

The UI is React: components are functions, state lives in hooks, and the tree
follows the two views of the shell.

```
App                          shell: top bar, view switch, status, shortcuts
├── components/shell/        TopBar, ProgressPanel, StatusBanner
├── components/report/       ReportView + one component per report section
│   ├── ExecSummary, SystemInfoTable, IniSnippet
│   ├── DiagnosticsList, ConfigFindingsList
│   ├── LogErrorsList, TimelineList, KbArticlesList
│   ├── TemporalClustersList, KbSuggestionsList
│   └── CorrelationGraphSection   (mounts CorrelationGraph.js through a ref)
├── components/definitions/  Studio panels (inventory, editor, dry-run, catalog)
└── components/common/       Section, ErrorBoundary
```

`Section` is the collapsible `<details>/<summary>` block every report section
uses: it sets the initial `open` state imperatively so the user keeps control of
the toggle afterwards.

`CorrelationGraph.js` is the one imperative survivor — a self-contained
force-directed SVG renderer with no dependencies. It is mounted into a host
`div` from an effect, not rewritten.

`ErrorBoundary` wraps the whole app: a rendering bug must not leave the user
with a blank window and a dead Analyze button.

### Utility Modules

#### utils/fileValidation.ts
The one input check the analyser needs: is this path something the analysis can
open? It replaces `utils/validators.js`, whose `FileValidator` took the
substring from the LAST dot and compared it against `['.txz', '.tar.xz']` — so
`supportconfig.tar.xz` reduced to `.xz` and the GUI refused the most common
supportconfig filename its own file dialog offers. Matching the longest
supported suffix fixed it; an already-extracted supportconfig directory is
accepted too, because the analysis supports directories.

```typescript
export function validateArchivePath(path: string): FileValidationResult {
  // supported archive suffix, or no extension at all (a directory)
}
```

The rest of the old validator surface (file size, form validation, DOM helpers,
formatting helpers) had no callers: React does not need them, and the components
that did are gone.

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

State lives in hooks, not in a global object. Each hook owns one concern and is
the only place that talks to the backend for it.

| Hook | Owns |
|------|------|
| `useStatus` | the single transient message (auto-dismissed unless it is an error) |
| `useTheme` | dark/light mode, persisted under `sssd-inspector-theme` |
| `useAnalysis` | path, anonymize flag, busy state, progress, report, zoom, exports |
| `useDefinitionsInventory` | the discovery inventory and the catalog description |
| `useRuleEditor` | the rules document, its verdict and the save outcome |
| `useRuleDryRun` | the dry-run target and its outcomes |

### State Persistence
- **localStorage**: the theme choice only (`sssd-inspector-theme`)
- **nothing else**: an analysis report is not cached; re-running is cheap and a
  stale report is worse than no report

## Event System

### Backend Events
```typescript
// Subscribed in an effect, unsubscribed on unmount. api/backend.ts guards the
// call: with no Go bridge (npm run dev in a browser) it returns a no-op.
useEffect(() => backend.onAnalyzeProgress((message, percentage) => {
  // Ignored when no analysis is in flight: Wails can deliver the final tick
  // after the promise resolved, and a stuck bar over a finished report looks
  // like a hung application.
  setProgress({ message, percentage });
}), []);
```

- `analyze-progress` — `(message, percentage)`; drives the progress panel and
  the Analyze button label
- `definitions-warning` — `(diagnostics)`; a rule or KB file was skipped, so the
  status banner says so (M0 gate)
- `OnFileDrop` — the native listener is registered once (the runtime has no
  per-callback unsubscribe) and delegates to the handler stored in `api/backend`

### Keyboard Shortcuts
`Ctrl+O` browse, `Ctrl+Enter` analyze, `Ctrl+P/S/J` export, `Ctrl+1/2` switch
view. Bound on `document` in one effect, guarded on `analysis.report` so an
export shortcut cannot fire on an empty window.

## Testing Framework

### Test Structure
```
src/tests/
├── setup.ts                 jsdom shims (matchMedia, print, cleanup)
├── helpers.ts               report/inventory fixtures + the backend stub
├── backend.test.ts          the bridge guard: no Go bridge = explained, not fatal
├── ReportView.test.tsx      every report section renders what Go produced
├── DefinitionsStudio.test.tsx  inventory, editor, validate, save refusal, dry-run
├── App.test.tsx             the shell: analysis flow, exports, views, Wails events
└── fileValidation.test.ts   the archive check, including the .tar.xz regression
```

### What Is Tested
Unit tests cover the pure functions (`validateArchivePath`), the report
rendering against fixtures that mirror the generated Wails models, and the
Studio flows that must not lie: a refused save is never rendered as a success, a
skipped catalog override is visible, a late progress event cannot resurrect the
progress bar.

Integration tests mount the real `App` with a stubbed `api/backend.ts`. The
stub records the Wails event callbacks the shell registers, so a test can fire
`analyze-progress`, `definitions-warning` and an OS file drop and assert what
the user sees — the behaviour the legacy browser test-runner exercised with
hand-written fake runtime objects, now against the components that ship.

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
> `npm test` runs the vitest suites.
>
> **Legacy JavaScript**: also gone (M3). `config/constants.js`,
> `config/FrontendConfig.js`, `utils/validators.js`, `utils/helpers.js`,
> `components/UIComponents.js` and the browser test-runner were deleted; the
> behaviour that was actually used moved to `config/ui.ts` and
> `utils/fileValidation.ts`, and the drag & drop / Wails event coverage moved
> into `App.test.tsx`. The only remaining JavaScript is
> `components/CorrelationGraph.js`, mounted through a ref.

### Technical Debt
- **Legacy code removal**: Clean up deprecated files
- **Performance monitoring**: Add performance metrics
- **Error tracking**: Implement error reporting
- **Automated testing**: Expand test coverage

## Conclusion

The refactored SSSD Inspector frontend provides a solid foundation for future development with improved maintainability, testability, and user experience. The modular architecture allows for easy extension and modification while following modern web development best practices.

The comprehensive testing framework ensures code quality and reliability, while the documentation system provides clear guidance for developers working on the project.

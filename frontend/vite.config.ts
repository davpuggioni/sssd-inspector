import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Vite configuration for the Wails frontend.
//
// The build output (dist/) is embedded into the Go binary via
// `//go:embed all:frontend/dist` in main_gui.go, so the output layout
// (index.html + assets/) is part of the application contract:
// integration_test.go discovers the hashed JS bundle inside dist/assets.
// Wails builds the frontend with `npm run build`; vitest runs the component
// tests. The `test` block is read only by vitest and has no effect on the
// production bundle, so the test files never reach dist/.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    // Wails serves the bundle from the desktop webview: source maps are
    // useful while developing but must not ship in the embedded binary.
    sourcemap: false,
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/tests/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
  },
});

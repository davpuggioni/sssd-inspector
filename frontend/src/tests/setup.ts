// setup.ts — vitest environment shims.
//
// jsdom does not implement the Wails runtime or the browser APIs the shell
// touches (matchMedia, print). Stubbing them here keeps every test focused on
// component behaviour instead of environment plumbing; backend calls are mocked
// per-test with vi.mock('../api/backend').
import '@testing-library/jest-dom/vitest';
import { afterEach, vi } from 'vitest';
import { cleanup } from '@testing-library/react';

afterEach(() => {
  cleanup();
});

// matchMedia: jsdom has no implementation; the theme hook queries it.
if (typeof window.matchMedia !== 'function') {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    }),
  });
}

// print: the PDF export path calls window.print(), which jsdom lacks.
window.print = vi.fn();

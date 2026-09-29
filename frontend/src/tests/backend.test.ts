// backend.test.ts — the guard that makes the Wails bridge optional.
//
// The UI must render (and fail with an actionable message, not a TypeError on
// window.go) when there is no Go backend, e.g. `npm run dev` in a plain browser.
import { afterEach, describe, expect, it } from 'vitest';
import {
  analyze,
  backendAvailable,
  listDefinitions,
  onAnalyzeProgress,
  onFileDrop,
  readRuleYaml,
} from '../api/backend';

declare global {
  interface Window {
    go?: unknown;
  }
}

afterEach(() => {
  delete window.go;
});


afterEach(() => {
  delete window.go;
});

describe('backendAvailable', () => {
  it('is false without a bridge', () => {
    expect(backendAvailable()).toBe(false);
  });

  it('is false for a partial bridge', () => {
    window.go = {};
    expect(backendAvailable()).toBe(false);
    window.go = { main: {} };
    expect(backendAvailable()).toBe(false);
  });

  it('is true once the App object is attached', () => {
    window.go = { main: { App: { Analyze: () => undefined } } };
    expect(backendAvailable()).toBe(true);
  });
});

describe('calls without a bridge', () => {
  // Every binding is async, so the guard surfaces as a rejected promise: a
  // component's try/catch handles it, and the message is what the user sees.
  it('explain themselves instead of throwing a TypeError', async () => {
    await expect(analyze('/tmp/supportconfig.txz', false)).rejects.toThrow(/backend is not attached/);
    await expect(listDefinitions()).rejects.toThrow(/backend is not attached/);
    await expect(readRuleYaml('user')).rejects.toThrow(/backend is not attached/);
  });

  it('return inert unsubscribe functions and skip native listeners', () => {
    const off = onAnalyzeProgress(() => undefined);
    expect(typeof off).toBe('function');
    expect(() => off()).not.toThrow();
    expect(onFileDrop(() => undefined)).toBe(false);
  });
});


// useTheme — dark/light mode with the same persistence contract as the legacy
// frontend: localStorage key 'sssd-inspector-theme' wins, otherwise the system
// preference is followed live.
import { useCallback, useEffect, useState } from 'react';

export type Theme = 'light' | 'dark';

const STORAGE_KEY = 'sssd-inspector-theme';

function systemPrefersDark(): boolean {
  return typeof window !== 'undefined' && typeof window.matchMedia === 'function'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
    : false;
}

function storedTheme(): Theme | null {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return saved === 'dark' || saved === 'light' ? saved : null;
  } catch {
    // Private mode / disabled storage: fall back to the system preference.
    return null;
  }
}

function preferredTheme(): Theme {
  return storedTheme() ?? (systemPrefersDark() ? 'dark' : 'light');
}

/** Applies the theme class to <html>. Exported so the entry point can run it
 * before the first paint and avoid a flash of the wrong theme. */
export function applyThemeClass(theme: Theme): void {
  document.documentElement.classList.toggle('dark-mode', theme === 'dark');
}

/** Called once by main.tsx, synchronously, before React renders. */
export function bootstrapTheme(): Theme {
  const theme = preferredTheme();
  applyThemeClass(theme);
  return theme;
}

export interface ThemeApi {
  theme: Theme;
  toggle: () => void;
}

export function useTheme(initial: Theme = preferredTheme()): ThemeApi {
  const [theme, setTheme] = useState<Theme>(initial);

  useEffect(() => {
    applyThemeClass(theme);
  }, [theme]);

  // Follow live system changes as long as the user made no manual choice.
  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return undefined;
    }
    const query = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = (event: MediaQueryListEvent) => {
      if (storedTheme() === null) {
        setTheme(event.matches ? 'dark' : 'light');
      }
    };
    query.addEventListener('change', onChange);
    return () => query.removeEventListener('change', onChange);
  }, []);

  const toggle = useCallback(() => {
    setTheme((current) => {
      const next: Theme = current === 'dark' ? 'light' : 'dark';
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        // Storage unavailable: the choice is session-only, not fatal.
      }
      return next;
    });
  }, []);

  return { theme, toggle };
}

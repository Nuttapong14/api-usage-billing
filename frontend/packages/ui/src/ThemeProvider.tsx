'use client';

import React, { createContext, useContext, useEffect, useState, useCallback, useMemo, ReactNode } from 'react';

// =====================================
// Theme Types
// =====================================

export type Theme = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';

interface ThemeColors {
  primary: string;
  secondary: string;
  accent: string;
  background: string;
  foreground: string;
  muted: string;
  mutedForeground: string;
  border: string;
  input: string;
  ring: string;
  destructive: string;
  destructiveForeground: string;
  success: string;
  successForeground: string;
  warning: string;
  warningForeground: string;
}

interface ThemeConfig {
  colors: ThemeColors;
  borderRadius: string;
  fontFamily: string;
}

// =====================================
// Default Theme Configurations
// =====================================

const lightTheme: ThemeConfig = {
  colors: {
    primary: '#2563eb',
    secondary: '#64748b',
    accent: '#8b5cf6',
    background: '#ffffff',
    foreground: '#0f172a',
    muted: '#f1f5f9',
    mutedForeground: '#64748b',
    border: '#e2e8f0',
    input: '#e2e8f0',
    ring: '#2563eb',
    destructive: '#ef4444',
    destructiveForeground: '#ffffff',
    success: '#22c55e',
    successForeground: '#ffffff',
    warning: '#f59e0b',
    warningForeground: '#ffffff',
  },
  borderRadius: '0.5rem',
  fontFamily: 'Inter, system-ui, sans-serif',
};

const darkTheme: ThemeConfig = {
  colors: {
    primary: '#3b82f6',
    secondary: '#94a3b8',
    accent: '#a78bfa',
    background: '#0f172a',
    foreground: '#f8fafc',
    muted: '#1e293b',
    mutedForeground: '#94a3b8',
    border: '#334155',
    input: '#334155',
    ring: '#3b82f6',
    destructive: '#f87171',
    destructiveForeground: '#0f172a',
    success: '#4ade80',
    successForeground: '#0f172a',
    warning: '#fbbf24',
    warningForeground: '#0f172a',
  },
  borderRadius: '0.5rem',
  fontFamily: 'Inter, system-ui, sans-serif',
};

// =====================================
// Theme Context
// =====================================

interface ThemeContextValue {
  theme: Theme;
  resolvedTheme: ResolvedTheme;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
  themeConfig: ThemeConfig;
  isDark: boolean;
  isLight: boolean;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

// =====================================
// Theme Provider
// =====================================

interface ThemeProviderProps {
  children: ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
  attribute?: 'class' | 'data-theme';
  enableSystem?: boolean;
  disableTransitionOnChange?: boolean;
  themes?: {
    light?: Partial<ThemeConfig>;
    dark?: Partial<ThemeConfig>;
  };
}

function camelToKebab(str: string): string {
  return str.replace(/([A-Z])/g, '-$1').toLowerCase();
}

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = 'ui-theme',
  attribute = 'class',
  enableSystem = true,
  disableTransitionOnChange = false,
  themes,
}: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(storageKey);
      if (stored === 'light' || stored === 'dark' || stored === 'system') {
        return stored;
      }
    }
    return defaultTheme;
  });

  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>('light');

  // Merge custom themes with defaults
  const mergedLightTheme = useMemo(
    () => ({
      ...lightTheme,
      ...themes?.light,
      colors: { ...lightTheme.colors, ...themes?.light?.colors },
    }),
    [themes?.light]
  );

  const mergedDarkTheme = useMemo(
    () => ({
      ...darkTheme,
      ...themes?.dark,
      colors: { ...darkTheme.colors, ...themes?.dark?.colors },
    }),
    [themes?.dark]
  );

  const themeConfig = useMemo(
    () => (resolvedTheme === 'dark' ? mergedDarkTheme : mergedLightTheme),
    [resolvedTheme, mergedDarkTheme, mergedLightTheme]
  );

  // Get system theme
  const getSystemTheme = useCallback((): ResolvedTheme => {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }, []);

  // Resolve theme
  useEffect(() => {
    const resolved = theme === 'system' ? getSystemTheme() : theme;
    setResolvedTheme(resolved);
  }, [theme, getSystemTheme]);

  // Apply theme to document
  useEffect(() => {
    const root = document.documentElement;

    if (disableTransitionOnChange) {
      root.style.setProperty('transition', 'none');
    }

    if (attribute === 'class') {
      root.classList.remove('light', 'dark');
      root.classList.add(resolvedTheme);
    } else {
      root.setAttribute('data-theme', resolvedTheme);
    }

    // Apply CSS custom properties
    Object.entries(themeConfig.colors).forEach(([key, value]) => {
      const cssVarName = `--color-${camelToKebab(key)}`;
      root.style.setProperty(cssVarName, value);
    });

    root.style.setProperty('--border-radius', themeConfig.borderRadius);
    root.style.setProperty('--font-family', themeConfig.fontFamily);

    if (disableTransitionOnChange) {
      // Force reflow
      void root.offsetHeight;
      root.style.removeProperty('transition');
    }
  }, [resolvedTheme, themeConfig, attribute, disableTransitionOnChange]);

  // Listen for system theme changes
  useEffect(() => {
    if (!enableSystem || theme !== 'system') return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    const handleChange = (e: MediaQueryListEvent) => {
      setResolvedTheme(e.matches ? 'dark' : 'light');
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [enableSystem, theme]);

  const setTheme = useCallback(
    (newTheme: Theme) => {
      setThemeState(newTheme);
      if (typeof window !== 'undefined') {
        localStorage.setItem(storageKey, newTheme);
      }
    },
    [storageKey]
  );

  const toggleTheme = useCallback(() => {
    setTheme(resolvedTheme === 'light' ? 'dark' : 'light');
  }, [resolvedTheme, setTheme]);

  const value = useMemo<ThemeContextValue>(
    () => ({
      theme,
      resolvedTheme,
      setTheme,
      toggleTheme,
      themeConfig,
      isDark: resolvedTheme === 'dark',
      isLight: resolvedTheme === 'light',
    }),
    [theme, resolvedTheme, setTheme, toggleTheme, themeConfig]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

// =====================================
// Hooks
// =====================================

export function useTheme(): ThemeContextValue {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}

export function useResolvedTheme(): ResolvedTheme {
  const { resolvedTheme } = useTheme();
  return resolvedTheme;
}

export function useThemeConfig(): ThemeConfig {
  const { themeConfig } = useTheme();
  return themeConfig;
}

// =====================================
// Theme Toggle Component
// =====================================

interface ThemeToggleProps {
  className?: string;
  showLabel?: boolean;
}

export function ThemeToggle({ className, showLabel = false }: ThemeToggleProps) {
  const { setTheme, isDark } = useTheme();

  return (
    <button
      onClick={() => setTheme(isDark ? 'light' : 'dark')}
      className={className}
      aria-label={`Switch to ${isDark ? 'light' : 'dark'} theme`}
    >
      {isDark ? (
        <svg
          className="h-5 w-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"
          />
        </svg>
      ) : (
        <svg
          className="h-5 w-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"
          />
        </svg>
      )}
      {showLabel && (
        <span className="ml-2">{isDark ? 'Light' : 'Dark'}</span>
      )}
    </button>
  );
}

// =====================================
// Theme Selector Component
// =====================================

interface ThemeSelectorProps {
  className?: string;
}

export function ThemeSelector({ className }: ThemeSelectorProps) {
  const { theme, setTheme } = useTheme();

  return (
    <select
      value={theme}
      onChange={(e) => setTheme(e.target.value as Theme)}
      className={className}
      aria-label="Select theme"
    >
      <option value="light">Light</option>
      <option value="dark">Dark</option>
      <option value="system">System</option>
    </select>
  );
}

// =====================================
// Script for preventing FOUC
// =====================================

interface ThemeScriptProps {
  storageKey?: string;
  defaultTheme?: Theme;
  attribute?: 'class' | 'data-theme';
}

export function ThemeScript({
  storageKey = 'ui-theme',
  defaultTheme = 'system',
  attribute = 'class',
}: ThemeScriptProps) {
  const classScript = "root.classList.remove('light', 'dark'); root.classList.add(resolved);";
  const attrScript = "root.setAttribute('data-theme', resolved);";

  const script = `
    (function() {
      try {
        var theme = localStorage.getItem('${storageKey}') || '${defaultTheme}';
        var resolved = theme;

        if (theme === 'system') {
          resolved = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
        }

        var root = document.documentElement;
        ${attribute === 'class' ? classScript : attrScript}
      } catch (e) {}
    })();
  `;

  return <script dangerouslySetInnerHTML={{ __html: script }} />;
}

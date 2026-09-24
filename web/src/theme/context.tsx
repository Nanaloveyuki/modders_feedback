import { createContext, useContext, useEffect, useMemo, useState, type CSSProperties, type ReactNode } from 'react';
import { paletteFor, readStoredDark, type Palette, type ThemeName } from './theme';

type Theme = {
  dark: boolean;
  theme: ThemeName;
  palette: Palette;
  toggle: () => void;
};

const ThemeContext = createContext<Theme | null>(null);

export function themeStyle(palette: Palette): CSSProperties {
  return {
    ['--canvas' as string]: palette.canvas,
    ['--wash' as string]: palette.wash,
    ['--bar' as string]: palette.bar,
    ['--surface' as string]: palette.surface,
    ['--surface-raised' as string]: palette.raised,
    ['--field' as string]: palette.field,
    ['--line' as string]: palette.line,
    ['--line-strong' as string]: palette.lineStrong,
    ['--text' as string]: palette.text,
    ['--muted' as string]: palette.muted,
    ['--faint' as string]: palette.faint,
    ['--gold' as string]: palette.gold,
    ['--gold-ink' as string]: palette.goldInk,
    ['--accent' as string]: palette.accent,
    ['--accent-ink' as string]: palette.accentInk,
    ['--hover' as string]: palette.hover,
    ['--active' as string]: palette.active,
    ['--active-ink' as string]: palette.activeInk,
    ['--bug' as string]: palette.bug,
    ['--feature' as string]: palette.feature,
    ['--question' as string]: palette.question,
    ['--open' as string]: palette.open,
    ['--progress' as string]: palette.progress,
    ['--resolved' as string]: palette.resolved,
    ['--closed' as string]: palette.closed,
    ['--danger' as string]: palette.danger,
    ['--scrim' as string]: palette.scrim,
    ['--shadow' as string]: palette.shadow,
    ['--green' as string]: palette.gold,
    ['--orange' as string]: palette.open,
    ['--red' as string]: palette.danger,
    ['--mono' as string]: "'DM Mono', monospace",
  };
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [dark, setDark] = useState(readStoredDark);
  const palette = useMemo(() => paletteFor(dark), [dark]);

  useEffect(() => {
    const theme: ThemeName = dark ? 'dark' : 'light';
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.colorScheme = theme;
    localStorage.setItem('feedback-theme', theme);
    document.querySelector('meta[name="theme-color"]')?.setAttribute('content', palette.canvas);
  }, [dark, palette.canvas]);

  const value = useMemo<Theme>(() => ({
    dark,
    theme: dark ? 'dark' : 'light',
    palette,
    toggle: () => setDark((current) => !current),
  }), [dark, palette]);

  return (
    <ThemeContext.Provider value={value}>
      <div className="app-shell" data-dark={dark ? 'true' : 'false'} style={themeStyle(palette)}>
        {children}
      </div>
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const value = useContext(ThemeContext);
  if (!value) throw new Error('useTheme must be used within ThemeProvider');
  return value;
}

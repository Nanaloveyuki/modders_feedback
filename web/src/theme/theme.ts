export type ThemeName = 'light' | 'dark';

export type Palette = {
  dark: boolean;
  canvas: string;
  wash: string;
  bar: string;
  surface: string;
  raised: string;
  field: string;
  line: string;
  lineStrong: string;
  text: string;
  muted: string;
  faint: string;
  gold: string;
  goldInk: string;
  accent: string;
  accentInk: string;
  hover: string;
  active: string;
  activeInk: string;
  bug: string;
  feature: string;
  question: string;
  open: string;
  progress: string;
  resolved: string;
  closed: string;
  danger: string;
  scrim: string;
  shadow: string;
};

const light: Palette = {
  dark: false,
  canvas: '#ffffff',
  wash: 'rgba(232, 248, 255, .55)',
  bar: '#ffffff',
  surface: '#ffffff',
  raised: '#ffffff',
  field: '#ffffff',
  line: '#e4e8ee',
  lineStrong: '#cfd6e0',
  text: '#243044',
  muted: '#5c6b80',
  faint: '#8b97a8',
  gold: '#3d7ea6',
  goldInk: '#1d4d68',
  accent: '#e8f8ff',
  accentInk: '#1d4d68',
  hover: '#fffbe8',
  active: '#e8f8ff',
  activeInk: '#1d4d68',
  bug: '#c45b70',
  feature: '#b8892e',
  question: '#3d7ea6',
  open: '#b8892e',
  progress: '#3d7ea6',
  resolved: '#3f8f6b',
  closed: '#7d8794',
  danger: '#c45b70',
  scrim: 'rgba(36, 48, 68, .36)',
  shadow: 'rgba(40, 48, 72, .10)',
};

const dark: Palette = {
  dark: true,
  canvas: '#151515',
  wash: 'rgba(213, 173, 98, .08)',
  bar: 'rgba(21, 21, 21, .96)',
  surface: '#1d1d1b',
  raised: '#272622',
  field: '#151915',
  line: '#3a3934',
  lineStrong: '#514a3d',
  text: '#e8e3d7',
  muted: '#aaa292',
  faint: '#777267',
  gold: '#d5ad62',
  goldInk: '#292116',
  accent: '#d8b66f',
  accentInk: '#1c1812',
  hover: '#24231f',
  active: '#2c2a24',
  activeInk: '#f0e6d0',
  bug: '#dc8870',
  feature: '#d5ad62',
  question: '#8db6bd',
  open: '#d5ae76',
  progress: '#92b9b9',
  resolved: '#c4b07a',
  closed: '#92978c',
  danger: '#dc8870',
  scrim: 'rgba(6, 8, 6, .77)',
  shadow: 'rgba(0, 0, 0, .32)',
};

export function paletteFor(darkMode: boolean): Palette {
  return darkMode ? dark : light;
}

export function readStoredDark(): boolean {
  const saved = localStorage.getItem('feedback-theme');
  if (saved === 'light') return false;
  if (saved === 'dark') return true;
  return true;
}

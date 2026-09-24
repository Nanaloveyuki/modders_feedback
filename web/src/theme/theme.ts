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
  canvas: '#f3eee2',
  wash: 'rgba(184, 138, 59, .13)',
  bar: 'rgba(248, 244, 235, .96)',
  surface: '#fbf8f1',
  raised: '#eee7d8',
  field: '#fbf8f1',
  line: '#d8cfbd',
  lineStrong: '#cdbb96',
  text: '#302b20',
  muted: '#716957',
  faint: '#9b927f',
  gold: '#b88a3b',
  goldInk: '#59451f',
  accent: '#b88a3b',
  accentInk: '#2c2414',
  hover: '#f4ecdc',
  active: '#e9ddc3',
  activeInk: '#4a3919',
  bug: '#b65f4d',
  feature: '#b88a3b',
  question: '#5d8490',
  open: '#b97843',
  progress: '#5d8490',
  resolved: '#6d8a4e',
  closed: '#716957',
  danger: '#b65f4d',
  scrim: 'rgba(51, 45, 34, .4)',
  shadow: 'rgba(48, 43, 32, .16)',
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
  accentInk: '#292116',
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

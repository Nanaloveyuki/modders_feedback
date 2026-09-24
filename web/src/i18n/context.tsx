import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import { locales, translate, type Locale, type MessageKey } from './messages';

const storageKey = 'feedback-locale';

type I18n = {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: MessageKey, values?: Record<string, string | number>) => string;
};

const I18nContext = createContext<I18n | null>(null);

function storedLocale(): Locale {
  const saved = localStorage.getItem(storageKey);
  if (saved === 'zh' || saved === 'en') return saved;
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh' : 'en';
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>(storedLocale);

  useEffect(() => {
    localStorage.setItem(storageKey, locale);
    document.documentElement.lang = locale === 'zh' ? 'zh-CN' : 'en';
    document.title = translate(locale, 'documentTitle');
    document.querySelector('meta[name="description"]')?.setAttribute('content', translate(locale, 'documentDescription'));
  }, [locale]);

  const value = useMemo<I18n>(() => ({
    locale,
    setLocale,
    t: (key, values) => translate(locale, key, values),
  }), [locale]);

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const value = useContext(I18nContext);
  if (!value) throw new Error('useI18n must be used within I18nProvider');
  return value;
}

export function localeTag(locale: Locale) {
  return locale === 'zh' ? 'zh-CN' : 'en';
}

export function nextLocale(locale: Locale): Locale {
  const index = locales.indexOf(locale);
  return locales[(index + 1) % locales.length];
}

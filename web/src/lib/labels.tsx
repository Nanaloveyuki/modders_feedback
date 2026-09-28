import type { ReactNode } from 'react';
import { AlertTriangle, CircleHelp, Sparkles } from 'lucide-react';
import { localeTag, useI18n } from '../i18n/context';
import type { Category, Status, StatusLabel } from '../types';

export const categoryIcons: Record<Category, ReactNode> = {
  bug: <AlertTriangle size={15} />,
  feature: <Sparkles size={15} />,
  question: <CircleHelp size={15} />,
};

export const categoryKeys = Object.keys(categoryIcons) as Category[];

export function statusLabel(statuses: StatusLabel[], key: Status, locale: 'zh' | 'en') {
  const item = statuses.find((entry) => entry.key === key);
  if (!item) return key;
  return locale === 'en' ? item.labelEn : item.labelZh;
}

export function authorStatuses(statuses: StatusLabel[]) {
  return statuses.filter((item) => item.author);
}

export function useLabels(statuses: StatusLabel[] = []) {
  const { locale, t } = useI18n();
  return {
    category: (key: Category) => t(key),
    status: (key: Status) => statusLabel(statuses, key, locale),
    ago(date: string) {
      const hours = Math.max(0, Math.floor((Date.now() - new Date(date).getTime()) / 3600000));
      if (hours < 1) return t('justNow');
      if (hours < 24) return t('hoursAgo', { count: hours });
      const days = Math.floor(hours / 24);
      return days < 30 ? t('daysAgo', { count: days }) : new Date(date).toLocaleDateString(localeTag(locale));
    },
    dateTime(date: string) {
      return new Date(date).toLocaleString(localeTag(locale));
    },
  };
}

export function excerpt(body: string) {
  const compact = body
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`([^`]*)`/g, '$1')
    .replace(/!\[[^\]]*]\([^)]*\)/g, ' ')
    .replace(/\[([^\]]+)]\([^)]*\)/g, '$1')
    .replace(/<\/?[a-z][^>]*>/gi, ' ')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/[*_~]/g, '')
    .replace(/\s+/g, ' ')
    .trim();
  return compact.length > 122 ? `${compact.slice(0, 122)}…` : compact;
}

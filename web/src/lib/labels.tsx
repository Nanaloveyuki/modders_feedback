import type { ReactNode } from 'react';
import { AlertTriangle, CircleHelp, Sparkles } from 'lucide-react';
import { localeTag, useI18n } from '../i18n/context';
import type { MessageKey } from '../i18n/messages';
import type { Category, Status } from '../types';

export const categoryIcons: Record<Category, ReactNode> = {
  bug: <AlertTriangle size={15} />,
  feature: <Sparkles size={15} />,
  question: <CircleHelp size={15} />,
};

export const categoryKeys = Object.keys(categoryIcons) as Category[];

export const statusKeys: Status[] = ['open', 'in_progress', 'resolved', 'closed', 'withdrawn'];
export const authorStatusKeys: Status[] = ['open', 'withdrawn'];

export const statusStyle: Record<Status, string> = {
  open: 'status-open',
  in_progress: 'status-progress',
  resolved: 'status-resolved',
  closed: 'status-closed',
  withdrawn: 'status-withdrawn',
};

export function useLabels() {
  const { locale, t } = useI18n();
  return {
    category: (key: Category) => t(key),
    status: (key: Status) => t(key as MessageKey),
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
  const compact = body.replace(/\s+/g, ' ');
  return compact.length > 122 ? `${compact.slice(0, 122)}…` : compact;
}

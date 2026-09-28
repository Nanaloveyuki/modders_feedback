import type { ReactNode } from 'react';
import { Bug, PawPrint, Rat, Shield, Sparkles, Squirrel } from 'lucide-react';
import type { MessageKey } from '../i18n/messages';
import type { SiteIcon, StatusLabel } from '../types';

export const siteIcons: Record<SiteIcon, { label: MessageKey; icon: ReactNode }> = {
  squirrel: { label: 'iconSquirrel', icon: <Squirrel size={18} strokeWidth={1.8} /> },
  rat: { label: 'iconRat', icon: <Rat size={18} strokeWidth={1.8} /> },
  bug: { label: 'iconBug', icon: <Bug size={18} strokeWidth={1.8} /> },
  spark: { label: 'iconSpark', icon: <Sparkles size={18} strokeWidth={1.8} /> },
  shield: { label: 'iconShield', icon: <Shield size={18} strokeWidth={1.8} /> },
  paw: { label: 'iconPaw', icon: <PawPrint size={18} strokeWidth={1.8} /> },
};

export const siteIconKeys = Object.keys(siteIcons) as SiteIcon[];

export const defaultStatuses: StatusLabel[] = [
  { key: 'open', labelZh: '待处理', labelEn: 'Open', light: '#b8892e', dark: '#e0b15a', author: true, archived: false },
  { key: 'in_progress', labelZh: '处理中', labelEn: 'In progress', light: '#3d7ea6', dark: '#7eb6d4', author: false, archived: false },
  { key: 'testing', labelZh: '测试中', labelEn: 'Testing', light: '#6a6db8', dark: '#a8aae0', author: false, archived: false },
  { key: 'fixed_unreleased', labelZh: '已修复未发布', labelEn: 'Fixed, unreleased', light: '#2f8a78', dark: '#5ec4ae', author: false, archived: true },
  { key: 'resolved', labelZh: '已解决', labelEn: 'Resolved', light: '#3f8f6b', dark: '#5dbe8a', author: false, archived: true },
  { key: 'closed', labelZh: '已关闭', labelEn: 'Closed', light: '#7d8794', dark: '#a8b0bc', author: false, archived: false },
  { key: 'withdrawn', labelZh: '已撤回', labelEn: 'Withdrawn', light: '#c45b70', dark: '#e08a98', author: true, archived: false },
];

export const defaultSettings = {
  modVersion: 'DEV BUILD',
  gameVersion: 'RIMWORLD 1.6',
  icon: 'squirrel' as SiteIcon,
  attachmentDir: '',
  statuses: defaultStatuses,
};

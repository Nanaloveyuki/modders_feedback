import type { ReactNode } from 'react';
import { AlertTriangle, CircleHelp, Sparkles } from 'lucide-react';
import type { Category, Status } from '../types';

export const categories: Record<Category, { label: string; plural: string; icon: ReactNode }> = {
  bug: { label: '错误报告', plural: '错误报告', icon: <AlertTriangle size={15} /> },
  feature: { label: '功能建议', plural: '功能建议', icon: <Sparkles size={15} /> },
  question: { label: '一般提问', plural: '一般提问', icon: <CircleHelp size={15} /> },
};

export const statusText: Record<Status, string> = {
  open: '待处理',
  in_progress: '处理中',
  resolved: '已解决',
  closed: '已关闭',
};

export const statusStyle: Record<Status, string> = {
  open: 'status-open',
  in_progress: 'status-progress',
  resolved: 'status-resolved',
  closed: 'status-closed',
};

export const categoryKeys = Object.keys(categories) as Category[];

export function ago(date: string) {
  const hours = Math.max(0, Math.floor((Date.now() - new Date(date).getTime()) / 3600000));
  if (hours < 1) return '刚刚';
  if (hours < 24) return `${hours} 小时前`;
  const days = Math.floor(hours / 24);
  return days < 30 ? `${days} 天前` : new Date(date).toLocaleDateString('zh-CN');
}

export function excerpt(body: string) {
  const compact = body.replace(/\s+/g, ' ');
  return compact.length > 122 ? `${compact.slice(0, 122)}…` : compact;
}

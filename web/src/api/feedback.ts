import type { Locale } from '../i18n/messages';
import type { Category, Feedback, FeedbackDraft, FeedbackUpdate, Mod, ModInput, SiteSettings, Status, User } from '../types';
import { api } from './client';

export async function listFeedback(mod: string, category: Category | 'all', locale: Locale) {
  const items: Feedback[] = [];
  const pageSize = 100;
  for (let offset = 0; offset <= 10000; offset += pageSize) {
    const params = new URLSearchParams({ mod, limit: String(pageSize), offset: String(offset) });
    if (category !== 'all') params.set('category', category);
    const page = await api<Feedback[]>(`/feedback?${params}`, locale);
    items.push(...page);
    if (page.length < pageSize) return items;
  }
  return items;
}

export function currentUser(locale: Locale) {
  return api<User>('/auth/me', locale);
}

export function login(username: string, password: string, locale: Locale) {
  return api<User>('/auth/login', locale, { method: 'POST', body: JSON.stringify({ username, password }) });
}

export function logout(locale: Locale) {
  return api<void>('/auth/logout', locale, { method: 'POST' });
}

export function createFeedback(mod: string, data: FeedbackDraft, locale: Locale) {
  return api<Feedback>(`/feedback?mod=${encodeURIComponent(mod)}`, locale, { method: 'POST', body: JSON.stringify(data) });
}

export function updateStatus(id: number, status: Status, locale: Locale) {
  return api<{ status: Status }>(`/feedback/${id}/status`, locale, { method: 'PATCH', body: JSON.stringify({ status }) });
}

export function updateFeedback(id: number, data: FeedbackUpdate, locale: Locale) {
  return api<Feedback>(`/feedback/${id}`, locale, { method: 'PATCH', body: JSON.stringify(data) });
}

export function deleteFeedback(id: number, locale: Locale) {
  return api<void>(`/feedback/${id}`, locale, { method: 'DELETE' });
}

export function siteSettings(locale: Locale) {
  return api<SiteSettings>('/settings', locale);
}

export function updateSiteSettings(data: SiteSettings, locale: Locale) {
  return api<SiteSettings>('/settings', locale, { method: 'PUT', body: JSON.stringify(data) });
}

export function listMods(locale: Locale) {
  return api<Mod[]>('/mods', locale);
}

export function createMod(data: ModInput, locale: Locale) {
  return api<Mod>('/mods', locale, { method: 'POST', body: JSON.stringify(data) });
}

export function updateMod(id: number, data: ModInput, locale: Locale) {
  return api<Mod>(`/mods/${id}`, locale, { method: 'PATCH', body: JSON.stringify(data) });
}

export function deleteMod(id: number, locale: Locale) {
  return api<void>(`/mods/${id}`, locale, { method: 'DELETE' });
}

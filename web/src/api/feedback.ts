import type { Category, Feedback, FeedbackDraft, Status, User } from '../types';
import { api } from './client';

export async function listFeedback(category: Category | 'all') {
  const items: Feedback[] = [];
  const pageSize = 100;
  for (let offset = 0; offset <= 10000; offset += pageSize) {
    const params = new URLSearchParams({ limit: String(pageSize), offset: String(offset) });
    if (category !== 'all') params.set('category', category);
    const page = await api<Feedback[]>(`/feedback?${params}`);
    items.push(...page);
    if (page.length < pageSize) return items;
  }
  return items;
}

export function currentUser() {
  return api<User>('/auth/me');
}

export function login(username: string, password: string) {
  return api<User>('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) });
}

export function logout() {
  return api<void>('/auth/logout', { method: 'POST' });
}

export function createFeedback(data: FeedbackDraft) {
  return api<Feedback>('/feedback', { method: 'POST', body: JSON.stringify(data) });
}

export function updateStatus(id: number, status: Status) {
  return api<{ status: Status }>(`/feedback/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) });
}

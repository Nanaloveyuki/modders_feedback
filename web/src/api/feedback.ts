import type { Category, Feedback, FeedbackDraft, Status, User } from '../types';
import { api } from './client';

export function listFeedback(category: Category | 'all') {
  const query = category === 'all' ? '' : `?category=${category}`;
  return api<Feedback[]>(`/feedback${query}`);
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

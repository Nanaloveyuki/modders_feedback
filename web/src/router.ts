import type { Category } from './types';

export type Route =
  | { name: 'board' }
  | { name: 'feedback'; slug: string; category: Category; publicId: string }
  | { name: 'admin'; page: AdminPage }
  | { name: 'account' };

export type AdminPage = 'settings' | 'mods' | 'feedback';

export const adminPages: AdminPage[] = ['settings', 'mods', 'feedback'];

export function adminPath(page: AdminPage) {
  return `/admin/${page}`;
}

export function parseRoute(pathname: string): Route {
  const path = pathname.replace(/\/+$/, '') || '/';
  if (path === '/admin') return { name: 'admin', page: 'settings' };
  if (path.startsWith('/admin/')) {
    const page = path.slice('/admin/'.length);
    if (page === 'settings' || page === 'mods' || page === 'feedback') return { name: 'admin', page };
    return { name: 'admin', page: 'settings' };
  }
  if (path === '/account') return { name: 'account' };
  const item = path.match(/^\/mod\/([a-z0-9]+(?:-[a-z0-9]+)*)\/(bugs|feature|question)\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$/);
  if (item) {
    const category: Category = item[2] === 'bugs' ? 'bug' : item[2] === 'feature' ? 'feature' : 'question';
    return { name: 'feedback', slug: item[1], category, publicId: item[3] };
  }
  return { name: 'board' };
}

export function currentRoute() {
  return parseRoute(window.location.pathname);
}

export function navigate(path: string) {
  if (window.location.pathname === path) return;
  window.history.pushState(null, '', path);
  window.dispatchEvent(new PopStateEvent('popstate'));
}

export function feedbackPath(slug: string, category: Category, publicId: string) {
  const segment = category === 'bug' ? 'bugs' : category;
  return `/mod/${slug}/${segment}/${publicId}`;
}

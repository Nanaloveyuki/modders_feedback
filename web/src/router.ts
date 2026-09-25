export type Route =
  | { name: 'board' }
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

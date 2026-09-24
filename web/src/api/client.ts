import { translate, type Locale } from '../i18n/messages';

export async function api<T>(path: string, locale: Locale, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', 'Accept-Language': locale, ...init?.headers },
  });
  if (!response.ok) {
    const fallback = translate(locale, 'requestFailed');
    const result = await response.json().catch(() => ({ error: fallback }));
    throw new Error(result.error || translate(locale, 'requestFailedShort'));
  }
  return response.status === 204 ? (undefined as T) : response.json();
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  });
  if (!response.ok) {
    const result = await response.json().catch(() => ({ error: '请求失败，请稍后重试' }));
    throw new Error(result.error || '请求失败');
  }
  return response.status === 204 ? (undefined as T) : response.json();
}

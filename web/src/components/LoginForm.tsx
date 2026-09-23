import { useState, type FormEvent } from 'react';
import { ArrowRight, LockKeyhole } from 'lucide-react';

type Props = {
  onSubmit: (username: string, password: string) => Promise<void>;
};

export function LoginForm({ onSubmit }: Props) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSubmit(String(form.get('username')), String(form.get('password')));
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : '登录失败');
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="form-stack" onSubmit={handleSubmit}>
      <label>用户名<input name="username" autoComplete="username" required maxLength={80} autoFocus placeholder="输入用户名" /></label>
      <label>密码<input name="password" type="password" autoComplete="current-password" required minLength={1} maxLength={256} placeholder="输入密码" /></label>
      {error && <p className="form-error" role="alert">{error}</p>}
      <button className="form-submit" disabled={pending}>{pending ? '正在验证…' : '登录'}<ArrowRight size={16} /></button>
      <p className="form-hint"><LockKeyhole size={13} />此站不开放自助注册。账号由模组维护者提供。</p>
    </form>
  );
}

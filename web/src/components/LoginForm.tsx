import { useState, type FormEvent } from 'react';
import { ArrowRight, LockKeyhole } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { useTheme } from '../theme/context';
import { ActionButton, Field } from './ui';

type Props = {
  mode: 'login' | 'register';
  onSubmit: (username: string, password: string, email: string, qq: string) => Promise<void>;
  onSwitch: () => void;
};

export function LoginForm({ mode, onSubmit, onSwitch }: Props) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field };

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSubmit(String(form.get('username') ?? ''), String(form.get('password') ?? ''), String(form.get('email') ?? ''), String(form.get('qq') ?? ''));
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t(mode === 'register' ? 'registerFailed' : 'loginFailed'));
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="form-stack" onSubmit={handleSubmit}>
      <Field label={t('username')}>
        <input name="username" autoComplete="username" required minLength={3} maxLength={32} pattern="[A-Za-z0-9_-]{3,32}" autoFocus style={control} />
      </Field>
      <Field label={t('password')}>
        <input name="password" type="password" autoComplete={mode === 'register' ? 'new-password' : 'current-password'} required minLength={mode === 'register' ? 12 : 1} maxLength={128} placeholder={mode === 'register' ? t('passwordRule') : undefined} style={control} />
      </Field>
      {mode === 'register' && (
        <>
          <Field label={t('emailOptional')}>
            <input name="email" type="email" autoComplete="email" maxLength={254} placeholder="name@example.com" style={control} />
          </Field>
          <Field label={t('qqOptional')}>
            <input name="qq" inputMode="numeric" autoComplete="off" pattern="[1-9][0-9]{4,10}" maxLength={11} style={control} />
          </Field>
        </>
      )}
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
      <ActionButton className="form-submit" disabled={pending}>{pending ? t(mode === 'register' ? 'registering' : 'verifying') : t(mode === 'register' ? 'register' : 'login')}<ArrowRight size={16} /></ActionButton>
      <button type="button" className="form-switch" style={{ color: palette.gold }} onClick={onSwitch}>
        <LockKeyhole size={13} />{t(mode === 'register' ? 'haveAccount' : 'needAccount')}
      </button>
    </form>
  );
}

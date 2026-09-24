import { useState, type FormEvent } from 'react';
import { ArrowRight, LockKeyhole } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { useTheme } from '../theme/context';
import { ActionButton, Field } from './ui';

type Props = {
  onSubmit: (username: string, password: string) => Promise<void>;
};

export function LoginForm({ onSubmit }: Props) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field, borderColor: palette.line };

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSubmit(String(form.get('username')), String(form.get('password')));
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('loginFailed'));
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="form-stack" onSubmit={handleSubmit}>
      <Field label={t('username')}>
        <input name="username" autoComplete="username" required maxLength={80} autoFocus placeholder={t('usernamePlaceholder')} style={control} />
      </Field>
      <Field label={t('password')}>
        <input name="password" type="password" autoComplete="current-password" required minLength={1} maxLength={256} placeholder={t('passwordPlaceholder')} style={control} />
      </Field>
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
      <ActionButton className="form-submit" disabled={pending}>{pending ? t('verifying') : t('login')}<ArrowRight size={16} /></ActionButton>
      <p className="form-hint" style={{ color: palette.faint }}><LockKeyhole size={13} />{t('noRegister')}</p>
    </form>
  );
}

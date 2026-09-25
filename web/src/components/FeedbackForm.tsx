import { useState, type FormEvent } from 'react';
import { ArrowRight } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { categoryIcons, categoryKeys } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Category, FeedbackDraft } from '../types';
import type { Palette } from '../theme/theme';
import { ActionButton, Field } from './ui';

type Props = {
  onSubmit: (data: FeedbackDraft) => Promise<void>;
};

function categoryColor(palette: Palette, category: Category) {
  if (category === 'bug') return palette.bug;
  if (category === 'feature') return palette.feature;
  return palette.question;
}

export function FeedbackForm({ onSubmit }: Props) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const [category, setCategory] = useState<Category>('bug');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field };

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    const data: FeedbackDraft = {
      category,
      title: String(form.get('title') ?? ''),
      body: String(form.get('body') ?? ''),
      gameVersion: String(form.get('gameVersion') ?? ''),
      modVersion: String(form.get('modVersion') ?? ''),
      modList: String(form.get('modList') ?? ''),
      saveLink: String(form.get('saveLink') ?? ''),
    };
    try {
      await onSubmit(data);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('submitFailed'));
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="form-stack feedback-form" onSubmit={handleSubmit}>
      <fieldset className="category-picker">
        <legend style={{ color: palette.muted }}>{t('feedbackType')}</legend>
        {categoryKeys.map((key) => {
          const picked = category === key;
          return (
            <button
              type="button"
              key={key}
              className={picked ? `picked picked-${key}` : ''}
              onClick={() => setCategory(key)}
              style={{
                color: picked ? palette.activeInk : palette.muted,
                background: picked ? palette.active : palette.field,
                borderColor: 'transparent',
              }}
            >
              <span style={{ color: categoryColor(palette, key) }}>{categoryIcons[key]}</span>{t(key)}
            </button>
          );
        })}
      </fieldset>
      <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} style={control} /></Field>
      <Field label={t('description')}><textarea name="body" required minLength={10} maxLength={12000} rows={5} style={control} /></Field>
      <div className="form-two">
        <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue="RimWorld 1.6" style={control} /></Field>
        <Field label={t('modVersion')}><input name="modVersion" maxLength={80} style={control} /></Field>
      </div>
      {category === 'bug' && (
        <>
          <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={2} style={control} /></Field>
          <Field label={t('saveLink')}><input name="saveLink" type="url" maxLength={500} style={control} /></Field>
        </>
      )}
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
      <ActionButton className="form-submit" disabled={pending}>{pending ? t('publishing') : t('create')}<ArrowRight size={16} /></ActionButton>
    </form>
  );
}

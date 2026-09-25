import { useState, type FormEvent } from 'react';
import { ArrowLeft, ArrowRight } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { AttachmentEditor, editorTarget } from './AttachmentEditor';
import { categoryIcons, categoryKeys } from '../lib/labels';
import { navigate } from '../router';
import { useTheme } from '../theme/context';
import type { Category, FeedbackDraft, Mod, SiteSettings } from '../types';
import type { Palette } from '../theme/theme';
import { ActionButton, Field } from './ui';

type Props = {
  mod?: Mod;
  settings: SiteSettings;
  onSubmit: (data: FeedbackDraft) => Promise<void>;
};

function categoryColor(palette: Palette, category: Category) {
  if (category === 'bug') return palette.bug;
  if (category === 'feature') return palette.feature;
  return palette.question;
}

export function FeedbackForm({ mod, settings, onSubmit }: Props) {
  const { t, locale } = useI18n();
  const { palette } = useTheme();
  const [category, setCategory] = useState<Category>('bug');
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [preview, setPreview] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field };

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const length = [...body.trim()].length;
    if (length < 10 || length > 12000) {
      setError(t('bodyLength'));
      setPending(false);
      setPreview(false);
      return;
    }
    const form = new FormData(event.currentTarget);
    const data: FeedbackDraft = {
      category,
      title,
      body,
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
    <main className="issue-page">
      <button type="button" className="issue-back" style={{ color: palette.gold }} onClick={() => navigate('/')}>
        <ArrowLeft size={15} />{t('adminBack')}
      </button>
      <header className="issue-header">
        <h1 style={{ color: palette.text }}>{t('create')}</h1>
      </header>
      <form className="issue-layout" onSubmit={handleSubmit}>
        <section className="issue-main" style={{ background: palette.surface }}>
          <div className="issue-comment-head" style={{ color: palette.muted }}>
            <strong style={{ color: palette.text }}>{t('description')}</strong>
            <button type="button" className="issue-edit" style={{ color: palette.gold }} onClick={() => setPreview((current) => !current)}>{preview ? t('writeMarkdown') : t('previewMarkdown')}</button>
          </div>
          <div className="issue-editor">
            <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} value={title} onChange={(event) => setTitle(event.target.value)} style={control} /></Field>
            <AttachmentEditor
              body={body}
              onBody={setBody}
              preview={preview}
              rows={16}
              locale={locale}
              target={editorTarget()}
              control={control}
              onError={setError}
            />
            {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
          </div>
        </section>
        <aside className="issue-side">
          <section style={{ background: palette.surface }}>
            <h2 style={{ color: palette.faint }}>{t('feedbackType')}</h2>
            <div className="compose-categories">
              {categoryKeys.map((key) => {
                const picked = category === key;
                return (
                  <button
                    type="button"
                    key={key}
                    className={picked ? 'picked' : ''}
                    onClick={() => setCategory(key)}
                    style={{
                      color: picked ? palette.activeInk : palette.muted,
                      background: picked ? palette.active : palette.field,
                    }}
                  >
                    <span style={{ color: categoryColor(palette, key) }}>{categoryIcons[key]}</span>{t(key)}
                  </button>
                );
              })}
            </div>
          </section>
          <section style={{ background: palette.surface }}>
            <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue={mod?.gameVersion || settings.gameVersion} style={control} /></Field>
            <Field label={t('modVersion')}><input name="modVersion" maxLength={80} defaultValue={mod?.modVersion || settings.modVersion} style={control} /></Field>
          </section>
          {category === 'bug' && (
            <section style={{ background: palette.surface }}>
              <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={4} style={control} /></Field>
              <Field label={t('saveLink')}><input name="saveLink" type="url" maxLength={500} style={control} /></Field>
            </section>
          )}
          <ActionButton className="form-submit" disabled={pending}>{pending ? t('publishing') : t('create')}<ArrowRight size={16} /></ActionButton>
        </aside>
      </form>
    </main>
  );
}

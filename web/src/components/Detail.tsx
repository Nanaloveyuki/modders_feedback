import { useEffect, useState, type FormEvent } from 'react';
import { ArrowLeft, Clipboard, ExternalLink } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { AttachmentEditor, editorTarget } from './AttachmentEditor';
import { authorStatusKeys, categoryIcons, statusKeys, statusStyle, useLabels } from '../lib/labels';
import { feedbackPath, navigate } from '../router';
import { useTheme } from '../theme/context';
import type { Feedback, FeedbackUpdate, Status } from '../types';
import { statusColors } from './FeedbackList';
import { MarkdownBody } from './MarkdownBody';
import { ActionButton, Avatar, Field } from './ui';

type Props = {
  item: Feedback;
  modSlug: string;
  canManage: boolean;
  canEdit: boolean;
  onStatus: (item: Feedback, status: Status) => Promise<void>;
  onSave: (item: Feedback, update: FeedbackUpdate) => Promise<void>;
};

export function Detail({ item, modSlug, canManage, canEdit, onStatus, onSave }: Props) {
  const { t, locale } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const [editing, setEditing] = useState(false);
  const [body, setBody] = useState(item.body);
  const [preview, setPreview] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field };
  const choices = canManage ? statusKeys : authorStatusKeys;
  const categoryColor = item.category === 'bug' ? palette.bug : item.category === 'feature' ? palette.feature : palette.question;
  const number = String(item.categoryNumber ?? item.id).padStart(3, '0');
  const canChange = canEdit || canManage;

  useEffect(() => {
    setBody(item.body);
    setEditing(false);
    setPreview(false);
    setError('');
  }, [item.id, item.body]);

  async function save(event: FormEvent<HTMLFormElement>) {
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
    try {
      await onSave(item, {
        title: String(form.get('title') ?? ''),
        body,
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        modList: String(form.get('modList') ?? ''),
        saveLink: String(form.get('saveLink') ?? ''),
      });
      setEditing(false);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('recordFailed'));
    } finally {
      setPending(false);
    }
  }

  function copyAddress() {
    const path = item.publicId ? feedbackPath(modSlug, item.category, item.publicId) : window.location.pathname;
    void navigator.clipboard?.writeText(new URL(path, window.location.origin).href);
  }

  return (
    <main className="issue-page">
      <button type="button" className="issue-back" style={{ color: palette.gold }} onClick={() => navigate('/')}>
        <ArrowLeft size={15} />{t('adminBack')}
      </button>
      <header className="issue-header">
        <h1 style={{ color: palette.text }}>
          {item.title} <span style={{ color: palette.faint }}>#{number}</span>
        </h1>
        <div className="issue-meta" style={{ color: palette.muted }}>
          <span className={`status-pill ${statusStyle[item.status]}`} style={statusColors(palette, item.status)}><i />{labels.status(item.status)}</span>
          <span className={`row-category cat-${item.category}`} style={{ color: categoryColor }}>
            {categoryIcons[item.category]}<span>{labels.category(item.category)}</span>
          </span>
          <span className="issue-opened">
            <Avatar name={item.author} />
            <strong style={{ color: palette.text }}>{item.author}</strong>
            <span style={{ color: palette.faint }}>{labels.dateTime(item.createdAt)}</span>
          </span>
        </div>
      </header>
      <div className="issue-layout">
        <article className="issue-main" style={{ background: palette.surface }}>
          <div className="issue-comment-head" style={{ color: palette.muted }}>
            <Avatar name={item.author} />
            <strong style={{ color: palette.text }}>{item.author}</strong>
            <span style={{ color: palette.faint }}>{labels.dateTime(item.createdAt)}</span>
            {canChange && !editing && (
              <button type="button" className="issue-edit" style={{ color: palette.gold }} onClick={() => setEditing(true)}>
                {t(canEdit ? 'editOwn' : 'editRecord')}
              </button>
            )}
          </div>
          {editing ? (
            <form className="form-stack issue-editor" onSubmit={(event) => void save(event)}>
              <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} defaultValue={item.title} style={control} /></Field>
              <Field label={t('description')}>
                <AttachmentEditor body={body} onBody={setBody} preview={preview} rows={12} locale={locale} target={editorTarget(item.id)} control={control} onError={setError} />
              </Field>
              <button type="button" className="form-switch own-edit" style={{ color: palette.gold }} onClick={() => setPreview((current) => !current)}>{preview ? t('writeMarkdown') : t('previewMarkdown')}</button>
              <div className="form-two">
                <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
                <Field label={t('modVersion')}><input name="modVersion" maxLength={80} defaultValue={item.modVersion} style={control} /></Field>
              </div>
              <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={3} defaultValue={item.modList} style={control} /></Field>
              <Field label={t('saveLink')}><input name="saveLink" type="url" maxLength={500} defaultValue={item.saveLink} style={control} /></Field>
              {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
              <div className="issue-editor-actions">
                <button type="button" className="form-switch" style={{ color: palette.muted }} onClick={() => { setEditing(false); setPreview(false); setBody(item.body); setError(''); }}>{t('close')}</button>
                <ActionButton className="form-submit" disabled={pending}>{pending ? t('savingRecord') : t('saveOwn')}</ActionButton>
              </div>
            </form>
          ) : (
            <div className="detail-body"><MarkdownBody source={item.body} /></div>
          )}
        </article>
        <aside className="issue-side">
          <section style={{ background: palette.surface }}>
            <h2 style={{ color: palette.faint }}>{t('updateStatus')}</h2>
            {(canManage || canEdit) ? (
              <label className="status-editor" style={{ color: palette.muted }}>
                <select value={item.status} disabled={!canManage && !choices.includes(item.status)} onChange={(event) => void onStatus(item, event.target.value as Status)} style={{ color: palette.text, background: palette.field }}>
                  {!choices.includes(item.status) && <option value={item.status}>{labels.status(item.status)}</option>}
                  {choices.map((key) => <option key={key} value={key}>{labels.status(key)}</option>)}
                </select>
              </label>
            ) : (
              <strong className={`status-pill ${statusStyle[item.status]}`} style={statusColors(palette, item.status)}><i />{labels.status(item.status)}</strong>
            )}
          </section>
          <section style={{ background: palette.surface }}>
            <h2 style={{ color: palette.faint }}>RIMWORLD</h2>
            <strong style={{ color: palette.text }}>{item.gameVersion || t('missing')}</strong>
            <h2 style={{ color: palette.faint }}>MOD VERSION</h2>
            <strong style={{ color: palette.text }}>{item.modVersion || t('missing')}</strong>
          </section>
          {item.modList && (
            <section style={{ background: palette.surface }}>
              <h2 style={{ color: palette.faint }}>{t('extraMods')}</h2>
              <pre style={{ color: palette.text }}>{item.modList}</pre>
            </section>
          )}
          <section style={{ background: palette.surface }}>
            {item.saveLink && <a className="save-link" href={item.saveLink} target="_blank" rel="noreferrer" style={{ color: palette.gold }}><ExternalLink size={14} />{t('openSave')}</a>}
            <button type="button" className="issue-copy" style={{ color: palette.gold }} onClick={copyAddress}>
              <Clipboard size={14} />{t('copyAddress')}
            </button>
          </section>
        </aside>
      </div>
    </main>
  );
}

import { useEffect, useState, type FormEvent } from 'react';
import { Clipboard, ExternalLink, X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { authorStatusKeys, categoryIcons, statusKeys, statusStyle, useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Feedback, FeedbackUpdate, Status } from '../types';
import { statusColors } from './FeedbackList';
import { ActionButton, Avatar, Field, IconButton, Panel, Scrim } from './ui';

type Props = {
  item: Feedback;
  canManage: boolean;
  canEdit: boolean;
  onClose: () => void;
  onStatus: (item: Feedback, status: Status) => Promise<void>;
  onSave: (item: Feedback, update: FeedbackUpdate) => Promise<void>;
};

export function Detail({ item, canManage, canEdit, onClose, onStatus, onSave }: Props) {
  const { t } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const [editing, setEditing] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field, borderColor: palette.line };
  const choices = canManage ? statusKeys : authorStatusKeys;
  const categoryColor = item.category === 'bug' ? palette.bug : item.category === 'feature' ? palette.feature : palette.question;
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);


  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSave(item, {
        title: String(form.get('title') ?? ''),
        body: String(form.get('body') ?? ''),
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
  return (
    <Scrim onClose={onClose}>
      <Panel className="detail-panel" labelledBy="detail-title">
        <div className="detail-top">
          <span className={`row-category cat-${item.category}`} style={{ color: categoryColor }}>
            {categoryIcons[item.category]}<span>{labels.category(item.category)}</span>
          </span>
          <IconButton label={t('close')} onClick={onClose}><X size={18} /></IconButton>
        </div>
        <h2 id="detail-title" className="detail-title" style={{ color: palette.text }}>{item.title}</h2>
        <div className="detail-byline" style={{ color: palette.faint }}>
          <Avatar name={item.author} />
          <strong style={{ color: palette.text }}>{item.author}</strong>
          <span>·</span>
          <span>{labels.dateTime(item.createdAt)}</span>
        </div>
        <div className="detail-body" style={{ color: palette.text }}>{item.body}</div>
        <div className="detail-facts" style={{ borderColor: palette.line }}>
          <div><span style={{ color: palette.faint }}>RIMWORLD</span><strong style={{ color: palette.muted }}>{item.gameVersion || t('missing')}</strong></div>
          <div><span style={{ color: palette.faint }}>MOD VERSION</span><strong style={{ color: palette.muted }}>{item.modVersion || t('missing')}</strong></div>
          <div><span style={{ color: palette.faint }}>STATUS</span><strong className={`status-pill ${statusStyle[item.status]}`} style={statusColors(palette, item.status)}><i />{labels.status(item.status)}</strong></div>
        </div>
        {item.modList && (
          <div className="detail-extra">
            <h3 style={{ color: palette.faint }}>{t('extraMods')}</h3>
            <pre style={{ color: palette.text, background: palette.field }}>{item.modList}</pre>
          </div>
        )}
        {item.saveLink && <a className="save-link" href={item.saveLink} target="_blank" rel="noreferrer" style={{ color: palette.gold }}><ExternalLink size={14} />{t('openSave')}</a>}
        {(canManage || canEdit) && (
          <label className="status-editor" style={{ color: palette.muted }}>
            {t('updateStatus')}
            <select value={item.status} disabled={!canManage && !choices.includes(item.status)} onChange={(event) => void onStatus(item, event.target.value as Status)} style={{ color: palette.text, background: palette.field, borderColor: palette.line }}>
              {!choices.includes(item.status) && <option value={item.status}>{labels.status(item.status)}</option>}
              {choices.map((key) => <option key={key} value={key}>{labels.status(key)}</option>)}
            </select>
          </label>
        )}
        {canEdit && !editing && <button type="button" className="form-switch own-edit" style={{ color: palette.gold }} onClick={() => setEditing(true)}>{t('editOwn')}</button>}
        {canEdit && editing && (
          <form className="form-stack own-editor" onSubmit={(event) => void save(event)}>
            <p className="form-hint" style={{ color: palette.faint }}>{t('editOwnHint')}</p>
            <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} defaultValue={item.title} style={control} /></Field>
            <Field label={t('description')}><textarea name="body" required minLength={10} maxLength={12000} rows={5} defaultValue={item.body} style={control} /></Field>
            <div className="form-two">
              <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
              <Field label={t('modVersion')}><input name="modVersion" maxLength={80} defaultValue={item.modVersion} style={control} /></Field>
            </div>
            <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={2} defaultValue={item.modList} style={control} /></Field>
            <Field label={t('saveLink')}><input name="saveLink" type="url" maxLength={500} defaultValue={item.saveLink} style={control} /></Field>
            {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
            <ActionButton className="form-submit" disabled={pending}>{pending ? t('savingRecord') : t('saveOwn')}</ActionButton>
          </form>
        )}
        <div className="detail-bottom" style={{ borderColor: palette.line, color: palette.faint }}>
          <span>ISSUE / {String(item.id).padStart(4, '0')}</span>
          <button onClick={() => { void navigator.clipboard?.writeText(window.location.href); }} style={{ color: palette.gold }}>
            <Clipboard size={14} />{t('copyAddress')}
          </button>
        </div>
      </Panel>
    </Scrim>
  );
}

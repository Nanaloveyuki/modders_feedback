import { useEffect, useState, type FormEvent } from 'react';
import { ArrowLeft, Plus, Trash2 } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { AttachmentEditor, editorTarget } from './AttachmentEditor';
import { siteIconKeys, siteIcons } from '../lib/icons';
import { useLabels } from '../lib/labels';
import { adminPages, adminPath, navigate, type AdminPage } from '../router';
import { useTheme } from '../theme/context';
import type { Feedback, FeedbackUpdate, Mod, ModInput, SiteIcon, SiteSettings, StatusLabel } from '../types';
import { ActionButton, Field } from './ui';

type Props = {
  page: AdminPage;
  items: Feedback[];
  settings: SiteSettings;
  mods: Mod[];
  onSaveSettings: (settings: SiteSettings) => Promise<void>;
  onSaveStatuses: (statuses: StatusLabel[]) => Promise<void>;
  onSaveRecord: (item: Feedback, update: FeedbackUpdate) => Promise<void>;
  onDeleteRecord: (item: Feedback) => Promise<void>;
  onSaveMod: (item: Mod, input: ModInput) => Promise<void>;
  onAddMod: (input: ModInput) => Promise<void>;
  onDeleteMod: (item: Mod) => Promise<void>;
};

const emptyMod = {
  slug: '',
  name: '',
  gameVersion: 'RIMWORLD 1.6',
  modVersion: 'DEV BUILD',
  steamUrl: '',
  githubUrl: '',
};

export function AdminPanel({ page, items, settings, mods, onSaveSettings, onSaveStatuses, onSaveRecord, onDeleteRecord, onSaveMod, onAddMod, onDeleteMod }: Props) {
  const { t } = useI18n();
  const { palette } = useTheme();

  return (
    <section className="admin-page">
      <header className="admin-page-head">
        <button type="button" className="admin-back" style={{ color: palette.gold }} onClick={() => navigate('/')}>
          <ArrowLeft size={15} />{t('adminBack')}
        </button>
        <h1 style={{ color: palette.text }}>{t(page === 'settings' ? 'adminSettings' : page === 'mods' ? 'adminMods' : 'adminFeedback')}</h1>
      </header>
      <nav className="admin-nav" aria-label={t('adminNav')} style={{ background: palette.surface }}>
        {adminPages.map((entry) => {
          const active = entry === page;
          return (
            <a
              key={entry}
              href={adminPath(entry)}
              aria-current={active ? 'page' : undefined}
              style={{
                color: active ? palette.activeInk : palette.muted,
                background: active ? palette.active : 'transparent',
              }}
              onClick={(event) => {
                event.preventDefault();
                navigate(adminPath(entry));
              }}
            >
              {t(entry === 'settings' ? 'adminSettings' : entry === 'mods' ? 'adminMods' : 'adminFeedback')}
            </a>
          );
        })}
      </nav>
      {page === 'settings' && <SettingsPage settings={settings} onSaveSettings={onSaveSettings} onSaveStatuses={onSaveStatuses} />}
      {page === 'mods' && <ModsPage mods={mods} onSaveMod={onSaveMod} onAddMod={onAddMod} onDeleteMod={onDeleteMod} />}
      {page === 'feedback' && <FeedbackPage items={items} onSaveRecord={onSaveRecord} onDeleteRecord={onDeleteRecord} />}
    </section>
  );
}

function SettingsPage({ settings, onSaveSettings, onSaveStatuses }: Pick<Props, 'settings' | 'onSaveSettings' | 'onSaveStatuses'>) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const control = { color: palette.text, background: palette.field };
  const [icon, setIcon] = useState<SiteIcon>(settings.icon);
  const [statuses, setStatuses] = useState<StatusLabel[]>(settings.statuses);
  const [pending, setPending] = useState(false);
  const [statusPending, setStatusPending] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => { setStatuses(settings.statuses); }, [settings.statuses]);

  async function saveSettings(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    setError('');
    try {
      await onSaveSettings({
        modVersion: String(form.get('modVersion') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        icon,
        attachmentDir: String(form.get('attachmentDir') ?? ''),
        statuses: settings.statuses,
      });
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('settingsFailed'));
    } finally {
      setPending(false);
    }
  }

  function updateStatus(index: number, patch: Partial<StatusLabel>) {
    setStatuses((current) => current.map((item, position) => position === index ? { ...item, ...patch } : item));
  }

  async function saveStatuses() {
    setStatusPending(true);
    setError('');
    try {
      await onSaveStatuses(statuses);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('statusesFailed'));
    } finally {
      setStatusPending(false);
    }
  }

  return (
    <div className="admin-stack">
      <form className="form-stack admin-settings admin-card" style={{ background: palette.surface }} onSubmit={(event) => void saveSettings(event)}>
        <div className="form-two">
          <Field label={t('currentMod')}><input name="modVersion" required maxLength={40} defaultValue={settings.modVersion} style={control} /></Field>
          <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} defaultValue={settings.gameVersion} style={control} /></Field>
        </div>
        <Field label={t('attachmentDir')}><input name="attachmentDir" required maxLength={240} defaultValue={settings.attachmentDir} style={control} /></Field>
        <IconPicker value={icon} onChange={setIcon} />
        <ActionButton className="form-submit" disabled={pending}>{pending ? t('savingSettings') : t('saveSettings')}</ActionButton>
      </form>
      <section className="form-stack admin-card" style={{ background: palette.surface }}>
        <h2 style={{ color: palette.text }}>{t('statusTags')}</h2>
        <div className="status-editor-list">
          {statuses.map((item, index) => (
            <div className="status-editor-row" key={`${item.key}-${index}`}>
              <Field label={t('statusKey')}><input value={item.key} maxLength={32} disabled={item.key === 'open'} onChange={(event) => updateStatus(index, { key: event.target.value })} style={control} /></Field>
              <Field label={t('statusNameZh')}><input value={item.labelZh} maxLength={24} onChange={(event) => updateStatus(index, { labelZh: event.target.value })} style={control} /></Field>
              <Field label={t('statusNameEn')}><input value={item.labelEn} maxLength={24} onChange={(event) => updateStatus(index, { labelEn: event.target.value })} style={control} /></Field>
              <label className="status-color" style={{ color: palette.muted }}>
                {t('statusLight')}
                <input type="color" value={item.light} onChange={(event) => updateStatus(index, { light: event.target.value })} />
                <span className="status-pill" style={{ color: '#1c1812', background: item.light }}><i />{item.labelZh}</span>
              </label>
              <label className="status-color" style={{ color: palette.muted }}>
                {t('statusDark')}
                <input type="color" value={item.dark} onChange={(event) => updateStatus(index, { dark: event.target.value })} />
                <span className="status-pill" style={{ color: '#f4efe4', background: item.dark }}><i />{item.labelZh}</span>
              </label>
              <label className="status-flag" style={{ color: palette.text }}>
                <input type="checkbox" checked={item.author} disabled={item.key === 'open'} onChange={(event) => updateStatus(index, { author: event.target.checked })} />
                {t('statusAuthor')}
              </label>
              <label className="status-flag" style={{ color: palette.text }}>
                <input type="checkbox" checked={item.archived} onChange={(event) => updateStatus(index, { archived: event.target.checked })} />
                {t('statusArchived')}
              </label>
              <button type="button" className="status-remove" disabled={item.key === 'open'} style={{ color: palette.danger }} onClick={() => setStatuses((current) => current.filter((_, position) => position !== index))}>
                <Trash2 size={14} />{t('removeStatus')}
              </button>
            </div>
          ))}
        </div>
        <div className="issue-editor-actions">
          <button type="button" className="form-switch" style={{ color: palette.gold }} onClick={() => setStatuses((current) => [...current, { key: `status_${current.length + 1}`, labelZh: '新状态', labelEn: 'New', light: '#3f8f6b', dark: '#5dbe8a', author: false, archived: false }])}>
            <Plus size={14} />{t('addStatus')}
          </button>
          <ActionButton disabled={statusPending} onClick={() => void saveStatuses()}>{statusPending ? t('savingStatuses') : t('saveStatuses')}</ActionButton>
        </div>
      </section>
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
    </div>
  );
}

function ModsPage({ mods, onSaveMod, onAddMod, onDeleteMod }: Pick<Props, 'mods' | 'onSaveMod' | 'onAddMod' | 'onDeleteMod'>) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const control = { color: palette.text, background: palette.field };
  const [modIcons, setModIcons] = useState<Record<number, SiteIcon>>({});
  const [draft, setDraft] = useState(emptyMod);
  const [newIcon, setNewIcon] = useState<SiteIcon>('squirrel');
  const [pending, setPending] = useState<number | 'new' | null>(null);
  const [error, setError] = useState('');

  function iconFor(item: Mod) {
    return modIcons[item.id] ?? item.icon;
  }

  async function saveMod(event: FormEvent<HTMLFormElement>, item: Mod) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(item.id);
    setError('');
    try {
      await onSaveMod(item, {
        slug: String(form.get('slug') ?? ''),
        name: String(form.get('name') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        icon: iconFor(item),
        steamUrl: String(form.get('steamUrl') ?? ''),
        githubUrl: String(form.get('githubUrl') ?? ''),
      });
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('modFailed'));
    } finally {
      setPending(null);
    }
  }

  async function addMod(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending('new');
    setError('');
    try {
      await onAddMod({ ...draft, icon: newIcon });
      setDraft(emptyMod);
      setNewIcon('squirrel');
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('modFailed'));
    } finally {
      setPending(null);
    }
  }

  async function removeMod(item: Mod) {
    if (!window.confirm(t('confirmDeleteMod'))) return;
    setPending(item.id);
    setError('');
    try {
      await onDeleteMod(item);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('deleteModFailed'));
    } finally {
      setPending(null);
    }
  }

  return (
    <div className="admin-mods">
      {mods.length === 0 && <p className="admin-empty" style={{ color: palette.faint }}>{t('modsEmpty')}</p>}
      {mods.map((item) => {
        const busy = pending === item.id;
        const picked = iconFor(item);
        return (
          <form key={`${item.id}-${item.slug}`} className="admin-record" style={{ background: palette.surface }} onSubmit={(event) => void saveMod(event, item)}>
            <div className="form-two">
              <Field label={t('modSlug')}><input name="slug" required maxLength={40} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" defaultValue={item.slug} style={control} /></Field>
              <Field label={t('modName')}><input name="name" required maxLength={40} defaultValue={item.name} style={control} /></Field>
            </div>
            <div className="form-two">
              <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
              <Field label={t('modVersion')}><input name="modVersion" required maxLength={40} defaultValue={item.modVersion} style={control} /></Field>
            </div>
            <div className="form-two">
              <Field label={t('steamLink')}><input name="steamUrl" type="url" maxLength={500} defaultValue={item.steamUrl} placeholder="https://" style={control} /></Field>
              <Field label={t('githubLink')}><input name="githubUrl" type="url" maxLength={500} defaultValue={item.githubUrl} placeholder="https://" style={control} /></Field>
            </div>
            <IconPicker value={picked} onChange={(key) => setModIcons((current) => ({ ...current, [item.id]: key }))} />
            <div className="admin-actions">
              <ActionButton className="form-submit" disabled={busy}>{busy ? t('savingMod') : t('saveMod')}</ActionButton>
              <button type="button" className="danger-button" disabled={busy || mods.length < 2} onClick={() => void removeMod(item)} style={{ color: palette.danger, background: palette.hover }}>
                <Trash2 size={14} />{busy ? t('deletingMod') : t('deleteMod')}
              </button>
            </div>
          </form>
        );
      })}
      <form className="admin-record" style={{ background: palette.surface }} onSubmit={(event) => void addMod(event)}>
        <h2 style={{ color: palette.text }}>{t('addMod')}</h2>
        <div className="form-two">
          <Field label={t('modSlug')}><input name="slug" required maxLength={40} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" placeholder="rhah" value={draft.slug} onChange={(event) => setDraft((current) => ({ ...current, slug: event.target.value }))} style={control} /></Field>
          <Field label={t('modName')}><input name="name" required maxLength={40} value={draft.name} onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))} style={control} /></Field>
        </div>
        <div className="form-two">
          <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} value={draft.gameVersion} onChange={(event) => setDraft((current) => ({ ...current, gameVersion: event.target.value }))} style={control} /></Field>
          <Field label={t('modVersion')}><input name="modVersion" required maxLength={40} value={draft.modVersion} onChange={(event) => setDraft((current) => ({ ...current, modVersion: event.target.value }))} style={control} /></Field>
        </div>
        <div className="form-two">
          <Field label={t('steamLink')}><input name="steamUrl" type="url" maxLength={500} value={draft.steamUrl} placeholder="https://" onChange={(event) => setDraft((current) => ({ ...current, steamUrl: event.target.value }))} style={control} /></Field>
          <Field label={t('githubLink')}><input name="githubUrl" type="url" maxLength={500} value={draft.githubUrl} placeholder="https://" onChange={(event) => setDraft((current) => ({ ...current, githubUrl: event.target.value }))} style={control} /></Field>
        </div>
        <IconPicker value={newIcon} onChange={setNewIcon} />
        <ActionButton className="form-submit" disabled={pending === 'new'}>{pending === 'new' ? t('addingMod') : <><Plus size={14} />{t('addMod')}</>}</ActionButton>
      </form>
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
    </div>
  );
}

function FeedbackPage({ items, onSaveRecord, onDeleteRecord }: Pick<Props, 'items' | 'onSaveRecord' | 'onDeleteRecord'>) {
  const { t } = useI18n();
  const { palette } = useTheme();
  const [pending, setPending] = useState<number | null>(null);
  const [error, setError] = useState('');

  async function saveRecord(event: FormEvent<HTMLFormElement>, item: Feedback, body: string) {
    event.preventDefault();
    const length = [...body.trim()].length;
    if (length < 10 || length > 12000) {
      setError(t('bodyLength'));
      return;
    }
    const form = new FormData(event.currentTarget);
    setPending(item.id);
    setError('');
    try {
      await onSaveRecord(item, {
        title: String(form.get('title') ?? ''),
        body,
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        modList: String(form.get('modList') ?? ''),
        saveLink: String(form.get('saveLink') ?? ''),
      });
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('recordFailed'));
    } finally {
      setPending(null);
    }
  }

  async function remove(item: Feedback) {
    if (!window.confirm(t('confirmDelete'))) return;
    setPending(item.id);
    setError('');
    try {
      await onDeleteRecord(item);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('deleteFailed'));
    } finally {
      setPending(null);
    }
  }

  return (
    <div className="admin-records">
      {items.length === 0 && <p className="admin-empty" style={{ color: palette.faint }}>{t('adminEmpty')}</p>}
      {items.map((item) => (
        <AdminRecord key={`${item.id}-${item.title}-${item.body}`} item={item} busy={pending === item.id} onSave={saveRecord} onRemove={(entry) => void remove(entry)} />
      ))}
      {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
    </div>
  );
}

function AdminRecord({ item, busy, onSave, onRemove }: { item: Feedback; busy: boolean; onSave: (event: FormEvent<HTMLFormElement>, item: Feedback, body: string) => void; onRemove: (item: Feedback) => void }) {
  const { t, locale } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const control = { color: palette.text, background: palette.field };
  const [body, setBody] = useState(item.body);
  const [preview, setPreview] = useState(false);
  return (
    <form className="admin-record" style={{ background: palette.surface }} onSubmit={(event) => onSave(event, item, body)}>
      <div className="admin-record-head" style={{ color: palette.faint }}>
        <span>{labels.category(item.category)} / {(item.categoryNumber ?? item.id).toString().padStart(3, '0')}</span>
        <span>{item.author}</span>
      </div>
      <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} defaultValue={item.title} style={control} /></Field>
      <Field label={t('description')}>
        <AttachmentEditor body={body} onBody={setBody} preview={preview} rows={6} locale={locale} target={editorTarget(item.id)} control={control} onError={() => undefined} />
      </Field>
      <button type="button" className="form-switch own-edit" style={{ color: palette.gold }} onClick={() => setPreview((current) => !current)}>{preview ? t('writeMarkdown') : t('previewMarkdown')}</button>
      <div className="form-two">
        <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
        <Field label={t('modVersion')}><input name="modVersion" maxLength={80} defaultValue={item.modVersion} style={control} /></Field>
      </div>
      <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={2} defaultValue={item.modList} style={control} /></Field>
      <Field label={t('saveLink')}><input name="saveLink" maxLength={500} defaultValue={item.saveLink} style={control} /></Field>
      <div className="admin-actions">
        <ActionButton className="form-submit" disabled={busy}>{busy ? t('savingRecord') : t('saveRecord')}</ActionButton>
        <button type="button" className="danger-button" disabled={busy} onClick={() => onRemove(item)} style={{ color: palette.danger, background: palette.hover }}>
          <Trash2 size={14} />{busy ? t('deletingRecord') : t('deleteRecord')}
        </button>
      </div>
    </form>
  );
}

function IconPicker({ value, onChange }: { value: SiteIcon; onChange: (icon: SiteIcon) => void }) {
  const { t } = useI18n();
  const { palette } = useTheme();
  return (
    <fieldset className="icon-picker">
      <legend style={{ color: palette.muted }}>{t('siteIcon')}</legend>
      {siteIconKeys.map((key) => {
        const picked = value === key;
        return (
          <button
            type="button"
            key={key}
            aria-pressed={picked}
            className={picked ? 'picked' : ''}
            onClick={() => onChange(key)}
            style={{
              color: picked ? palette.activeInk : palette.muted,
              background: picked ? palette.active : palette.field,
              borderColor: 'transparent',
            }}
          >
            <span style={{ color: palette.gold }}>{siteIcons[key].icon}</span>
            {t(siteIcons[key].label)}
          </button>
        );
      })}
    </fieldset>
  );
}


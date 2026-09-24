import { useState, type FormEvent } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { siteIconKeys, siteIcons } from '../lib/icons';
import { useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Feedback, FeedbackUpdate, Mod, ModInput, SiteIcon, SiteSettings } from '../types';
import { ActionButton, Field } from './ui';

type Props = {
  items: Feedback[];
  settings: SiteSettings;
  mods: Mod[];
  onSaveSettings: (settings: SiteSettings) => Promise<void>;
  onSaveRecord: (item: Feedback, update: FeedbackUpdate) => Promise<void>;
  onDeleteRecord: (item: Feedback) => Promise<void>;
  onSaveMod: (item: Mod, input: ModInput) => Promise<void>;
  onAddMod: (input: ModInput) => Promise<void>;
  onDeleteMod: (item: Mod) => Promise<void>;
};

export function AdminPanel({ items, settings, mods, onSaveSettings, onSaveRecord, onDeleteRecord, onSaveMod, onAddMod, onDeleteMod }: Props) {
  const { t } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const control = { color: palette.text, background: palette.field, borderColor: palette.line };
  const [icon, setIcon] = useState<SiteIcon>(settings.icon);
  const [settingsPending, setSettingsPending] = useState(false);
  const [settingsError, setSettingsError] = useState('');
  const [recordPending, setRecordPending] = useState<number | null>(null);
  const [recordError, setRecordError] = useState('');
  const [modIcons, setModIcons] = useState<Record<number, SiteIcon>>({});
  const [newIcon, setNewIcon] = useState<SiteIcon>('squirrel');
  const [modPending, setModPending] = useState<number | 'new' | null>(null);
  const [modError, setModError] = useState('');

  async function saveSettings(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSettingsPending(true);
    setSettingsError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSaveSettings({
        modVersion: String(form.get('modVersion') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        icon,
      });
    } catch (problem) {
      setSettingsError(problem instanceof Error ? problem.message : t('settingsFailed'));
    } finally {
      setSettingsPending(false);
    }
  }

  async function saveRecord(event: FormEvent<HTMLFormElement>, item: Feedback) {
    event.preventDefault();
    setRecordPending(item.id);
    setRecordError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSaveRecord(item, {
        title: String(form.get('title') ?? ''),
        body: String(form.get('body') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        modList: String(form.get('modList') ?? ''),
        saveLink: String(form.get('saveLink') ?? ''),
      });
    } catch (problem) {
      setRecordError(problem instanceof Error ? problem.message : t('recordFailed'));
    } finally {
      setRecordPending(null);
    }
  }

  async function remove(item: Feedback) {
    if (!window.confirm(t('confirmDelete'))) return;
    setRecordPending(item.id);
    setRecordError('');
    try {
      await onDeleteRecord(item);
    } catch (problem) {
      setRecordError(problem instanceof Error ? problem.message : t('deleteFailed'));
    } finally {
      setRecordPending(null);
    }
  }

  function iconFor(item: Mod) {
    return modIcons[item.id] ?? item.icon;
  }

  async function saveMod(event: FormEvent<HTMLFormElement>, item: Mod) {
    event.preventDefault();
    setModPending(item.id);
    setModError('');
    const form = new FormData(event.currentTarget);
    try {
      await onSaveMod(item, {
        slug: String(form.get('slug') ?? ''),
        name: String(form.get('name') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        icon: iconFor(item),
      });
    } catch (problem) {
      setModError(problem instanceof Error ? problem.message : t('modFailed'));
    } finally {
      setModPending(null);
    }
  }

  async function addMod(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setModPending('new');
    setModError('');
    const form = new FormData(event.currentTarget);
    try {
      await onAddMod({
        slug: String(form.get('slug') ?? ''),
        name: String(form.get('name') ?? ''),
        gameVersion: String(form.get('gameVersion') ?? ''),
        modVersion: String(form.get('modVersion') ?? ''),
        icon: newIcon,
      });
      event.currentTarget.reset();
      setNewIcon('squirrel');
    } catch (problem) {
      setModError(problem instanceof Error ? problem.message : t('modFailed'));
    } finally {
      setModPending(null);
    }
  }

  async function removeMod(item: Mod) {
    if (!window.confirm(t('confirmDeleteMod'))) return;
    setModPending(item.id);
    setModError('');
    try {
      await onDeleteMod(item);
    } catch (problem) {
      setModError(problem instanceof Error ? problem.message : t('deleteModFailed'));
    } finally {
      setModPending(null);
    }
  }

  return (
    <div className="admin-panel">
      <form className="form-stack admin-settings" onSubmit={saveSettings}>
        <h3 style={{ color: palette.text }}>{t('siteSettings')}</h3>
        <div className="form-two">
          <Field label={t('currentMod')}><input name="modVersion" required maxLength={40} defaultValue={settings.modVersion} style={control} /></Field>
          <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} defaultValue={settings.gameVersion} style={control} /></Field>
        </div>
        <fieldset className="icon-picker">
          <legend style={{ color: palette.muted }}>{t('siteIcon')}</legend>
          {siteIconKeys.map((key) => {
            const picked = icon === key;
            return (
              <button
                type="button"
                key={key}
                aria-pressed={picked}
                className={picked ? 'picked' : ''}
                onClick={() => setIcon(key)}
                style={{
                  color: picked ? palette.activeInk : palette.muted,
                  background: picked ? palette.active : palette.field,
                  borderColor: picked ? palette.gold : palette.line,
                }}
              >
                <span style={{ color: palette.gold }}>{siteIcons[key].icon}</span>
                {t(siteIcons[key].label)}
              </button>
            );
          })}
        </fieldset>
        {settingsError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{settingsError}</p>}
        <ActionButton className="form-submit" disabled={settingsPending}>{settingsPending ? t('savingSettings') : t('saveSettings')}</ActionButton>
      </form>
      <div className="admin-mods">
        <h3 style={{ color: palette.text }}>{t('modListTitle')}</h3>
        {mods.length === 0 && <p className="admin-empty" style={{ color: palette.faint }}>{t('modsEmpty')}</p>}
        {mods.map((item) => {
          const pending = modPending === item.id;
          const picked = iconFor(item);
          return (
            <form key={`${item.id}-${item.slug}`} className="admin-record" style={{ borderColor: palette.line }} onSubmit={(event) => void saveMod(event, item)}>
              <div className="form-two">
                <Field label={t('modSlug')}><input name="slug" required maxLength={40} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" defaultValue={item.slug} style={control} /></Field>
                <Field label={t('modName')}><input name="name" required maxLength={40} defaultValue={item.name} style={control} /></Field>
              </div>
              <div className="form-two">
                <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
                <Field label={t('modVersion')}><input name="modVersion" required maxLength={40} defaultValue={item.modVersion} style={control} /></Field>
              </div>
              <fieldset className="icon-picker">
                <legend style={{ color: palette.muted }}>{t('siteIcon')}</legend>
                {siteIconKeys.map((key) => (
                  <button
                    type="button"
                    key={key}
                    aria-pressed={picked === key}
                    onClick={() => setModIcons((current) => ({ ...current, [item.id]: key }))}
                    style={{
                      color: picked === key ? palette.activeInk : palette.muted,
                      background: picked === key ? palette.active : palette.field,
                      borderColor: picked === key ? palette.gold : palette.line,
                    }}
                  >
                    <span style={{ color: palette.gold }}>{siteIcons[key].icon}</span>
                    {t(siteIcons[key].label)}
                  </button>
                ))}
              </fieldset>
              <div className="admin-actions">
                <ActionButton className="form-submit" disabled={pending}>{pending ? t('savingMod') : t('saveMod')}</ActionButton>
                <button type="button" className="danger-button" disabled={pending || mods.length < 2} onClick={() => void removeMod(item)} style={{ color: palette.danger, borderColor: palette.danger }}>
                  <Trash2 size={14} />{pending ? t('deletingMod') : t('deleteMod')}
                </button>
              </div>
            </form>
          );
        })}
        <form className="admin-record" style={{ borderColor: palette.line }} onSubmit={(event) => void addMod(event)}>
          <h3 style={{ color: palette.text }}>{t('addMod')}</h3>
          <div className="form-two">
            <Field label={t('modSlug')}><input name="slug" required maxLength={40} pattern="[a-z0-9]+(?:-[a-z0-9]+)*" placeholder={t('modSlugHint')} style={control} /></Field>
            <Field label={t('modName')}><input name="name" required maxLength={40} style={control} /></Field>
          </div>
          <div className="form-two">
            <Field label={t('gameVersion')}><input name="gameVersion" required maxLength={40} defaultValue="RIMWORLD 1.6" style={control} /></Field>
            <Field label={t('modVersion')}><input name="modVersion" required maxLength={40} defaultValue="DEV BUILD" style={control} /></Field>
          </div>
          <fieldset className="icon-picker">
            <legend style={{ color: palette.muted }}>{t('siteIcon')}</legend>
            {siteIconKeys.map((key) => (
              <button
                type="button"
                key={key}
                aria-pressed={newIcon === key}
                onClick={() => setNewIcon(key)}
                style={{
                  color: newIcon === key ? palette.activeInk : palette.muted,
                  background: newIcon === key ? palette.active : palette.field,
                  borderColor: newIcon === key ? palette.gold : palette.line,
                }}
              >
                <span style={{ color: palette.gold }}>{siteIcons[key].icon}</span>
                {t(siteIcons[key].label)}
              </button>
            ))}
          </fieldset>
          <ActionButton className="form-submit" disabled={modPending === 'new'}>{modPending === 'new' ? t('addingMod') : <><Plus size={14} />{t('addMod')}</>}</ActionButton>
        </form>
        {modError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{modError}</p>}
      </div>
      <div className="admin-records">
        <h3 style={{ color: palette.text }}>{t('editRecord')}</h3>
        {items.length === 0 && <p className="admin-empty" style={{ color: palette.faint }}>{t('adminEmpty')}</p>}
        {items.map((item) => {
          const pending = recordPending === item.id;
          return (
            <form key={`${item.id}-${item.title}-${item.body}`} className="admin-record" style={{ borderColor: palette.line }} onSubmit={(event) => void saveRecord(event, item)}>
              <div className="admin-record-head" style={{ color: palette.faint }}>
                <span>{labels.category(item.category)} / {(item.categoryNumber ?? item.id).toString().padStart(3, '0')}</span>
                <span>{item.author}</span>
              </div>
              <Field label={t('title')}><input name="title" required minLength={5} maxLength={120} defaultValue={item.title} style={control} /></Field>
              <Field label={t('description')}><textarea name="body" required minLength={10} maxLength={12000} rows={4} defaultValue={item.body} style={control} /></Field>
              <div className="form-two">
                <Field label={t('gameVersionField')}><input name="gameVersion" maxLength={40} defaultValue={item.gameVersion} style={control} /></Field>
                <Field label={t('modVersion')}><input name="modVersion" maxLength={80} defaultValue={item.modVersion} style={control} /></Field>
              </div>
              <Field label={t('modList')}><textarea name="modList" maxLength={6000} rows={2} defaultValue={item.modList} style={control} /></Field>
              <Field label={t('saveLink')}><input name="saveLink" maxLength={500} defaultValue={item.saveLink} style={control} /></Field>
              <div className="admin-actions">
                <ActionButton className="form-submit" disabled={pending}>{pending ? t('savingRecord') : t('saveRecord')}</ActionButton>
                <button type="button" className="danger-button" disabled={pending} onClick={() => void remove(item)} style={{ color: palette.danger, borderColor: palette.danger }}>
                  <Trash2 size={14} />{pending ? t('deletingRecord') : t('deleteRecord')}
                </button>
              </div>
            </form>
          );
        })}
        {recordError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{recordError}</p>}
      </div>
    </div>
  );
}

import { useEffect, useState } from 'react';
import { AlertTriangle, Check, ShieldCheck, ExternalLink, X } from 'lucide-react';
import { createFeedback, createMod, currentUser, deleteFeedback, deleteMod, listFeedback, listMods, login, logout, register, siteSettings, updateFeedback, updateMod, updateSiteSettings, updateStatus } from './api/feedback';
import { AdminPanel } from './components/AdminPanel';
import { Detail } from './components/Detail';
import { FeedbackForm } from './components/FeedbackForm';
import { FeedbackList } from './components/FeedbackList';
import { LoginForm } from './components/LoginForm';
import { Modal } from './components/Modal';
import { Sidebar } from './components/Sidebar';
import { Topbar } from './components/Topbar';
import { Notice } from './components/ui';
import { useI18n } from './i18n/context';
import { defaultSettings } from './lib/icons';
import { useTheme } from './theme/context';
import type { Category, Feedback, FeedbackDraft, FeedbackUpdate, Mod, ModInput, SiteSettings, Status, User } from './types';

const pageSize = 7;
const feedbackRepo = 'https://github.com/Nanaloveyuki/modders_feedback';
const defaultModSlug = 'rhah';

export function App() {
  const { locale, t } = useI18n();
  const [items, setItems] = useState<Feedback[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [filter, setFilter] = useState<Category | 'all'>('all');
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('all');
  const [sort, setSort] = useState<'new' | 'old'>('new');
  const [modal, setModal] = useState<'login' | 'register' | 'create' | 'admin' | null>(null);
  const [selected, setSelected] = useState<Feedback | null>(null);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [loading, setLoading] = useState(true);
  const { palette } = useTheme();
  const [page, setPage] = useState(1);
  const [settings, setSettings] = useState<SiteSettings>(defaultSettings);
  const [mods, setMods] = useState<Mod[]>([]);
  const [modSlug, setModSlug] = useState(defaultModSlug);

  async function refresh(slug = modSlug) {
    setLoading(true);
    try {
      const catalog = await listMods(locale).catch(() => [] as Mod[]);
      const active = catalog.find((item) => item.slug === slug) ?? catalog[0];
      const nextSlug = active?.slug ?? slug;
      const [list, current, site] = await Promise.all([
        listFeedback(nextSlug, filter, locale),
        currentUser(locale).catch(() => null),
        siteSettings(locale).catch(() => defaultSettings),
      ]);
      setMods(catalog);
      setModSlug(nextSlug);
      setItems(list);
      setUser(current);
      setSettings(site);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('loadFailed'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void refresh(); }, [filter, locale]);

  useEffect(() => { setPage(1); }, [filter, query, status]);

  const filtered = items.filter((item) => {
    const matchesText = `${item.title} ${item.body} ${item.author}`.toLowerCase().includes(query.toLowerCase());
    return matchesText && (status === 'all' || item.status === status);
  }).sort((a, b) => sort === 'new' ? b.id - a.id : a.id - b.id);
  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const shown = filtered.slice((page - 1) * pageSize, page * pageSize);
  const count = (kind: Category | 'all') => kind === 'all' ? items.length : items.filter((item) => item.category === kind).length;
  const activeMod = mods.find((item) => item.slug === modSlug) ?? mods[0];

  function openComposer() {
    setError('');
    setModal(user ? 'create' : 'login');
  }

  async function signOut() {
    await logout(locale).catch(() => undefined);
    setUser(null);
    setNotice(t('signedOut'));
  }

  async function submitLogin(username: string, password: string) {
    await login(username, password, locale);
    setModal(null);
    setError('');
    setNotice(t('signedIn'));
    await refresh();
  }

  async function submitRegister(username: string, password: string, email: string, qq: string) {
    await register(username, password, email, qq, locale);
    setModal(null);
    setError('');
    setNotice(t('registered'));
    await refresh();
  }

  async function submitFeedback(data: FeedbackDraft) {
    const created = await createFeedback(modSlug, data, locale);
    setItems((previous) => [created, ...previous]);
    setFilter(data.category);
    setStatus('all');
    setModal(null);
    setError('');
    setNotice(t('published'));
  }

  function selectMod(slug: string) {
    if (slug === modSlug) return;
    setQuery('');
    setStatus('all');
    setPage(1);
    setSelected(null);
    void refresh(slug);
  }

  async function addMod(input: ModInput) {
    const created = await createMod(input, locale);
    setMods((previous) => [...previous, created]);
    setNotice(t('modAdded'));
    selectMod(created.slug);
  }

  async function saveMod(item: Mod, input: ModInput) {
    const saved = await updateMod(item.id, input, locale);
    setMods((previous) => previous.map((entry) => entry.id === item.id ? saved : entry));
    if (item.slug === modSlug && saved.slug !== modSlug) setModSlug(saved.slug);
    setNotice(t('modSaved'));
  }

  async function removeMod(item: Mod) {
    await deleteMod(item.id, locale);
    const remaining = mods.filter((entry) => entry.id !== item.id);
    setMods(remaining);
    setNotice(t('modDeleted'));
    if (item.slug === modSlug) {
      const next = remaining[0]?.slug ?? defaultModSlug;
      setModSlug(next);
      void refresh(next);
    }
  }

  async function saveSettings(next: SiteSettings) {
    const saved = await updateSiteSettings(next, locale);
    setSettings(saved);
    setNotice(t('settingsSaved'));
  }

  async function saveRecord(item: Feedback, update: FeedbackUpdate) {
    const saved = await updateFeedback(item.id, update, locale);
    setItems((previous) => previous.map((entry) => entry.id === item.id ? saved : entry));
    setSelected((previous) => previous?.id === item.id ? saved : previous);
    setNotice(t('recordSaved'));
  }

  async function removeRecord(item: Feedback) {
    await deleteFeedback(item.id, locale);
    setItems((previous) => previous.filter((entry) => entry.id !== item.id));
    setSelected((previous) => previous?.id === item.id ? null : previous);
    setNotice(t('recordDeleted'));
  }

  async function changeStatus(item: Feedback, next: Status) {
    try {
      await updateStatus(item.id, next, locale);
      setItems((previous) => previous.map((entry) => entry.id === item.id ? { ...entry, status: next } : entry));
      setSelected((previous) => previous?.id === item.id ? { ...previous, status: next } : previous);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('statusFailed'));
    }
  }

  return (
    <>
      <Topbar user={user} icon={activeMod?.icon ?? settings.icon} name={activeMod?.name} onLogin={() => { setError(''); setModal('login'); }} onLogout={() => void signOut()} onAdmin={() => { if (user?.role === 'admin') { setError(''); setModal('admin'); } }} />
      <main id="top" className="main-layout">
        <Sidebar filter={filter} settings={settings} mods={mods} modSlug={modSlug} count={count} onFilter={setFilter} onMod={selectMod} />
        <FeedbackList
          filter={filter}
          shown={shown}
          filteredCount={filtered.length}
          query={query}
          status={status}
          sort={sort}
          page={page}
          pageSize={pageSize}
          totalPages={totalPages}
          loading={loading}
          onQuery={setQuery}
          onStatus={setStatus}
          onSort={() => setSort((current) => current === 'new' ? 'old' : 'new')}
          onPage={setPage}
          onOpen={setSelected}
          onCreate={openComposer}
        />
      </main>
      <div className="bottom-rail" style={{ background: palette.bar, borderColor: palette.line, color: palette.faint }}>
        <span><ShieldCheck size={14} style={{ color: palette.gold }} />{t('footerNote')}</span>
        <a href={feedbackRepo} target="_blank" rel="noreferrer">{t('footerRepo')} <ExternalLink size={12} /></a>
        <a href="https://steamcommunity.com/" target="_blank" rel="noreferrer">{t('steamCommunity')} <ExternalLink size={12} /></a>
      </div>
      {error && <Notice kind="error" closeLabel={t('closeNotice')} onClose={() => setError('')} closeIcon={<X size={15} />}><AlertTriangle size={16} />{error}</Notice>}
      {notice && <Notice kind="ok" closeLabel={t('closeNotice')} onClose={() => setNotice('')} closeIcon={<X size={15} />}><Check size={16} />{notice}</Notice>}
      {modal && (
        <Modal
          title={modal === 'login' ? t('loginTitle') : modal === 'register' ? t('registerTitle') : modal === 'admin' ? t('admin') : t('create')}
          subtitle={modal === 'login' ? t('loginSubtitle') : modal === 'register' ? t('registerSubtitle') : modal === 'admin' ? t('adminSubtitle') : t('createSubtitle')}
          wide={modal === 'admin'}
          onClose={() => setModal(null)}
        >
          {modal === 'login' || modal === 'register' ? <LoginForm mode={modal} onSubmit={modal === 'register' ? submitRegister : async (username, password) => submitLogin(username, password)} onSwitch={() => setModal(modal === 'register' ? 'login' : 'register')} /> : modal === 'admin' ? <AdminPanel items={items} settings={settings} mods={mods} onSaveSettings={saveSettings} onSaveRecord={saveRecord} onDeleteRecord={removeRecord} onSaveMod={saveMod} onAddMod={addMod} onDeleteMod={removeMod} /> : <FeedbackForm onSubmit={submitFeedback} />}
        </Modal>
      )}
      {selected && <Detail item={selected} canManage={user?.role === 'admin'} canEdit={Boolean(user && selected.author === user.username)} onClose={() => setSelected(null)} onStatus={changeStatus} onSave={saveRecord} />}
    </>
  );
}

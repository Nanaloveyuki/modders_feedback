import { useEffect, useState } from 'react';
import { AlertTriangle, Check, ShieldCheck, ExternalLink, X } from 'lucide-react';
import { createFeedback, currentUser, listFeedback, login, logout, updateStatus } from './api/feedback';
import { Detail } from './components/Detail';
import { FeedbackForm } from './components/FeedbackForm';
import { FeedbackList } from './components/FeedbackList';
import { LoginForm } from './components/LoginForm';
import { Modal } from './components/Modal';
import { Sidebar } from './components/Sidebar';
import { Topbar } from './components/Topbar';
import type { Category, Feedback, FeedbackDraft, Status, User } from './types';

const pageSize = 7;

export function App() {
  const [items, setItems] = useState<Feedback[]>([]);
  const [user, setUser] = useState<User | null>(null);
  const [filter, setFilter] = useState<Category | 'all'>('all');
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('all');
  const [sort, setSort] = useState<'new' | 'old'>('new');
  const [modal, setModal] = useState<'login' | 'create' | null>(null);
  const [selected, setSelected] = useState<Feedback | null>(null);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  async function refresh() {
    setLoading(true);
    try {
      const [list, current] = await Promise.all([
        listFeedback(filter),
        currentUser().catch(() => null),
      ]);
      setItems(list);
      setUser(current);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : '读取失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { void refresh(); }, [filter]);
  useEffect(() => { setPage(1); }, [filter, query, status]);

  const filtered = items.filter((item) => {
    const matchesText = `${item.title} ${item.body} ${item.author}`.toLowerCase().includes(query.toLowerCase());
    return matchesText && (status === 'all' || item.status === status);
  }).sort((a, b) => sort === 'new' ? b.id - a.id : a.id - b.id);
  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize));
  const shown = filtered.slice((page - 1) * pageSize, page * pageSize);
  const count = (kind: Category | 'all') => kind === 'all' ? items.length : items.filter((item) => item.category === kind).length;
  const openCount = items.filter((item) => item.status === 'open' || item.status === 'in_progress').length;

  function openComposer() {
    setError('');
    setModal(user ? 'create' : 'login');
  }

  async function signOut() {
    await logout().catch(() => undefined);
    setUser(null);
    setNotice('已退出登录');
  }

  async function submitLogin(username: string, password: string) {
    await login(username, password);
    setModal(null);
    setError('');
    setNotice('登录成功');
    await refresh();
  }

  async function submitFeedback(data: FeedbackDraft) {
    const created = await createFeedback(data);
    setItems((previous) => [created, ...previous]);
    setFilter(data.category);
    setStatus('all');
    setModal(null);
    setError('');
    setNotice('反馈已发布');
  }

  async function changeStatus(item: Feedback, next: Status) {
    try {
      await updateStatus(item.id, next);
      setItems((previous) => previous.map((entry) => entry.id === item.id ? { ...entry, status: next } : entry));
      setSelected((previous) => previous?.id === item.id ? { ...previous, status: next } : previous);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : '状态更新失败');
    }
  }

  return (
    <div className="app-shell">
      <Topbar user={user} onLogin={() => { setError(''); setModal('login'); }} onLogout={() => void signOut()} />
      <main id="top" className="main-layout">
        <Sidebar filter={filter} count={count} openCount={openCount} onFilter={setFilter} />
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
      <div className="bottom-rail">
        <span><ShieldCheck size={14} />仅用于本模组的反馈交流</span>
        <span>INDEPENDENT MOD · NANALOVEYUKI.RATKIN.HUNGERANDHAVOC</span>
        <a href="https://steamcommunity.com/" target="_blank" rel="noreferrer">Steam 社区 <ExternalLink size={12} /></a>
      </div>
      {error && <div className="toast toast-error" role="alert"><AlertTriangle size={16} />{error}<button aria-label="关闭通知" onClick={() => setError('')}><X size={15} /></button></div>}
      {notice && <div className="toast" role="status"><Check size={16} />{notice}<button aria-label="关闭通知" onClick={() => setNotice('')}><X size={15} /></button></div>}
      {modal && (
        <Modal
          title={modal === 'login' ? '登录反馈站' : '发布反馈'}
          subtitle={modal === 'login' ? '使用站点提供的账号登录。' : '描述你遇到的情况，帮助我们复现与验证。'}
          onClose={() => setModal(null)}
        >
          {modal === 'login' ? <LoginForm onSubmit={submitLogin} /> : <FeedbackForm onSubmit={submitFeedback} />}
        </Modal>
      )}
      {selected && <Detail item={selected} loggedIn={Boolean(user)} onClose={() => setSelected(null)} onStatus={changeStatus} />}
    </div>
  );
}

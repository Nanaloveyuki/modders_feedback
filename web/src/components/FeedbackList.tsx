import { useState } from 'react';
import { ArrowDownUp, ArrowLeft, ArrowRight, ChevronDown, Clock3, Filter, MessageSquareText, Plus, Search, X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { categoryIcons, excerpt, statusLabel, useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Category, Feedback, Status, StatusLabel, User } from '../types';
import type { Palette } from '../theme/theme';
import { ActionButton, Avatar } from './ui';

type Props = {
  filter: Category | 'all';
  shown: Feedback[];
  filteredCount: number;
  query: string;
  status: string;
  sort: 'new' | 'old';
  page: number;
  pageSize: number;
  totalPages: number;
  loading: boolean;
  onQuery: (query: string) => void;
  onStatus: (status: string) => void;
  onSort: () => void;
  onPage: (page: number) => void;
  onOpen: (item: Feedback) => void;
  onCreate: () => void;
  statuses: StatusLabel[];
  user: User | null;
  onChangeStatus: (item: Feedback, status: Status) => Promise<void>;
  onChangeStatuses: (ids: number[], status: Status) => Promise<void>;
};

function categoryColor(palette: Palette, category: Category) {
  if (category === 'bug') return palette.bug;
  if (category === 'feature') return palette.feature;
  return palette.question;
}

export function statusColors(statuses: StatusLabel[], status: Status, dark: boolean) {
  const item = statuses.find((entry) => entry.key === status);
  const background = item ? (dark ? item.dark : item.light) : (dark ? '#a8b0bc' : '#7d8794');
  return { color: inkFor(background), background };
}

function inkFor(hex: string) {
  const value = hex.replace('#', '');
  if (value.length !== 6) return '#1c1812';
  const red = Number.parseInt(value.slice(0, 2), 16);
  const green = Number.parseInt(value.slice(2, 4), 16);
  const blue = Number.parseInt(value.slice(4, 6), 16);
  const luma = (red * 299 + green * 587 + blue * 114) / 1000;
  return luma > 150 ? '#1c1812' : '#f4efe4';
}

export function FeedbackList({
  filter, shown, filteredCount, query, status, sort, page, pageSize, totalPages, loading,
  onQuery, onStatus, onSort, onPage, onOpen, onCreate, statuses, user, onChangeStatus, onChangeStatuses,
}: Props) {
  const { locale, t } = useI18n();
  const labels = useLabels(statuses);
  const { palette } = useTheme();
  const [picked, setPicked] = useState<number[]>([]);
  const [bulk, setBulk] = useState('');
  const [busy, setBusy] = useState(false);
  const heading = filter === 'all' ? t('allFeedback') : labels.category(filter);
  const narrowed = Boolean(query) || status !== 'all';
  const start = shown.length ? (page - 1) * pageSize + 1 : 0;
  const end = Math.min(page * pageSize, filteredCount);
  const field = { color: palette.text };
  const visible = shown.filter((item) => canSwitch(user, item, statuses));
  const allVisible = visible.length > 0 && visible.every((item) => picked.includes(item.id));

  function toggle(id: number) {
    setPicked((current) => current.includes(id) ? current.filter((entry) => entry !== id) : [...current, id]);
  }

  async function applyBulk() {
    if (!bulk || picked.length === 0) return;
    setBusy(true);
    try {
      await onChangeStatuses(picked, bulk);
      setPicked([]);
      setBulk('');
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="feed" aria-label={t('feedLabel')}>
      <div className="feed-heading">
        <h2 style={{ color: palette.text }}>{heading}</h2>
        <ActionButton onClick={onCreate}><Plus size={17} />{t('create')}</ActionButton>
      </div>
      <div className="toolbar">
        <label className="search-field" style={field}>
          <Search size={16} />
          <input aria-label={t('search')} value={query} onChange={(event) => onQuery(event.target.value)} style={{ color: palette.text }} />
          {query && <button aria-label={t('clearSearch')} onClick={() => onQuery('')} style={{ color: palette.muted }}><X size={13} /></button>}
        </label>
        <label className="filter-select" style={field}>
          <Filter size={15} />
          <select aria-label={t('filterStatus')} value={status} onChange={(event) => onStatus(event.target.value)} style={{ color: palette.text }}>
            <option value="all">{t('allStatus')}</option>
            {statuses.map((item) => <option key={item.key} value={item.key}>{labels.status(item.key)}</option>)}
          </select>
          <ChevronDown size={13} />
        </label>
        <button className="sort-button" onClick={onSort}>
          <ArrowDownUp size={14} />{sort === 'new' ? t('newest') : t('oldest')}
        </button>
      </div>
      <div className="list-header" style={{ color: palette.faint }}>
        <span />
        <span>{t('colIssue')}</span><span>{t('colStatus')}</span><span>{t('colAuthor')}</span><span>{t('colTime')}</span>
      </div>
      {picked.length > 0 && (
        <div className="bulk-bar" style={{ background: palette.surface, color: palette.text }}>
          <span>{t('selectedCount', { count: picked.length })}</span>
          <label className="filter-select" style={field}>
            <select aria-label={t('bulkStatus')} value={bulk} onChange={(event) => setBulk(event.target.value)} style={{ color: palette.text }}>
              <option value="">{t('chooseStatus')}</option>
              {switchable(user, statuses).map((item) => <option key={item.key} value={item.key}>{statusLabel(statuses, item.key, locale)}</option>)}
            </select>
            <ChevronDown size={13} />
          </label>
          <ActionButton disabled={busy || !bulk} onClick={() => void applyBulk()}>{t('applyStatus')}</ActionButton>
          <button type="button" className="bulk-clear" style={{ color: palette.muted }} onClick={() => setPicked([])}>{t('clearSelection')}</button>
        </div>
      )}
      <div className="feedback-list">
        {loading ? <div className="empty-state" style={{ color: palette.muted }}><span className="loading-dash" style={{ background: palette.gold }} />{t('loading')}</div> : shown.length ? shown.map((item, index) => {
          const editable = canSwitch(user, item, statuses);
          return (
          <article key={item.id} className="feedback-row" style={{ animationDelay: `${index * 36}ms`, background: palette.surface }}>
            {editable ? (
              <label className="row-check" style={{ color: palette.muted }}>
                <input
                  type="checkbox"
                  checked={picked.includes(item.id)}
                  aria-label={t('selectItem', { title: item.title })}
                  onChange={() => toggle(item.id)}
                />
              </label>
            ) : <span className="row-check" />}
            <button className="row-main" onClick={() => onOpen(item)}>
              <span className={`row-category cat-${item.category}`} style={{ color: categoryColor(palette, item.category) }}>
                {categoryIcons[item.category]}
                <span>{labels.category(item.category)} / {(item.categoryNumber ?? item.id).toString().padStart(3, '0')}</span>
              </span>
              <span className="row-title" style={{ color: palette.text }}>{item.title}</span>
              <span className="row-excerpt" style={{ color: palette.muted }}>{excerpt(item.body)}</span>
              <span className="row-meta-mobile" style={{ color: palette.faint }}>
                <StatusMenu item={item} statuses={statuses} user={user} dark={palette.dark} locale={locale} label={labels.status(item.status)} onChange={onChangeStatus} />
                <span className="row-author"><Avatar name={item.author} src={item.authorAvatar} />{item.author} · {labels.ago(item.createdAt)}</span>
              </span>
            </button>
            <StatusMenu item={item} statuses={statuses} user={user} dark={palette.dark} locale={locale} label={labels.status(item.status)} onChange={onChangeStatus} className="row-status" />
            <span className="row-author" style={{ color: palette.muted }}><Avatar name={item.author} src={item.authorAvatar} />{item.author}</span>
            <span className="row-time" style={{ color: palette.faint }}><Clock3 size={13} />{labels.ago(item.createdAt)}</span>
            <button className="row-arrow" aria-label={t('viewItem', { title: item.title })} onClick={() => onOpen(item)} style={{ color: palette.faint }}><ArrowRight size={16} /></button>
          </article>
          );
        }) : (
          <div className="empty-state" style={{ color: palette.muted }}>
            <div className="empty-glyph" style={{ color: palette.accentInk, background: palette.accent }}>
              {narrowed ? <Search size={21} /> : <MessageSquareText size={21} />}
            </div>
            <strong style={{ color: palette.text }}>{narrowed ? t('noMatch') : t('quiet')}</strong>
            {!narrowed && (
              <button onClick={onCreate} style={{ color: palette.goldInk, background: palette.accent }}>
                <Plus size={15} />{t('firstItem')}
              </button>
            )}
          </div>
        )}
      </div>
      <footer className="feed-footer" style={{ color: palette.faint }}>
        <span className="footer-range">
          {visible.length > 0 && (
            <label className="row-check" style={{ color: palette.muted }}>
              <input type="checkbox" checked={allVisible} aria-label={t('selectPage')} onChange={() => setPicked(allVisible ? picked.filter((id) => !visible.some((item) => item.id === id)) : [...new Set([...picked, ...visible.map((item) => item.id)])])} />
            </label>
          )}
          <span>{t('range', { start, end, total: filteredCount })}</span>
        </span>
        <div className="pagination" style={{ color: palette.muted }}>
          <button aria-label={t('prevPage')} disabled={page === 1} onClick={() => onPage(page - 1)} style={{ color: palette.text, background: palette.surface }}>
            <ArrowLeft size={15} />
          </button>
          <span>{page.toString().padStart(2, '0')} <i>/</i> {totalPages.toString().padStart(2, '0')}</span>
          <button aria-label={t('nextPage')} disabled={page >= totalPages} onClick={() => onPage(page + 1)} style={{ color: palette.text, background: palette.surface }}>
            <ArrowRight size={15} />
          </button>
        </div>
      </footer>
    </section>
  );
}

function StatusMenu({ item, statuses, user, dark, locale, label, onChange, className = '' }: {
  item: Feedback;
  statuses: StatusLabel[];
  user: User | null;
  dark: boolean;
  locale: 'zh' | 'en';
  label: string;
  onChange: (item: Feedback, status: Status) => Promise<void>;
  className?: string;
}) {
  const choices = choicesFor(user, item, statuses);
  const pill = <span className={`status-pill ${className}`.trim()} style={statusColors(statuses, item.status, dark)}><i />{label}</span>;
  if (choices.length < 2 || !choices.some((entry) => entry.key === item.status)) return pill;
  return (
    <label className={`status-menu ${className}`.trim()}>
      {pill}
      <select aria-label={label} value={item.status} onChange={(event) => void onChange(item, event.target.value)}>
        {choices.map((entry) => <option key={entry.key} value={entry.key}>{statusLabel(statuses, entry.key, locale)}</option>)}
      </select>
      <ChevronDown size={12} />
    </label>
  );
}

function choicesFor(user: User | null, item: Feedback, statuses: StatusLabel[]) {
  if (!user) return [];
  if (user.role === 'admin') return statuses;
  if (user.username !== item.author) return [];
  const current = statuses.find((entry) => entry.key === item.status);
  if (!current?.author) return [];
  return statuses.filter((entry) => entry.author);
}

function canSwitch(user: User | null, item: Feedback, statuses: StatusLabel[]) {
  return choicesFor(user, item, statuses).length > 1;
}

function switchable(user: User | null, statuses: StatusLabel[]) {
  if (!user) return [];
  if (user.role === 'admin') return statuses;
  return statuses.filter((entry) => entry.author);
}

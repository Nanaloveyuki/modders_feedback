import { ArrowDownUp, ArrowLeft, ArrowRight, ChevronDown, Clock3, Filter, MessageSquareText, Plus, Search, X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { categoryIcons, excerpt, statusKeys, statusStyle, useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Category, Feedback, Status } from '../types';
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
};

function categoryColor(palette: Palette, category: Category) {
  if (category === 'bug') return palette.bug;
  if (category === 'feature') return palette.feature;
  return palette.question;
}

export function statusColors(palette: Palette, status: Status) {
  const color = status === 'open' ? palette.open
    : status === 'in_progress' ? palette.progress
      : status === 'resolved' ? palette.resolved
        : status === 'withdrawn' ? palette.danger
          : palette.closed;
  return { color, background: palette.accent };
}

export function FeedbackList({
  filter, shown, filteredCount, query, status, sort, page, pageSize, totalPages, loading,
  onQuery, onStatus, onSort, onPage, onOpen, onCreate,
}: Props) {
  const { t } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const heading = filter === 'all' ? t('allFeedback') : labels.category(filter);
  const narrowed = Boolean(query) || status !== 'all';
  const start = shown.length ? (page - 1) * pageSize + 1 : 0;
  const end = Math.min(page * pageSize, filteredCount);
  const field = { color: palette.text };

  return (
    <section className="feed" aria-label={t('feedLabel')}>
      <div className="feed-heading">
        <h2 style={{ color: palette.text }}>{heading}</h2>
        <ActionButton onClick={onCreate}><Plus size={17} />{t('create')}</ActionButton>
      </div>
      <div className="toolbar">
        <label className="search-field" style={field}>
          <Search size={16} />
          <input aria-label={t('search')} placeholder={t('searchPlaceholder')} value={query} onChange={(event) => onQuery(event.target.value)} style={{ color: palette.text }} />
          {query && <button aria-label={t('clearSearch')} onClick={() => onQuery('')} style={{ color: palette.muted }}><X size={13} /></button>}
        </label>
        <label className="filter-select" style={field}>
          <Filter size={15} />
          <select aria-label={t('filterStatus')} value={status} onChange={(event) => onStatus(event.target.value)} style={{ color: palette.text }}>
            <option value="all">{t('allStatus')}</option>
            {statusKeys.map((key) => <option key={key} value={key}>{labels.status(key)}</option>)}
          </select>
          <ChevronDown size={13} />
        </label>
        <button className="sort-button" onClick={onSort}>
          <ArrowDownUp size={14} />{sort === 'new' ? t('newest') : t('oldest')}
        </button>
      </div>
      <div className="list-header" style={{ color: palette.faint }}>
        <span>{t('colIssue')}</span><span>{t('colStatus')}</span><span>{t('colAuthor')}</span><span>{t('colTime')}</span>
      </div>
      <div className="feedback-list">
        {loading ? <div className="empty-state" style={{ color: palette.muted }}><span className="loading-dash" style={{ background: palette.gold }} />{t('loading')}</div> : shown.length ? shown.map((item, index) => (
          <article key={item.id} className="feedback-row" style={{ animationDelay: `${index * 36}ms`, background: palette.surface }}>
            <button className="row-main" onClick={() => onOpen(item)}>
              <span className={`row-category cat-${item.category}`} style={{ color: categoryColor(palette, item.category) }}>
                {categoryIcons[item.category]}
                <span>{labels.category(item.category)} / {(item.categoryNumber ?? item.id).toString().padStart(3, '0')}</span>
              </span>
              <span className="row-title" style={{ color: palette.text }}>{item.title}</span>
              <span className="row-excerpt" style={{ color: palette.muted }}>{excerpt(item.body)}</span>
              <span className="row-meta-mobile" style={{ color: palette.faint }}>
                <span className={`status-pill ${statusStyle[item.status]}`} style={statusColors(palette, item.status)}><i />{labels.status(item.status)}</span>
                <span>{item.author} · {labels.ago(item.createdAt)}</span>
              </span>
            </button>
            <span className={`row-status status-pill ${statusStyle[item.status]}`} style={statusColors(palette, item.status)}><i />{labels.status(item.status)}</span>
            <span className="row-author" style={{ color: palette.muted }}><Avatar name={item.author} />{item.author}</span>
            <span className="row-time" style={{ color: palette.faint }}><Clock3 size={13} />{labels.ago(item.createdAt)}</span>
            <button className="row-arrow" aria-label={t('viewItem', { title: item.title })} onClick={() => onOpen(item)} style={{ color: palette.faint }}><ArrowRight size={16} /></button>
          </article>
        )) : (
          <div className="empty-state" style={{ color: palette.muted }}>
            <div className="empty-glyph" style={{ color: palette.accentInk, background: palette.accent }}>
              {narrowed ? <Search size={21} /> : <MessageSquareText size={21} />}
            </div>
            <strong style={{ color: palette.text }}>{narrowed ? t('noMatch') : t('quiet')}</strong>
            <span style={{ color: palette.faint }}>{narrowed ? t('noMatchHint') : t('quietHint')}</span>
            {!narrowed && (
              <button onClick={onCreate} style={{ color: palette.goldInk, background: palette.accent }}>
                <Plus size={15} />{t('firstItem')}
              </button>
            )}
          </div>
        )}
      </div>
      <footer className="feed-footer" style={{ color: palette.faint }}>
        <span>{t('range', { start, end, total: filteredCount })}</span>
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

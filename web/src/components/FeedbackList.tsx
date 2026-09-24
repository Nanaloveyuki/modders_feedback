import { ArrowDownUp, ArrowLeft, ArrowRight, ChevronDown, Clock3, Filter, MessageSquareText, Plus, Search, X } from 'lucide-react';
import type { Category, Feedback, Status } from '../types';
import { ago, categories, excerpt, statusStyle, statusText } from '../lib/labels';

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

export function FeedbackList({
  filter, shown, filteredCount, query, status, sort, page, pageSize, totalPages, loading,
  onQuery, onStatus, onSort, onPage, onOpen, onCreate,
}: Props) {
  const heading = filter === 'all' ? '全部反馈' : categories[filter].plural;
  const narrowed = Boolean(query) || status !== 'all';
  const start = shown.length ? (page - 1) * pageSize + 1 : 0;

  return (
    <section className="feed" aria-label="社区反馈">
      <div className="feed-heading">
        <div>
          <div className="section-kicker"><span className="kicker-line" />社区讨论<span className="kicker-en">COMMUNITY LOG</span></div>
          <h2>{heading}<span className="heading-count">{filteredCount.toString().padStart(2, '0')}</span></h2>
        </div>
        <button className="create-button" onClick={onCreate}><Plus size={17} />发布反馈</button>
      </div>
      <div className="toolbar">
        <label className="search-field">
          <Search size={16} />
          <input aria-label="搜索反馈" placeholder="搜索反馈内容…" value={query} onChange={(event) => onQuery(event.target.value)} />
          {query && <button aria-label="清空搜索" onClick={() => onQuery('')}><X size={13} /></button>}
        </label>
        <label className="filter-select">
          <Filter size={15} />
          <select aria-label="按状态筛选" value={status} onChange={(event) => onStatus(event.target.value)}>
            <option value="all">全部状态</option>
            {(Object.keys(statusText) as Status[]).map((key) => <option key={key} value={key}>{statusText[key]}</option>)}
          </select>
          <ChevronDown size={13} />
        </label>
        <button className="sort-button" onClick={onSort}><ArrowDownUp size={14} />{sort === 'new' ? '最新优先' : '最早优先'}</button>
      </div>
      <div className="list-header"><span>讨论 / ISSUE</span><span>状态</span><span>作者</span><span>时间</span></div>
      <div className="feedback-list">
        {loading ? <div className="empty-state"><span className="loading-dash" />正在载入反馈记录</div> : shown.length ? shown.map((item, index) => (
          <article key={item.id} className="feedback-row" style={{ animationDelay: `${index * 36}ms` }}>
            <button className="row-main" onClick={() => onOpen(item)}>
              <span className={`row-category cat-${item.category}`}>{categories[item.category].icon}<span>{categories[item.category].label} / {(item.categoryNumber ?? item.id).toString().padStart(3, '0')}</span></span>
              <span className="row-title">{item.title}</span>
              <span className="row-excerpt">{excerpt(item.body)}</span>
              <span className="row-meta-mobile">
                <span className={`status-pill ${statusStyle[item.status]}`}><i />{statusText[item.status]}</span>
                <span>{item.author} · {ago(item.createdAt)}</span>
              </span>
            </button>
            <span className={`row-status status-pill ${statusStyle[item.status]}`}><i />{statusText[item.status]}</span>
            <span className="row-author"><span className="author-avatar">{item.author.slice(0, 1).toUpperCase()}</span>{item.author}</span>
            <span className="row-time"><Clock3 size={13} />{ago(item.createdAt)}</span>
            <button className="row-arrow" aria-label={`查看：${item.title}`} onClick={() => onOpen(item)}><ArrowRight size={16} /></button>
          </article>
        )) : (
          <div className="empty-state">
            <div className="empty-glyph">{narrowed ? <Search size={21} /> : <MessageSquareText size={21} />}</div>
            <strong>{narrowed ? '没有匹配的记录' : '这里还很安静'}</strong>
            <span>{narrowed ? '换个关键词或状态试试。' : '发现第一处异常，或想到一项改进？把线索写下来。'}</span>
            {!narrowed && <button onClick={onCreate}><Plus size={15} />添加第一条反馈</button>}
          </div>
        )}
      </div>
      <footer className="feed-footer">
        <span>显示 <strong>{start}–{Math.min(page * pageSize, filteredCount)}</strong> 条，共 <strong>{filteredCount}</strong> 条记录</span>
        <div className="pagination">
          <button aria-label="上一页" disabled={page === 1} onClick={() => onPage(page - 1)}><ArrowLeft size={15} /></button>
          <span>{page.toString().padStart(2, '0')} <i>/</i> {totalPages.toString().padStart(2, '0')}</span>
          <button aria-label="下一页" disabled={page >= totalPages} onClick={() => onPage(page + 1)}><ArrowRight size={15} /></button>
        </div>
      </footer>
    </section>
  );
}

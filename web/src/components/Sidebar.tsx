import { Clipboard } from 'lucide-react';
import type { Category } from '../types';
import { categories, categoryKeys } from '../lib/labels';

type Props = {
  filter: Category | 'all';
  count: (kind: Category | 'all') => number;
  openCount: number;
  onFilter: (filter: Category | 'all') => void;
};

export function Sidebar({ filter, count, openCount, onFilter }: Props) {
  return (
    <aside className="sidebar">
      <div className="project-block">
        <div className="project-eyebrow">MOD FEEDBACK / PLAYER NOTES</div>
      </div>
      <div className="side-rule" />
      <div className="side-label">反馈板块</div>
      <nav className="category-nav" aria-label="反馈分类">
        <button className={`category-link ${filter === 'all' ? 'active' : ''}`} onClick={() => onFilter('all')}>
          <span className="category-symbol"><Clipboard size={16} /></span>
          <span>全部反馈</span>
          <span className="category-count">{count('all')}</span>
        </button>
        {categoryKeys.map((key) => (
          <button key={key} className={`category-link ${filter === key ? 'active' : ''}`} onClick={() => onFilter(key)}>
            <span className={`category-symbol icon-${key}`}>{categories[key].icon}</span>
            <span>{categories[key].label}</span>
            <span className="category-count">{count(key)}</span>
          </button>
        ))}
      </nav>
      <div className="side-rule lower-rule" />
      <div className="field-notes">
        <div className="notes-title"><span>野外记录</span><span className="notes-live"><i />LIVE</span></div>
        <div className="note-row"><span>开放中的线索</span><strong>{openCount.toString().padStart(2, '0')}</strong></div>
        <div className="note-row"><span>当前模组版本</span><strong>DEV BUILD</strong></div>
        <div className="note-row"><span>游戏版本</span><strong>RIMWORLD 1.6</strong></div>
      </div>
      <div className="side-footer"><span>RHAH / COMMUNITY ARCHIVE</span><span>© 2026</span></div>
    </aside>
  );
}

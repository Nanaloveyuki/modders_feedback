import { useEffect } from 'react';
import { Clipboard, ExternalLink, X } from 'lucide-react';
import type { Feedback, Status } from '../types';
import { categories, statusStyle, statusText } from '../lib/labels';

type Props = {
  item: Feedback;
  loggedIn: boolean;
  onClose: () => void;
  onStatus: (item: Feedback, status: Status) => Promise<void>;
};

export function Detail({ item, loggedIn, onClose, onStatus }: Props) {
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

  return (
    <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <article className="modal-panel detail-panel" role="dialog" aria-modal="true" aria-labelledby="detail-title">
        <div className="detail-top">
          <span className={`row-category cat-${item.category}`}>{categories[item.category].icon}<span>{categories[item.category].label}</span></span>
          <button className="icon-button" aria-label="关闭" onClick={onClose}><X size={18} /></button>
        </div>
        <h2 id="detail-title" className="detail-title">{item.title}</h2>
        <div className="detail-byline">
          <span className="author-avatar">{item.author.slice(0, 1).toUpperCase()}</span>
          <strong>{item.author}</strong>
          <span>·</span>
          <span>{new Date(item.createdAt).toLocaleString('zh-CN')}</span>
        </div>
        <div className="detail-body">{item.body}</div>
        <div className="detail-facts">
          <div><span>RIMWORLD</span><strong>{item.gameVersion || '未提供'}</strong></div>
          <div><span>MOD VERSION</span><strong>{item.modVersion || '未提供'}</strong></div>
          <div><span>STATUS</span><strong className={`status-pill ${statusStyle[item.status]}`}><i />{statusText[item.status]}</strong></div>
        </div>
        {item.modList && <div className="detail-extra"><h3>相关模组 / 加载顺序</h3><pre>{item.modList}</pre></div>}
        {item.saveLink && <a className="save-link" href={item.saveLink} target="_blank" rel="noreferrer"><ExternalLink size={14} />打开存档分享链接</a>}
        {loggedIn && (
          <label className="status-editor">
            更新处理状态
            <select value={item.status} onChange={(event) => void onStatus(item, event.target.value as Status)}>
              <option value="open">待处理</option>
              <option value="in_progress">处理中</option>
              <option value="resolved">已解决</option>
              <option value="closed">已关闭</option>
            </select>
          </label>
        )}
        <div className="detail-bottom">
          <span>ISSUE / {String(item.id).padStart(4, '0')}</span>
          <button onClick={() => { void navigator.clipboard?.writeText(window.location.href); }}><Clipboard size={14} />复制页面地址</button>
        </div>
      </article>
    </div>
  );
}

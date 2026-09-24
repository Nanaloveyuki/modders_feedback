import { useEffect } from 'react';
import { Clipboard, ExternalLink, X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { categoryIcons, statusKeys, statusStyle, useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Feedback, Status } from '../types';
import { statusColors } from './FeedbackList';
import { Avatar, IconButton, Panel, Scrim } from './ui';

type Props = {
  item: Feedback;
  loggedIn: boolean;
  onClose: () => void;
  onStatus: (item: Feedback, status: Status) => Promise<void>;
};

export function Detail({ item, loggedIn, onClose, onStatus }: Props) {
  const { t } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const categoryColor = item.category === 'bug' ? palette.bug : item.category === 'feature' ? palette.feature : palette.question;
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

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
        {loggedIn && (
          <label className="status-editor" style={{ color: palette.muted }}>
            {t('updateStatus')}
            <select value={item.status} onChange={(event) => void onStatus(item, event.target.value as Status)} style={{ color: palette.text, background: palette.field, borderColor: palette.line }}>
              {statusKeys.map((key) => <option key={key} value={key}>{labels.status(key)}</option>)}
            </select>
          </label>
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

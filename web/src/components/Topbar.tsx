import { ExternalLink, Github, LockKeyhole, LogOut, Moon, Settings, Sun, UserRound } from 'lucide-react';
import { nextLocale, useI18n } from '../i18n/context';
import { siteIcons } from '../lib/icons';
import { navigate } from '../router';
import { useTheme } from '../theme/context';
import type { SiteIcon, User } from '../types';
import { IconButton } from './ui';

type Props = {
  user: User | null;
  icon: SiteIcon;
  name?: string;
  onLogin: () => void;
  onLogout: () => void;
  onAdmin: () => void;
};

export function Topbar({ user, icon, name, onLogin, onLogout, onAdmin }: Props) {
  const { locale, setLocale, t } = useI18n();
  const { dark, palette, toggle } = useTheme();
  const themeLabel = dark ? t('themeToLight') : t('themeToDark');
  return (
    <header className="topbar" style={{ background: palette.surface, color: palette.muted }}>
      <a className="brand" href="/" aria-label={t('brandHome')}>
        <span className="brand-mark" style={{ color: palette.accentInk, background: palette.accent }}>
          {siteIcons[icon].icon}
        </span>
        <span className="brand-name" style={{ color: palette.text }}>
          {name ?? (locale === 'zh' ? <>鼠族<span style={{ color: palette.gold }}>：</span>饥与祸</> : <>Ratkin<span style={{ color: palette.gold }}>:</span> Hunger and Havoc</>)}
          <small style={{ color: palette.faint }}>{t('brandSmall')}</small>
        </span>
      </a>
      <nav className="top-links" aria-label={t('navLabel')} style={{ color: palette.muted }}>
        <span className="build-tag" style={{ color: palette.gold }}>
          <span className="live-dot" style={{ background: palette.gold, boxShadow: `0 0 11px ${palette.gold}` }} />
          RIMWORLD 1.6
        </span>
        <span className="nav-divider" style={{ background: palette.line }} />
        <a href="https://steamcommunity.com/sharedfiles/filedetails/?id=" target="_blank" rel="noreferrer">{t('workshop')} <ExternalLink size={13} /></a>
        <a href="https://github.com/Nanaloveyuki/modders_feedback" target="_blank" rel="noreferrer" aria-label={t('projectHome')}><Github size={15} /></a>
      </nav>
      <div className="account-area" style={{ color: palette.muted }}>
        <button
          className="lang-toggle"
          aria-label={t('language')}
          style={{ color: palette.text, background: palette.hover }}
          onClick={() => setLocale(nextLocale(locale))}
        >
          {locale === 'zh' ? 'EN' : '中文'}
        </button>
        <IconButton className="theme-toggle" label={themeLabel} tone="gold" onClick={toggle}>
          {dark ? <Sun size={17} /> : <Moon size={17} />}
        </IconButton>
        {user ? (
          <>
            {user.role === 'admin' && <IconButton label={t('adminOpen')} tone="gold" onClick={onAdmin}><Settings size={17} /></IconButton>}
            <span className="user-avatar" style={{ color: palette.goldInk, background: palette.accent }}>
              {user.avatarUrl ? <img src={user.avatarUrl} alt="" /> : user.username.slice(0, 1).toUpperCase()}
            </span>
            <span className="account-name">{user.username}</span>
            <IconButton label={t('accountOpen')} onClick={() => navigate('/account')}><UserRound size={16} /></IconButton>
            <IconButton label={t('logout')} onClick={onLogout}><LogOut size={16} /></IconButton>
          </>
        ) : (
          <button className="login-button" style={{ color: palette.text, background: palette.hover }} onClick={onLogin}>
            <LockKeyhole size={14} /> {t('login')}
          </button>
        )}
      </div>
    </header>
  );
}

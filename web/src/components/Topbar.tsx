import { ExternalLink, LockKeyhole, LogOut, Moon, Settings, Sun, UserRound } from 'lucide-react';
import { nextLocale, useI18n } from '../i18n/context';
import { siteIcons } from '../lib/icons';
import { navigate } from '../router';
import { useTheme } from '../theme/context';
import type { Mod, SiteIcon, User } from '../types';
import { IconButton } from './ui';

/*!
 * GitHub paths retained from lucide-react 0.468.0 (ISC License).
 * Copyright (c) for portions of Lucide are held by Cole Bemis 2013-2022 as part of
 * Feather (MIT). All other copyright (c) for Lucide are held by Lucide Contributors 2022.
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */
const githubIcon = (
  <svg xmlns="http://www.w3.org/2000/svg" width={15} height={15} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-github" aria-hidden="true" focusable="false">
    <path d="M15 22v-4a4.8 4.8 0 0 0-1-3.5c3 0 6-2 6-5.5.08-1.25-.27-2.48-1-3.5.28-1.15.28-2.35 0-3.5 0 0-1 0-3 1.5-2.64-.5-5.36-.5-8 0C6 2 5 2 5 2c-.3 1.15-.3 2.35 0 3.5A5.403 5.403 0 0 0 4 9c0 3.5 3 5.5 6 5.5-.39.49-.68 1.05-.85 1.65-.17.6-.22 1.23-.15 1.85v4" />
    <path d="M9 18c-4.51 2-5-2-7-2" />
  </svg>
);

type Props = {
  user: User | null;
  mod?: Mod;
  icon: SiteIcon;
  name?: string;
  onLogin: () => void;
  onLogout: () => void;
  onAdmin: () => void;
};

export function Topbar({ user, mod, icon, name, onLogin, onLogout, onAdmin }: Props) {
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
          {mod?.gameVersion || 'RIMWORLD 1.6'}
        </span>
        {(mod?.steamUrl || mod?.githubUrl) && <span className="nav-divider" style={{ background: palette.line }} />}
        {mod?.steamUrl && <a href={mod.steamUrl} target="_blank" rel="noreferrer">{t('workshop')} <ExternalLink size={13} /></a>}
        {mod?.githubUrl && <a href={mod.githubUrl} target="_blank" rel="noreferrer" aria-label={t('projectHome')}>{githubIcon}</a>}
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

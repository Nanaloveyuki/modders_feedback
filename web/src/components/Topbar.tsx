import { ExternalLink, Github, LockKeyhole, LogOut, Moon, Sun, Squirrel } from 'lucide-react';
import type { User } from '../types';

type Props = {
  user: User | null;
  theme: 'light' | 'dark';
  onThemeChange: () => void;
  onLogin: () => void;
  onLogout: () => void;
};

export function Topbar({ user, theme, onThemeChange, onLogin, onLogout }: Props) {
  return (
    <header className="topbar">
      <a className="brand" href="#top" aria-label="鼠族：饥与祸反馈站首页">
        <span className="brand-mark"><Squirrel size={22} strokeWidth={1.8} /></span>
        <span className="brand-name">鼠族<span>：</span>饥与祸<small>社区反馈站</small></span>
      </a>
      <nav className="top-links" aria-label="模组信息">
        <span className="build-tag"><span className="live-dot" />RIMWORLD 1.6</span>
        <span className="nav-divider" />
        <a href="https://steamcommunity.com/sharedfiles/filedetails/?id=" target="_blank" rel="noreferrer">Steam 创意工坊 <ExternalLink size={13} /></a>
        <a href="https://github.com/" target="_blank" rel="noreferrer" aria-label="项目主页"><Github size={15} /></a>
      </nav>
      <div className="account-area">
        <button className="icon-button theme-toggle" title={theme === 'dark' ? '切换到日间模式' : '切换到夜间模式'} aria-label={theme === 'dark' ? '切换到日间模式' : '切换到夜间模式'} onClick={onThemeChange}>{theme === 'dark' ? <Sun size={17} /> : <Moon size={17} />}</button>
        {user ? <><span className="user-avatar">{user.username.slice(0, 1).toUpperCase()}</span><span className="account-name">{user.username}</span><button className="icon-button" title="退出登录" onClick={onLogout}><LogOut size={17} /></button></> : <button className="login-button" onClick={onLogin}><LockKeyhole size={14} /> 登录</button>}
      </div>
    </header>
  );
}

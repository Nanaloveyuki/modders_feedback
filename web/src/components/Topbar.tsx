import { ExternalLink, Github, LockKeyhole, LogOut, Squirrel } from 'lucide-react';
import type { User } from '../types';

type Props = {
  user: User | null;
  onLogin: () => void;
  onLogout: () => void;
};

export function Topbar({ user, onLogin, onLogout }: Props) {
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
        {user ? (
          <>
            <span className="user-avatar">{user.username.slice(0, 1).toUpperCase()}</span>
            <span className="account-name">{user.username}</span>
            <button className="icon-button" title="退出登录" onClick={onLogout}><LogOut size={17} /></button>
          </>
        ) : (
          <button className="login-button" onClick={onLogin}><LockKeyhole size={14} /> 登录</button>
        )}
      </div>
    </header>
  );
}

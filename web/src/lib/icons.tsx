import type { ReactNode } from 'react';
import { Bug, PawPrint, Rat, Shield, Sparkles, Squirrel } from 'lucide-react';
import type { MessageKey } from '../i18n/messages';
import type { SiteIcon } from '../types';

export const siteIcons: Record<SiteIcon, { label: MessageKey; icon: ReactNode }> = {
  squirrel: { label: 'iconSquirrel', icon: <Squirrel size={18} strokeWidth={1.8} /> },
  rat: { label: 'iconRat', icon: <Rat size={18} strokeWidth={1.8} /> },
  bug: { label: 'iconBug', icon: <Bug size={18} strokeWidth={1.8} /> },
  spark: { label: 'iconSpark', icon: <Sparkles size={18} strokeWidth={1.8} /> },
  shield: { label: 'iconShield', icon: <Shield size={18} strokeWidth={1.8} /> },
  paw: { label: 'iconPaw', icon: <PawPrint size={18} strokeWidth={1.8} /> },
};

export const siteIconKeys = Object.keys(siteIcons) as SiteIcon[];

export const defaultSettings = {
  modVersion: 'DEV BUILD',
  gameVersion: 'RIMWORLD 1.6',
  icon: 'squirrel' as SiteIcon,
};

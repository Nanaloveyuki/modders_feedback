import { useMemo, useState } from 'react';
import { Clipboard, Search, X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { categoryIcons, categoryKeys, useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Category, Mod, SiteSettings } from '../types';
import { siteIcons } from '../lib/icons';
import { TextButton } from './ui';

type Props = {
  filter: Category | 'all';
  settings: SiteSettings;
  mods: Mod[];
  modSlug: string;
  count: (kind: Category | 'all') => number;
  onFilter: (filter: Category | 'all') => void;
  onMod: (slug: string) => void;
};

export function Sidebar({ filter, settings, mods, modSlug, count, onFilter, onMod }: Props) {
  const { t } = useI18n();
  const labels = useLabels();
  const { palette } = useTheme();
  const [modQuery, setModQuery] = useState('');
  const symbol = (key: Category | 'all') => {
    if (key === 'bug') return palette.bug;
    if (key === 'feature') return palette.feature;
    if (key === 'question') return palette.question;
    return palette.muted;
  };
  const active = mods.find((item) => item.slug === modSlug);
  const needle = modQuery.trim().toLowerCase();
  const visibleMods = useMemo(() => {
    if (!needle) return mods;
    return mods.filter((mod) => `${mod.name} ${mod.slug}`.toLowerCase().includes(needle));
  }, [mods, needle]);

  return (
    <aside className="sidebar">
      <section className="side-card mod-card">
        <div className="side-label" style={{ color: palette.faint }}>{t('mods')}</div>
        <label className="mod-search">
          <Search size={14} />
          <input
            aria-label={t('modSearch')}
            value={modQuery}
            onChange={(event) => setModQuery(event.target.value)}
          />
          {modQuery && (
            <button type="button" aria-label={t('clearModSearch')} onClick={() => setModQuery('')}>
              <X size={13} />
            </button>
          )}
        </label>
        <nav className="category-nav mod-nav" aria-label={t('modsLabel')}>
          {visibleMods.length ? visibleMods.map((mod) => (
            <TextButton key={mod.slug} active={mod.slug === modSlug} onClick={() => onMod(mod.slug)}>
              <span className="category-symbol" style={{ color: palette.gold }}>{siteIcons[mod.icon].icon}</span>
              <span className="mod-copy">
                <span className="mod-name">{mod.name}</span>
                <span className="mod-slug">{mod.slug}</span>
              </span>
            </TextButton>
          )) : (
            <p className="mod-empty" style={{ color: palette.faint }}>{t('modsNone')}</p>
          )}
        </nav>
      </section>
      <section className="side-card board-card">
        <div className="side-label" style={{ color: palette.faint }}>{t('boards')}</div>
        <nav className="category-nav" aria-label={t('boardsLabel')}>
          <TextButton active={filter === 'all'} onClick={() => onFilter('all')}>
            <span className="category-symbol" style={{ color: symbol('all') }}><Clipboard size={16} /></span>
            <span>{t('allFeedback')}</span>
            <span className="category-count" style={{ color: filter === 'all' ? palette.gold : palette.faint }}>{count('all')}</span>
          </TextButton>
          {categoryKeys.map((key) => (
            <TextButton key={key} active={filter === key} onClick={() => onFilter(key)}>
              <span className={`category-symbol icon-${key}`} style={{ color: symbol(key) }}>{categoryIcons[key]}</span>
              <span>{labels.category(key)}</span>
              <span className="category-count" style={{ color: filter === key ? palette.gold : palette.faint }}>{count(key)}</span>
            </TextButton>
          ))}
        </nav>
      </section>
      <section className="side-card note-card">
        <div className="field-notes">
          <div className="note-row" style={{ color: palette.faint }}><span>{t('currentMod')}</span><strong style={{ color: palette.muted }}>{active?.modVersion || settings.modVersion}</strong></div>
          <div className="note-row" style={{ color: palette.faint }}><span>{t('gameVersion')}</span><strong style={{ color: palette.muted }}>{active?.gameVersion || settings.gameVersion}</strong></div>
        </div>
      </section>
    </aside>
  );
}

import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { useTheme } from '../theme/context';
import type { Palette } from '../theme/theme';

type Tone = 'neutral' | 'accent' | 'gold' | 'bug' | 'feature' | 'question' | 'danger';

function toneColor(palette: Palette, tone: Tone) {
  if (tone === 'accent') return palette.accent;
  if (tone === 'gold') return palette.gold;
  if (tone === 'bug' || tone === 'danger') return palette.bug;
  if (tone === 'feature') return palette.feature;
  if (tone === 'question') return palette.question;
  return palette.muted;
}

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string;
  tone?: Tone;
};

export function IconButton({ label, tone = 'neutral', className, style, children, ...props }: IconButtonProps) {
  const { palette } = useTheme();
  const color = tone === 'neutral' ? palette.text : toneColor(palette, tone);
  return (
    <button
      {...props}
      className={className ? `icon-button ${className}` : 'icon-button'}
      aria-label={label}
      title={props.title ?? label}
      style={{ color, background: palette.surface, ...style }}
    >
      {children}
    </button>
  );
}

type ActionButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
};

export function ActionButton({ className, style, children, ...props }: ActionButtonProps) {
  const { palette } = useTheme();
  return (
    <button
      {...props}
      className={className ? `${className} action-button` : 'create-button action-button'}
      style={{ ...style }}
    >
      {children}
    </button>
  );
}

type TextButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  active?: boolean;
  children: ReactNode;
};

export function TextButton({ active = false, style, children, ...props }: TextButtonProps) {
  const { palette } = useTheme();
  return (
    <button
      {...props}
      className={active ? 'category-link active' : 'category-link'}
      style={{
        color: active ? palette.activeInk : palette.muted,
        background: active ? palette.active : 'transparent',
        ...style,
      }}
    >
      {children}
    </button>
  );
}

export function Avatar({ name }: { name: string }) {
  const { palette } = useTheme();
  return (
    <span className="author-avatar" style={{ color: palette.goldInk, background: palette.accent }}>
      {name.slice(0, 1).toUpperCase()}
    </span>
  );
}

export function Panel({
  className,
  children,
  labelledBy,
}: {
  className?: string;
  children: ReactNode;
  labelledBy: string;
}) {
  const { palette } = useTheme();
  return (
    <section
      className={className ? `modal-panel ${className}` : 'modal-panel'}
      role="dialog"
      aria-modal="true"
      aria-labelledby={labelledBy}
      style={{
        background: palette.surface,
        color: palette.text,
        boxShadow: `0 18px 48px ${palette.shadow}`,
      }}
    >
      {children}
    </section>
  );
}

export function Scrim({ onClose, children }: { onClose: () => void; children: ReactNode }) {
  const { palette } = useTheme();
  return (
    <div
      className="modal-backdrop"
      style={{ background: palette.scrim }}
      onMouseDown={(event) => { if (event.target === event.currentTarget) onClose(); }}
    >
      {children}
    </div>
  );
}

export function Field({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  const { palette } = useTheme();
  return <label style={{ color: palette.muted }}>{label}{children}</label>;
}

type NoticeProps = {
  kind: 'ok' | 'error';
  children: ReactNode;
  closeLabel: string;
  onClose: () => void;
  closeIcon: ReactNode;
};

export function Notice({ kind, children, closeLabel, onClose, closeIcon }: NoticeProps) {
  const { palette } = useTheme();
  const error = kind === 'error';
  return (
    <div
      className={error ? 'toast toast-error' : 'toast'}
      role={error ? 'alert' : 'status'}
      style={{
        color: error ? palette.danger : palette.text,
        background: palette.surface,
        boxShadow: `0 16px 40px ${palette.shadow}`,
      }}
    >
      {children}
      <button aria-label={closeLabel} onClick={onClose} style={{ color: 'inherit' }}>{closeIcon}</button>
    </div>
  );
}

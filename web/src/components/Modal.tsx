import { useEffect, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { useI18n } from '../i18n/context';
import { useTheme } from '../theme/context';
import { IconButton, Panel, Scrim } from './ui';

type Props = {
  title: string;
  onClose: () => void;
  children: ReactNode;
};

export function Modal({ title, onClose, children }: Props) {
  const { t } = useI18n();
  const { palette } = useTheme();
  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [onClose]);

  return (
    <Scrim onClose={onClose}>
      <Panel labelledBy="modal-title">
        <div className="modal-heading">
          <h2 id="modal-title" style={{ color: palette.text }}>{title}</h2>
          <IconButton className="close-button" label={t('close')} onClick={onClose}><X size={19} /></IconButton>
        </div>
        {children}
      </Panel>
    </Scrim>
  );
}

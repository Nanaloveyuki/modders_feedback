import { useEffect, useRef, useState, type PointerEvent } from 'react';
import { ArrowLeft, ImagePlus } from 'lucide-react';
import { updateAvatar, updatePassword, updateProfile } from '../api/feedback';
import { useI18n } from '../i18n/context';
import { navigate } from '../router';
import { useTheme } from '../theme/context';
import type { User } from '../types';
import { ActionButton, Field } from './ui';

const avatarSize = 256;
const maxSource = 12 * 1024 * 1024;

type Props = {
  user: User;
  onUser: (user: User) => void;
};

type Crop = { image: HTMLImageElement; scale: number; x: number; y: number };
type Drag = { x: number; y: number; originX: number; originY: number };

export function AccountPage({ user, onUser }: Props) {
  const { locale, t } = useI18n();
  const { palette } = useTheme();
  const control = { color: palette.text, background: palette.field };
  const [profile, setProfile] = useState({ username: user.username, email: user.email, qq: user.qq });
  const [profilePending, setProfilePending] = useState(false);
  const [profileError, setProfileError] = useState('');
  const [passwords, setPasswords] = useState({ current: '', next: '' });
  const [passwordPending, setPasswordPending] = useState(false);
  const [passwordError, setPasswordError] = useState('');
  const [crop, setCrop] = useState<Crop | null>(null);
  const [source, setSource] = useState('');
  const [avatarPending, setAvatarPending] = useState(false);
  const [avatarError, setAvatarError] = useState('');
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const drag = useRef<Drag | null>(null);

  useEffect(() => {
    const node = canvasRef.current;
    if (!node || !crop) return;
    const context = node.getContext('2d');
    if (!context) return;
    context.clearRect(0, 0, avatarSize, avatarSize);
    context.drawImage(crop.image, crop.x, crop.y, crop.image.width * crop.scale, crop.image.height * crop.scale);
  }, [crop]);

  async function saveProfile() {
    setProfilePending(true);
    setProfileError('');
    try {
      const saved = await updateProfile(profile, locale);
      onUser(saved);
      setProfile({ username: saved.username, email: saved.email, qq: saved.qq });
    } catch (problem) {
      setProfileError(problem instanceof Error ? problem.message : t('profileFailed'));
    } finally {
      setProfilePending(false);
    }
  }

  async function savePassword() {
    setPasswordPending(true);
    setPasswordError('');
    try {
      await updatePassword(passwords.current, passwords.next, locale);
      setPasswords({ current: '', next: '' });
    } catch (problem) {
      setPasswordError(problem instanceof Error ? problem.message : t('passwordFailed'));
    } finally {
      setPasswordPending(false);
    }
  }

  function chooseFile(file: File | undefined) {
    setAvatarError('');
    if (!file) return;
    if (!['image/png', 'image/jpeg', 'image/gif', 'image/webp'].includes(file.type)) {
      setAvatarError(t('avatarType'));
      return;
    }
    if (file.size > maxSource) {
      setAvatarError(t('avatarSourceLimit'));
      return;
    }
    const url = URL.createObjectURL(file);
    const image = new Image();
    image.onload = () => {
      URL.revokeObjectURL(url);
      const reader = new FileReader();
      reader.onload = () => {
        const encoded = String(reader.result ?? '').split(',')[1] ?? '';
        const scale = Math.max(avatarSize / image.width, avatarSize / image.height);
        setSource(encoded);
        setCrop({ image, scale, x: (avatarSize - image.width * scale) / 2, y: (avatarSize - image.height * scale) / 2 });
      };
      reader.onerror = () => setAvatarError(t('avatarType'));
      reader.readAsDataURL(file);
    };
    image.onerror = () => {
      URL.revokeObjectURL(url);
      setAvatarError(t('avatarType'));
    };
    image.src = url;
  }

  function moveCrop(event: PointerEvent<HTMLCanvasElement>) {
    const start = drag.current;
    if (!start || !crop) return;
    const width = crop.image.width * crop.scale;
    const height = crop.image.height * crop.scale;
    const x = Math.min(0, Math.max(avatarSize - width, start.originX + event.clientX - start.x));
    const y = Math.min(0, Math.max(avatarSize - height, start.originY + event.clientY - start.y));
    setCrop({ ...crop, x, y });
  }

  async function saveAvatar() {
    if (!crop || !source) return;
    setAvatarPending(true);
    setAvatarError('');
    const size = Math.round(avatarSize / crop.scale);
    const cropBox = {
      x: Math.max(0, Math.min(crop.image.width - size, Math.round(-crop.x / crop.scale))),
      y: Math.max(0, Math.min(crop.image.height - size, Math.round(-crop.y / crop.scale))),
      size,
    };
    try {
      const saved = await updateAvatar(source, cropBox, locale);
      onUser({ ...saved, avatarUrl: `${saved.avatarUrl}&v=${Date.now()}` });
      setCrop(null);
      setSource('');
    } catch (problem) {
      setAvatarError(problem instanceof Error ? problem.message : t('avatarFailed'));
    } finally {
      setAvatarPending(false);
    }
  }

  const minimum = crop ? Math.max(avatarSize / crop.image.width, avatarSize / crop.image.height) : 1;

  return (
    <section className="admin-page account-page">
      <header className="admin-page-head">
        <button type="button" className="admin-back" style={{ color: palette.gold }} onClick={() => navigate('/')}>
          <ArrowLeft size={15} />{t('adminBack')}
        </button>
        <h1 style={{ color: palette.text }}>{t('accountTitle')}</h1>
      </header>
      <form className="admin-record" style={{ background: palette.surface }} onSubmit={(event) => { event.preventDefault(); void saveProfile(); }}>
        <h2 style={{ color: palette.text }}>{t('profileTitle')}</h2>
        <Field label={t('username')}><input required minLength={3} maxLength={32} pattern="[A-Za-z0-9_-]{3,32}" value={profile.username} onChange={(event) => setProfile((current) => ({ ...current, username: event.target.value }))} style={control} /></Field>
        <div className="form-two">
          <Field label={t('emailOptional')}><input type="email" maxLength={254} value={profile.email} onChange={(event) => setProfile((current) => ({ ...current, email: event.target.value }))} style={control} /></Field>
          <Field label={t('qqOptional')}><input inputMode="numeric" pattern="[1-9][0-9]{4,10}" maxLength={11} value={profile.qq} onChange={(event) => setProfile((current) => ({ ...current, qq: event.target.value }))} style={control} /></Field>
        </div>
        {profileError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{profileError}</p>}
        <ActionButton className="form-submit" disabled={profilePending}>{profilePending ? t('savingProfile') : t('saveProfile')}</ActionButton>
      </form>
      <form className="admin-record" style={{ background: palette.surface }} onSubmit={(event) => { event.preventDefault(); void savePassword(); }}>
        <h2 style={{ color: palette.text }}>{t('passwordTitle')}</h2>
        <div className="form-two">
          <Field label={t('currentPassword')}><input type="password" autoComplete="current-password" required maxLength={128} value={passwords.current} onChange={(event) => setPasswords((current) => ({ ...current, current: event.target.value }))} style={control} /></Field>
          <Field label={t('newPassword')}><input type="password" autoComplete="new-password" required minLength={12} maxLength={128} value={passwords.next} onChange={(event) => setPasswords((current) => ({ ...current, next: event.target.value }))} style={control} /></Field>
        </div>
        {passwordError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{passwordError}</p>}
        <ActionButton className="form-submit" disabled={passwordPending}>{passwordPending ? t('savingPassword') : t('savePassword')}</ActionButton>
      </form>
      <div className="admin-record" style={{ background: palette.surface }}>
        <h2 style={{ color: palette.text }}>{t('avatarTitle')}</h2>
        <div className="avatar-editor">
          {crop ? (
            <canvas
              width={avatarSize}
              height={avatarSize}
              className="avatar-crop"
              ref={canvasRef}
              onPointerDown={(event) => {
                event.currentTarget.setPointerCapture(event.pointerId);
                drag.current = { x: event.clientX, y: event.clientY, originX: crop.x, originY: crop.y };
              }}
              onPointerMove={moveCrop}
              onPointerUp={() => { drag.current = null; }}
            />
          ) : user.avatarUrl ? (
            <img className="avatar-preview" src={user.avatarUrl} alt="" />
          ) : (
            <span className="avatar-preview avatar-letter" style={{ color: palette.danger, background: palette.hover }}>{user.username.slice(0, 1).toUpperCase()}</span>
          )}
          <div className="avatar-controls">
            <label className="create-button">
              <ImagePlus size={15} />{t('chooseAvatar')}
              <input type="file" accept="image/png,image/jpeg,image/gif,image/webp" hidden onChange={(event) => chooseFile(event.target.files?.[0])} />
            </label>
            {crop && (
              <>
                <label className="avatar-zoom" style={{ color: palette.muted }}>
                  {t('avatarZoom')}
                  <input type="range" min={minimum} max={minimum * 4} step={0.01} value={crop.scale} onChange={(event) => setCrop(clampScale(crop, Number(event.target.value)))} />
                </label>
                <ActionButton className="form-submit" disabled={avatarPending} onClick={() => void saveAvatar()}>{avatarPending ? t('savingAvatar') : t('saveAvatar')}</ActionButton>
              </>
            )}
            {avatarError && <p className="form-error" role="alert" style={{ color: palette.danger }}>{avatarError}</p>}
          </div>
        </div>
      </div>
    </section>
  );
}

function clampScale(crop: Crop, scale: number) {
  const width = crop.image.width * scale;
  const height = crop.image.height * scale;
  return { ...crop, scale, x: Math.min(0, Math.max(avatarSize - width, crop.x)), y: Math.min(0, Math.max(avatarSize - height, crop.y)) };
}

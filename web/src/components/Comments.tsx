import { useEffect, useMemo, useState } from 'react';
import { ChevronDown, MessageSquare, Quote } from 'lucide-react';
import { createComment, deleteComment, listTimeline, updateComment } from '../api/feedback';
import { useI18n } from '../i18n/context';
import { useLabels } from '../lib/labels';
import { useTheme } from '../theme/context';
import type { Feedback, StatusLabel, TimelineEvent, User } from '../types';
import { AttachmentEditor, editorTarget } from './AttachmentEditor';
import { statusColors } from './FeedbackList';
import { MarkdownBody } from './MarkdownBody';
import { ActionButton, Avatar } from './ui';

const visibleEdges = 2;

type Props = {
  item: Feedback;
  user: User | null;
  statuses: StatusLabel[];
  onLogin: () => void;
};

export function Comments({ item, user, statuses, onLogin }: Props) {
  const { t, locale } = useI18n();
  const labels = useLabels(statuses);
  const { palette } = useTheme();
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [open, setOpen] = useState(false);
  const [body, setBody] = useState('');
  const [preview, setPreview] = useState(false);
  const [editing, setEditing] = useState<number | null>(null);
  const [editBody, setEditBody] = useState('');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const control = { color: palette.text, background: palette.field };
  const canPost = Boolean(user);

  useEffect(() => {
    let cancelled = false;
    setOpen(false);
    setEvents([]);
    void listTimeline(item.id, locale).then((timeline) => {
      if (!cancelled) setEvents(timeline.events);
    }).catch((problem: unknown) => {
      if (!cancelled) setError(problem instanceof Error ? problem.message : t('commentsFailed'));
    });
    return () => { cancelled = true; };
  }, [item.id, locale]);

  const comments = events.filter((event) => event.kind === 'comment' && event.commentId);
  const byComment: Record<number, TimelineEvent> = {};
  for (const event of comments) {
    if (event.commentId) byComment[event.commentId] = event;
  }
  const hidden = events.length > visibleEdges * 2 + 1 ? events.slice(visibleEdges, events.length - visibleEdges) : [];
  const shown = open || hidden.length === 0
    ? events
    : [...events.slice(0, visibleEdges), ...events.slice(events.length - visibleEdges)];

  async function reload() {
    const timeline = await listTimeline(item.id, locale);
    setEvents(timeline.events);
  }

  function quote(event: TimelineEvent) {
    if (!user) {
      onLogin();
      return;
    }
    const source = (event.body ?? '').trim();
    const excerpt = source.split('\n').slice(0, 8).map((line) => `> ${line}`).join('\n');
    const block = excerpt ? `${excerpt}\n\n` : `> @${event.actor}\n\n`;
    setBody((current) => current.trim() ? `${current.trim()}\n\n${block}` : block);
    setPreview(false);
    setError('');
  }

  async function publish() {
    const length = [...body.trim()].length;
    if (length < 1 || length > 8000) {
      setError(t('commentLength'));
      return;
    }
    setPending(true);
    setError('');
    try {
      await createComment(item.id, body.trim(), 0, locale);
      setBody('');
      setPreview(false);
      await reload();
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('commentFailed'));
    } finally {
      setPending(false);
    }
  }

  async function saveEdit(commentId: number) {
    const length = [...editBody.trim()].length;
    if (length < 1 || length > 8000) {
      setError(t('commentLength'));
      return;
    }
    setPending(true);
    setError('');
    try {
      await updateComment(item.id, commentId, editBody.trim(), locale);
      setEditing(null);
      await reload();
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('commentFailed'));
    } finally {
      setPending(false);
    }
  }

  async function remove(commentId: number) {
    setPending(true);
    setError('');
    try {
      await deleteComment(item.id, commentId, locale);
      setEditing(null);
      await reload();
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : t('commentFailed'));
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="timeline" aria-label={t('comments')}>
      {shown.map((event, index) => {
        const foldHere = !open && hidden.length > 0 && index === visibleEdges;
        return (
          <div key={event.id}>
            {foldHere && (
              <button type="button" className="timeline-fold" style={{ color: palette.gold }} onClick={() => setOpen(true)}>
                <ChevronDown size={14} />{t('showComments', { count: hidden.length })}
              </button>
            )}
            {event.kind === 'comment' ? (
              <article className="timeline-card" id={event.commentId ? `comment-${event.commentId}` : undefined} style={{ background: palette.surface }}>
                <header className="issue-comment-head" style={{ color: palette.muted }}>
                  <Avatar name={event.actor} src={event.actorAvatar} />
                  <strong style={{ color: palette.text }}>{event.actor}</strong>
                  <span style={{ color: palette.faint }}>{labels.dateTime(event.createdAt)}</span>
                  <span className="timeline-actions">
                    <button type="button" style={{ color: palette.gold }} onClick={() => quote(event)}><Quote size={13} />{t('reply')}</button>
                    {user && event.commentId && (user.role === 'admin' || user.username === event.actor) && editing !== event.commentId && (
                      <button type="button" style={{ color: palette.gold }} onClick={() => { setEditing(event.commentId ?? null); setEditBody(event.body ?? ''); }}>{t('editComment')}</button>
                    )}
                    {user && event.commentId && (user.role === 'admin' || user.username === event.actor) && (
                      <button type="button" style={{ color: palette.danger }} disabled={pending} onClick={() => void remove(event.commentId ?? 0)}>{t('deleteComment')}</button>
                    )}
                  </span>
                </header>
                {event.replyTo ? <ReplyChip event={byComment[event.replyTo]} paletteText={palette.text} missing={t('commentMissing')} /> : null}
                {editing === event.commentId ? (
                  <div className="timeline-editor">
                    <AttachmentEditor body={editBody} onBody={setEditBody} preview={false} rows={6} locale={locale} target={editorTarget(item.id)} control={control} onError={setError} />
                    <div className="issue-editor-actions">
                      <button type="button" className="form-switch" style={{ color: palette.muted }} onClick={() => setEditing(null)}>{t('close')}</button>
                      <ActionButton className="form-submit" disabled={pending} onClick={() => void saveEdit(event.commentId ?? 0)}>{t('saveComment')}</ActionButton>
                    </div>
                  </div>
                ) : (
                  <div className="detail-body"><MarkdownBody source={event.body ?? ''} /></div>
                )}

              </article>
            ) : (
              <p className="timeline-event" style={{ color: palette.muted }}>
                <Avatar name={event.actor} src={event.actorAvatar} />
                <strong style={{ color: palette.text }}>{event.actor}</strong>
                <span>{eventText(event, t)}</span>
                {event.status && <span className="status-pill" style={statusColors(statuses, event.status, palette.dark)}><i />{labels.status(event.status)}</span>}
                <time style={{ color: palette.faint }}>{labels.dateTime(event.createdAt)}</time>
              </p>
            )}
          </div>
        );
      })}
      {open && hidden.length > 0 && (
        <button type="button" className="timeline-fold" style={{ color: palette.gold }} onClick={() => setOpen(false)}>
          <ChevronDown size={14} />{t('hideComments')}
        </button>
      )}
      <article className="timeline-card" style={{ background: palette.surface }}>
        <header className="issue-comment-head" style={{ color: palette.muted }}>
          <MessageSquare size={15} />
          <strong style={{ color: palette.text }}>{canPost ? user?.username : t('commentLogin')}</strong>
        </header>
        {canPost ? (
          <div className="timeline-editor">
            <AttachmentEditor body={body} onBody={setBody} preview={preview} rows={6} locale={locale} target={editorTarget(item.id)} control={control} onError={setError} />
            <button type="button" className="form-switch own-edit" style={{ color: palette.gold }} onClick={() => setPreview((current) => !current)}>{preview ? t('writeMarkdown') : t('previewMarkdown')}</button>
            {error && <p className="form-error" role="alert" style={{ color: palette.danger }}>{error}</p>}
            <ActionButton className="form-submit" disabled={pending} onClick={() => void publish()}>{pending ? t('publishingComment') : t('publishComment')}</ActionButton>
          </div>
        ) : (
          <button type="button" className="timeline-login" style={{ color: palette.gold }} onClick={onLogin}>{t('login')}</button>
        )}
      </article>
    </section>
  );
}

function ReplyChip({ event, paletteText, missing }: { event?: TimelineEvent; paletteText: string; missing: string }) {
  if (!event) return <a className="timeline-reply" href={`#comment-missing`} style={{ color: paletteText }}>{missing}</a>;
  const line = (event.body ?? '').split('\n').find((item) => item.trim() && !item.trim().startsWith('>')) ?? event.actor;
  return <a className="timeline-reply" href={`#comment-${event.commentId}`} style={{ color: paletteText }}><Quote size={12} />{event.actor}: {line}</a>;
}

function eventText(event: TimelineEvent, t: (key: 'changedStatus' | 'editedFeedback' | 'editedComment' | 'deletedComment') => string) {
  if (event.kind === 'status') return t('changedStatus');
  if (event.kind === 'edited') return t('editedFeedback');
  if (event.kind === 'comment_edited') return t('editedComment');
  return t('deletedComment');
}

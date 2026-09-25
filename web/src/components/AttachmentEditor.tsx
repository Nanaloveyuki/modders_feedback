import { useRef, useState, type ClipboardEvent, type DragEvent } from 'react';
import { Paperclip } from 'lucide-react';
import { useI18n } from '../i18n/context';
import type { Locale } from '../i18n/messages';
import { uploadAttachment, type UploadTarget } from '../api/feedback';
import { MarkdownBody } from './MarkdownBody';

const maxBytes = 12 * 1024 * 1024;

type Props = {
  body: string;
  onBody: (value: string) => void;
  preview: boolean;
  rows: number;
  locale: Locale;
  target: UploadTarget;
  control: { color: string; background: string };
  onError: (message: string) => void;
};

export function AttachmentEditor({ body, onBody, preview, rows, locale, target, control, onError }: Props) {
  const { t } = useI18n();
  const input = useRef<HTMLInputElement>(null);
  const area = useRef<HTMLTextAreaElement>(null);
  const [uploading, setUploading] = useState(0);
  const [over, setOver] = useState(false);

  async function send(files: File[], caret: number) {
    const accepted = files.filter((file) => file.size > 0 && file.size <= maxBytes);
    if (accepted.length !== files.length) onError(t('uploadLimit'));
    if (accepted.length === 0) return;
    setUploading((count) => count + accepted.length);
    let cursor = caret;
    let next = body;
    try {
      for (const file of accepted) {
        const saved = await uploadAttachment(target, file, locale);
        const snippet = markdownFor(saved.name, saved.url, saved.contentType);
        next = `${next.slice(0, cursor)}${snippet}${next.slice(cursor)}`;
        cursor += snippet.length;
      }
      onBody(next);
      onError('');
    } catch (problem) {
      onError(problem instanceof Error ? problem.message : t('uploadFailed'));
    } finally {
      setUploading((count) => count - accepted.length);
    }
  }

  function paste(event: ClipboardEvent<HTMLTextAreaElement>) {
    const files = [...event.clipboardData.files];
    if (files.length === 0) return;
    event.preventDefault();
    void send(files, event.currentTarget.selectionStart);
  }

  function drop(event: DragEvent<HTMLDivElement>) {
    const files = [...event.dataTransfer.files];
    setOver(false);
    if (files.length === 0) return;
    event.preventDefault();
    void send(files, area.current?.selectionStart ?? body.length);
  }

  return (
    <div
      className={over ? 'attachment-editor over' : 'attachment-editor'}
      onDragEnter={(event) => { if (event.dataTransfer.types.includes('Files')) setOver(true); }}
      onDragOver={(event) => { if (event.dataTransfer.types.includes('Files')) { event.preventDefault(); setOver(true); } }}
      onDragLeave={() => setOver(false)}
      onDrop={drop}
    >
      {preview ? (
        <div className="markdown-preview" style={control}>
          {body.trim() ? <MarkdownBody source={body} /> : <span>{t('markdownEmpty')}</span>}
        </div>
      ) : (
        <textarea ref={area} name="body" rows={rows} value={body} onChange={(event) => onBody(event.target.value)} onPaste={paste} style={control} />
      )}
      <div className="attachment-bar">
        <button type="button" disabled={uploading > 0} onClick={() => input.current?.click()}>
          <Paperclip size={14} />{uploading > 0 ? t('uploading') : t('attachFile')}
        </button>
        <input
          ref={input}
          type="file"
          multiple
          accept="image/png,image/jpeg,image/gif,image/webp,application/pdf,text/plain,.log,.txt,.md,.csv"
          onChange={(event) => {
            const files = [...event.target.files ?? []];
            event.target.value = '';
            void send(files, area.current?.selectionStart ?? body.length);
          }}
        />
      </div>
    </div>
  );
}

export function editorTarget(feedbackId?: number): UploadTarget {
  return feedbackId ? { kind: 'feedback', id: feedbackId } : { kind: 'draft' };
}

function markdownFor(name: string, url: string, contentType: string) {
  const label = name.replace(/[[\]]/g, '');
  if (contentType.startsWith('image/')) return `\n![${label}](${url})\n`;
  return `\n[${label}](${url})\n`;
}

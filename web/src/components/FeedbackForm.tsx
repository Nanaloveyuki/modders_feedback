import { useState, type FormEvent } from 'react';
import { ArrowRight } from 'lucide-react';
import type { Category, FeedbackDraft } from '../types';
import { categories, categoryKeys } from '../lib/labels';

type Props = {
  onSubmit: (data: FeedbackDraft) => Promise<void>;
};

const placeholders: Record<Category, { title: string; body: string }> = {
  bug: {
    title: '例如：冬季事件触发后殖民者状态异常',
    body: '发生了什么？预期结果是什么？请按步骤说明复现过程。',
  },
  feature: {
    title: '描述你希望增加或调整的内容',
    body: '目前的使用场景是什么？你希望它如何工作？',
  },
  question: {
    title: '用一句话概括你的问题',
    body: '提供相关背景信息，方便我们准确回答。',
  },
};

export function FeedbackForm({ onSubmit }: Props) {
  const [category, setCategory] = useState<Category>('bug');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError('');
    const form = new FormData(event.currentTarget);
    const data: FeedbackDraft = {
      category,
      title: String(form.get('title') ?? ''),
      body: String(form.get('body') ?? ''),
      gameVersion: String(form.get('gameVersion') ?? ''),
      modVersion: String(form.get('modVersion') ?? ''),
      modList: String(form.get('modList') ?? ''),
      saveLink: String(form.get('saveLink') ?? ''),
    };
    try {
      await onSubmit(data);
    } catch (problem) {
      setError(problem instanceof Error ? problem.message : '提交失败');
    } finally {
      setPending(false);
    }
  }

  return (
    <form className="form-stack feedback-form" onSubmit={handleSubmit}>
      <fieldset className="category-picker">
        <legend>反馈类型</legend>
        {categoryKeys.map((key) => (
          <button type="button" key={key} className={category === key ? `picked picked-${key}` : ''} onClick={() => setCategory(key)}>
            {categories[key].icon}{categories[key].label}
          </button>
        ))}
      </fieldset>
      <label>标题<input name="title" required minLength={5} maxLength={120} placeholder={placeholders[category].title} /></label>
      <label>详细描述<textarea name="body" required minLength={10} maxLength={12000} rows={5} placeholder={placeholders[category].body} /></label>
      <div className="form-two">
        <label>游戏版本<input name="gameVersion" maxLength={40} defaultValue="RimWorld 1.6" /></label>
        <label>模组版本<input name="modVersion" maxLength={80} placeholder="例如：0.8.2" /></label>
      </div>
      {category === 'bug' && (
        <>
          <label>相关模组 / 加载顺序<textarea name="modList" maxLength={6000} rows={2} placeholder="粘贴可能相关的模组和加载顺序（选填）" /></label>
          <label>存档分享链接<input name="saveLink" type="url" maxLength={500} placeholder="https://…（选填，请勿填写私密链接）" /></label>
        </>
      )}
      {error && <p className="form-error" role="alert">{error}</p>}
      <button className="form-submit" disabled={pending}>{pending ? '正在发布…' : '发布反馈'}<ArrowRight size={16} /></button>
    </form>
  );
}

export type Category = 'bug' | 'feature' | 'question';
export type Status = 'open' | 'in_progress' | 'resolved' | 'closed';

export type Feedback = {
  id: number;
  category: Category;
  categoryNumber?: number;
  title: string;
  body: string;
  author: string;
  gameVersion: string;
  modVersion: string;
  modList: string;
  saveLink: string;
  status: Status;
  createdAt: string;
};

export type User = { username: string };

export type FeedbackDraft = {
  category: Category;
  title: string;
  body: string;
  gameVersion: string;
  modVersion: string;
  modList: string;
  saveLink: string;
};

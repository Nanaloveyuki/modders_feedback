export type Category = 'bug' | 'feature' | 'question';
export type Status = 'open' | 'in_progress' | 'resolved' | 'closed' | 'withdrawn';

export type Feedback = {
  id: number;
  modId?: number;
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

export type User = { username: string; role: 'admin' | 'member' };

export type FeedbackDraft = {
  category: Category;
  title: string;
  body: string;
  gameVersion: string;
  modVersion: string;
  modList: string;
  saveLink: string;
};

export type SiteSettings = {
  modVersion: string;
  gameVersion: string;
  icon: SiteIcon;
};

export type SiteIcon = 'squirrel' | 'rat' | 'bug' | 'spark' | 'shield' | 'paw';

export type FeedbackUpdate = {
  title: string;
  body: string;
  gameVersion: string;
  modVersion: string;
  modList: string;
  saveLink: string;
};

export type Mod = {
  id: number;
  slug: string;
  name: string;
  gameVersion: string;
  modVersion: string;
  icon: SiteIcon;
};

export type ModInput = {
  slug: string;
  name: string;
  gameVersion: string;
  modVersion: string;
  icon: SiteIcon;
};

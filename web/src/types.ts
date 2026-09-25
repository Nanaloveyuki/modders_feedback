export type Category = 'bug' | 'feature' | 'question';
export type Status = 'open' | 'in_progress' | 'resolved' | 'closed' | 'withdrawn';

export type Feedback = {
  id: number;
  publicId: string;
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
  attachments: Attachment[];
}

export type Attachment = {
  id: string;
  name: string;
  contentType: string;
  size: number;
  url: string;
};

export type User = {
  username: string;
  role: 'admin' | 'member';
  email: string;
  qq: string;
  avatarUrl: string;
};

export type ProfileUpdate = {
  username: string;
  email: string;
  qq: string;
};

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
  attachmentDir: string;
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
  steamUrl: string;
  githubUrl: string;
};

export type ModInput = {
  slug: string;
  name: string;
  gameVersion: string;
  modVersion: string;
  icon: SiteIcon;
  steamUrl: string;
  githubUrl: string;
};

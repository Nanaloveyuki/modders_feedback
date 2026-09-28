export type Category = 'bug' | 'feature' | 'question';
export type Status = string;

export type Feedback = {
  id: number;
  publicId: string;
  modId?: number;
  category: Category;
  categoryNumber?: number;
  title: string;
  body: string;
  author: string;
  authorAvatar: string;
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

export type StatusLabel = {
  key: string;
  labelZh: string;
  labelEn: string;
  light: string;
  dark: string;
  author: boolean;
  archived: boolean;
};

export type SiteSettings = {
  modVersion: string;
  gameVersion: string;
  icon: SiteIcon;
  attachmentDir: string;
  statuses: StatusLabel[];
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

export type Comment = {
  id: number;
  feedbackId: number;
  author: string;
  authorAvatar: string;
  body: string;
  replyTo?: number;
  createdAt: string;
  updatedAt: string;
};

export type TimelineKind = 'comment' | 'status' | 'edited' | 'comment_edited' | 'deleted';

export type TimelineEvent = {
  id: number;
  kind: TimelineKind;
  actor: string;
  actorAvatar: string;
  body?: string;
  replyTo?: number;
  status?: Status;
  commentId?: number;
  createdAt: string;
};

export type Timeline = {
  events: TimelineEvent[];
};

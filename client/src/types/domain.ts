export interface User {
  userId: string;
  username: string;
}

export interface Member {
  userId: string;
  username: string;
}

export interface Room {
  id: string;
  name: string;
  members: Member[];
}

export interface ReactionView {
  emoji: string;
  userIds: string[];
  count: number;
}

export interface ReplyToSummary {
  authorName: string;
  content: string;
}

export interface Message {
  id: string;
  roomId: string;
  authorId: string;
  authorName: string;
  content: string;
  clientTempId?: string;
  createdAt: string;
  editedAt?: string;
  pending?: boolean;
  reactions: ReactionView[];
  replyToId?: string;
  replyToSummary?: ReplyToSummary;
}

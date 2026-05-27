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
}

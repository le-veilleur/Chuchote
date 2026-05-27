import { create } from 'zustand';
import type { Message } from '../types/domain';

interface TypingUser {
  userId: string;
  username: string;
}

interface ChatState {
  messagesByRoom: Record<string, Message[]>;
  typingByRoom: Record<string, TypingUser[]>;
  addMessage: (roomId: string, message: Message) => void;
  confirmMessage: (roomId: string, clientTempId: string, confirmed: Message) => void;
  setHistory: (roomId: string, messages: Message[]) => void;
  setTyping: (roomId: string, user: TypingUser, isTyping: boolean) => void;
  editMessage: (roomId: string, messageId: string, content: string, editedAt: string) => void;
  deleteMessage: (roomId: string, messageId: string) => void;
}

export const useChatStore = create<ChatState>((set) => ({
  messagesByRoom: {},
  typingByRoom: {},

  addMessage: (roomId, message) =>
    set((s) => ({
      messagesByRoom: {
        ...s.messagesByRoom,
        [roomId]: [...(s.messagesByRoom[roomId] ?? []), message],
      },
    })),

  confirmMessage: (roomId, clientTempId, confirmed) =>
    set((s) => ({
      messagesByRoom: {
        ...s.messagesByRoom,
        [roomId]: (s.messagesByRoom[roomId] ?? []).map((m) =>
          m.clientTempId === clientTempId ? confirmed : m
        ),
      },
    })),

  setHistory: (roomId, messages) =>
    set((s) => ({
      messagesByRoom: { ...s.messagesByRoom, [roomId]: messages },
    })),

  setTyping: (roomId, user, isTyping) =>
    set((s) => {
      const current = s.typingByRoom[roomId] ?? [];
      const filtered = current.filter((u) => u.userId !== user.userId);
      return {
        typingByRoom: {
          ...s.typingByRoom,
          [roomId]: isTyping ? [...filtered, user] : filtered,
        },
      };
    }),

  editMessage: (roomId, messageId, content, editedAt) =>
    set((s) => ({
      messagesByRoom: {
        ...s.messagesByRoom,
        [roomId]: (s.messagesByRoom[roomId] ?? []).map((m) =>
          m.id === messageId ? { ...m, content, editedAt } : m
        ),
      },
    })),

  deleteMessage: (roomId, messageId) =>
    set((s) => ({
      messagesByRoom: {
        ...s.messagesByRoom,
        [roomId]: (s.messagesByRoom[roomId] ?? []).filter((m) => m.id !== messageId),
      },
    })),
}));

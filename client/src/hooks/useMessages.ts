import { useChatStore } from '../store/chat.store';
import { useAuthStore } from '../store/auth.store';
import { wsService } from '../services/websocket.service';
import type { Message } from '../types/domain';

const EMPTY_MESSAGES: Message[] = [];

export function useMessages(roomId: string) {
  const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? EMPTY_MESSAGES);
  const addMessage = useChatStore((s) => s.addMessage);
  const { userId, username } = useAuthStore();

  const send = (content: string) => {
    if (!content.trim() || !userId || !username) return;
    const clientTempId = crypto.randomUUID();

    const optimistic: Message = {
      id: clientTempId,
      roomId,
      authorId: userId,
      authorName: username,
      content,
      clientTempId,
      createdAt: new Date().toISOString(),
      pending: true,
    };
    addMessage(roomId, optimistic);

    wsService.send({
      type: 'message.send',
      requestId: crypto.randomUUID(),
      roomId,
      payload: { content, clientTempId },
    });
  };

  return { messages, send };
}

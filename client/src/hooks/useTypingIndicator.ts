import { useCallback, useRef } from 'react';
import { useChatStore } from '../store/chat.store';
import { wsService } from '../services/websocket.service';
import { useAuthStore } from '../store/auth.store';

const EMPTY_TYPING: { userId: string; username: string }[] = [];

export function useTypingIndicator(roomId: string) {
  const allTyping = useChatStore((s) => s.typingByRoom[roomId] ?? EMPTY_TYPING);
  const myId = useAuthStore((s) => s.userId);
  const typingUsers = allTyping.filter((u) => u.userId !== myId);
  const stopTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isTypingRef = useRef(false);

  const onKeystroke = useCallback(() => {
    if (!isTypingRef.current) {
      isTypingRef.current = true;
      wsService.send({ type: 'typing.start', requestId: null, roomId, payload: {} });
    }
    if (stopTimerRef.current) clearTimeout(stopTimerRef.current);
    stopTimerRef.current = setTimeout(() => {
      isTypingRef.current = false;
      wsService.send({ type: 'typing.stop', requestId: null, roomId, payload: {} });
    }, 2000);
  }, [roomId]);

  return { typingUsers, onKeystroke };
}

import { useEffect, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import { TypingIndicator } from './TypingIndicator';
import { useChatStore } from '../../store/chat.store';
import { useAuthStore } from '../../store/auth.store';
import type { Message } from '../../types/domain';

const EMPTY_MESSAGES: Message[] = [];
const EMPTY_TYPING: { userId: string; username: string }[] = [];

interface Props {
  roomId: string;
}

export function MessageList({ roomId }: Props) {
  const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? EMPTY_MESSAGES);
  const allTyping = useChatStore((s) => s.typingByRoom[roomId] ?? EMPTY_TYPING);
  const myId = useAuthStore((s) => s.userId);
  const typingUsers = allTyping.filter((u) => u.userId !== myId);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, typingUsers]);

  return (
    <div style={{
      flex: 1,
      overflowY: 'auto',
      display: 'flex',
      flexDirection: 'column',
      gap: 12,
      padding: '16px',
    }}>
      {messages.map((m) => (
        <MessageBubble key={m.clientTempId ?? m.id} message={m} />
      ))}
      <TypingIndicator usernames={typingUsers.map((u) => u.username)} />
      <div ref={bottomRef} />
    </div>
  );
}

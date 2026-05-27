import { useEffect, useRef } from 'react';
import { MessageBubble } from './MessageBubble';
import { TypingIndicator } from './TypingIndicator';
import { useChatStore } from '../../store/chat.store';
import { useAuthStore } from '../../store/auth.store';
import { useMessages } from '../../hooks/useMessages';
import type { Message } from '../../types/domain';

const EMPTY_MESSAGES: Message[] = [];
const EMPTY_TYPING: { userId: string; username: string }[] = [];

interface Props {
  roomId: string;
  onReply: (message: Message) => void;
}

export function MessageList({ roomId, onReply }: Props) {
  const messages = useChatStore((s) => s.messagesByRoom[roomId] ?? EMPTY_MESSAGES);
  const allTyping = useChatStore((s) => s.typingByRoom[roomId] ?? EMPTY_TYPING);
  const myId = useAuthStore((s) => s.userId);
  const typingUsers = allTyping.filter((u) => u.userId !== myId);
  const { edit, remove, toggleReaction } = useMessages(roomId);
  const bottomRef = useRef<HTMLDivElement>(null);
  const isFirstLoad = useRef(true);

  useEffect(() => {
    if (isFirstLoad.current) {
      bottomRef.current?.scrollIntoView({ behavior: 'instant' });
      isFirstLoad.current = false;
    } else {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, typingUsers]);

  // Reset first-load flag when room changes
  useEffect(() => {
    isFirstLoad.current = true;
  }, [roomId]);

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
        <MessageBubble
          key={m.clientTempId ?? m.id}
          message={m}
          onEdit={edit}
          onDelete={remove}
          onReply={onReply}
          onReaction={toggleReaction}
        />
      ))}
      <TypingIndicator usernames={typingUsers.map((u) => u.username)} />
      <div ref={bottomRef} />
    </div>
  );
}

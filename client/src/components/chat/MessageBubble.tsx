import { Avatar } from '../ui/Avatar';
import type { Message } from '../../types/domain';
import { useAuthStore } from '../../store/auth.store';

interface Props {
  message: Message;
}

export function MessageBubble({ message }: Props) {
  const myId = useAuthStore((s) => s.userId);
  const isMine = message.authorId === myId;

  return (
    <div style={{
      display: 'flex',
      flexDirection: isMine ? 'row-reverse' : 'row',
      gap: 8,
      alignItems: 'flex-end',
      opacity: message.pending ? 0.6 : 1,
    }}>
      {!isMine && <Avatar username={message.authorName} size={28} />}
      <div style={{
        maxWidth: '70%',
        background: isMine ? 'var(--color-accent)' : 'var(--color-bg-2)',
        color: isMine ? '#fff' : 'var(--color-text)',
        borderRadius: isMine ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
        padding: '8px 12px',
        fontSize: 14,
        lineHeight: 1.5,
        wordBreak: 'break-word',
      }}>
        {!isMine && (
          <div style={{ fontSize: 11, fontWeight: 600, marginBottom: 2, opacity: 0.7 }}>
            {message.authorName}
          </div>
        )}
        <div>{message.content}</div>
        <div style={{ fontSize: 10, marginTop: 4, opacity: 0.6, textAlign: 'right' }}>
          {new Date(message.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </div>
      </div>
    </div>
  );
}

import { useState } from 'react';
import { useRoomStore } from '../../store/room.store';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import type { Message } from '../../types/domain';

export function ChatWindow() {
  const activeRoomId = useRoomStore((s) => s.activeRoomId);
  const rooms = useRoomStore((s) => s.rooms);
  const onlineByRoom = useRoomStore((s) => s.onlineByRoom);
  const room = rooms.find((r) => r.id === activeRoomId);
  const [replyingTo, setReplyingTo] = useState<Message | null>(null);

  if (!activeRoomId || !room) {
    return (
      <div style={{
        flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: 'var(--color-text-muted)', fontSize: 14,
      }}>
        Sélectionne une room pour commencer à chatter
      </div>
    );
  }

  const onlineCount = onlineByRoom[activeRoomId] ?? 0;

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{
        padding: '12px 16px',
        borderBottom: '1px solid var(--color-border)',
        fontWeight: 600,
        fontSize: 16,
        display: 'flex',
        alignItems: 'center',
        gap: 8,
      }}>
        #{room.name}
        <span style={{ fontWeight: 400, fontSize: 12, color: 'var(--color-text-muted)', display: 'flex', alignItems: 'center', gap: 4 }}>
          <span style={{ width: 7, height: 7, borderRadius: '50%', background: onlineCount > 0 ? '#22c55e' : '#6b7280', display: 'inline-block' }} />
          {onlineCount} en ligne
        </span>
      </div>
      <MessageList roomId={activeRoomId} onReply={setReplyingTo} />
      <MessageInput roomId={activeRoomId} replyingTo={replyingTo} onCancelReply={() => setReplyingTo(null)} />
    </div>
  );
}

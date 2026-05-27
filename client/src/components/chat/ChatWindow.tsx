import { useRoomStore } from '../../store/room.store';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';

export function ChatWindow() {
  const activeRoomId = useRoomStore((s) => s.activeRoomId);
  const rooms = useRoomStore((s) => s.rooms);
  const room = rooms.find((r) => r.id === activeRoomId);

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

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      <div style={{
        padding: '12px 16px',
        borderBottom: '1px solid var(--color-border)',
        fontWeight: 600,
        fontSize: 16,
      }}>
        #{room.name}
        <span style={{ fontWeight: 400, fontSize: 12, marginLeft: 8, color: 'var(--color-text-muted)' }}>
          {room.members.length} membre{room.members.length > 1 ? 's' : ''}
        </span>
      </div>
      <MessageList roomId={activeRoomId} />
      <MessageInput roomId={activeRoomId} />
    </div>
  );
}

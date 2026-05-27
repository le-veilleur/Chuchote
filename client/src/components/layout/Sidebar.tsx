import { useState } from 'react';
import { useRooms } from '../../hooks/useRooms';
import { useAuth } from '../../hooks/useAuth';

export function Sidebar() {
  const { rooms, activeRoomId, joinRoom, createRoom } = useRooms();
  const { username, logout } = useAuth();
  const [newRoomName, setNewRoomName] = useState('');

  const handleCreate = async () => {
    const name = newRoomName.trim();
    if (!name) return;
    const room = await createRoom(name);
    setNewRoomName('');
    joinRoom(room.id);
  };

  return (
    <div style={{
      width: 240,
      flexShrink: 0,
      background: 'var(--color-bg-2)',
      borderRight: '1px solid var(--color-border)',
      display: 'flex',
      flexDirection: 'column',
      padding: '16px 0',
    }}>
      <div style={{ padding: '0 16px 12px', fontWeight: 700, fontSize: 18 }}>Chuchote</div>

      <div style={{ padding: '0 16px 8px', fontSize: 11, fontWeight: 600, color: 'var(--color-text-muted)', textTransform: 'uppercase', letterSpacing: 1 }}>
        Rooms
      </div>

      <div style={{ flex: 1, overflowY: 'auto' }}>
        {rooms.map((room) => (
          <button
            key={room.id}
            onClick={() => joinRoom(room.id)}
            style={{
              width: '100%',
              textAlign: 'left',
              padding: '8px 16px',
              border: 'none',
              background: activeRoomId === room.id ? 'var(--color-accent)' : 'transparent',
              color: activeRoomId === room.id ? '#fff' : 'var(--color-text)',
              cursor: 'pointer',
              fontSize: 14,
              borderRadius: 0,
            }}
          >
            #{room.name}
          </button>
        ))}
      </div>

      <div style={{ padding: '12px 16px', display: 'flex', gap: 6 }}>
        <input
          value={newRoomName}
          onChange={(e) => setNewRoomName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
          placeholder="Nouvelle room…"
          style={{
            flex: 1, padding: '6px 8px', fontSize: 13, borderRadius: 6,
            border: '1px solid var(--color-border)', background: 'var(--color-bg)',
            color: 'var(--color-text)',
          }}
        />
        <button
          onClick={handleCreate}
          style={{
            padding: '6px 10px', borderRadius: 6, border: 'none',
            background: 'var(--color-accent)', color: '#fff', cursor: 'pointer', fontSize: 13,
          }}
        >+</button>
      </div>

      <div style={{
        padding: '12px 16px',
        borderTop: '1px solid var(--color-border)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        fontSize: 13,
      }}>
        <span style={{ fontWeight: 600 }}>{username}</span>
        <button
          onClick={logout}
          style={{ fontSize: 12, border: 'none', background: 'none', cursor: 'pointer', color: 'var(--color-text-muted)' }}
        >
          Déconnexion
        </button>
      </div>
    </div>
  );
}

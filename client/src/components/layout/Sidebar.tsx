import { useState } from 'react';
import { Plus, LogOut } from 'lucide-react';
import { useRooms } from '../../hooks/useRooms';
import { useAuth } from '../../hooks/useAuth';

export function Sidebar() {
  const { rooms, activeRoomId, joinRoom, createRoom } = useRooms();
  const { username, logout } = useAuth();
  const [newRoomName, setNewRoomName] = useState('');
  const [addHovered, setAddHovered] = useState(false);
  const [logoutHovered, setLogoutHovered] = useState(false);

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
          <RoomButton
            key={room.id}
            name={room.name}
            active={activeRoomId === room.id}
            onClick={() => joinRoom(room.id)}
          />
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
          onMouseEnter={() => setAddHovered(true)}
          onMouseLeave={() => setAddHovered(false)}
          title="Créer la room"
          style={{
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            width: 32, height: 32, borderRadius: 6, border: 'none',
            background: addHovered ? 'var(--color-accent-hover, #7c3aed)' : 'var(--color-accent)',
            color: '#fff', cursor: 'pointer',
            transition: 'background 0.15s',
          }}
        >
          <Plus size={16} strokeWidth={2.5} />
        </button>
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
          onMouseEnter={() => setLogoutHovered(true)}
          onMouseLeave={() => setLogoutHovered(false)}
          title="Déconnexion"
          style={{
            display: 'flex', alignItems: 'center', gap: 4,
            fontSize: 12, border: 'none', borderRadius: 4,
            background: logoutHovered ? 'rgba(255,255,255,0.08)' : 'none',
            cursor: 'pointer',
            color: logoutHovered ? 'var(--color-text)' : 'var(--color-text-muted)',
            padding: '4px 6px',
            transition: 'background 0.15s, color 0.15s',
          }}
        >
          <LogOut size={13} />
          Déconnexion
        </button>
      </div>
    </div>
  );
}

function RoomButton({ name, active, onClick }: { name: string; active: boolean; onClick: () => void }) {
  const [hovered, setHovered] = useState(false);
  return (
    <button
      onClick={onClick}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        width: '100%',
        textAlign: 'left',
        padding: '8px 16px',
        border: 'none',
        background: active
          ? 'var(--color-accent)'
          : hovered ? 'rgba(255,255,255,0.06)' : 'transparent',
        color: active ? '#fff' : 'var(--color-text)',
        cursor: 'pointer',
        fontSize: 14,
        borderRadius: 0,
        transition: 'background 0.1s',
      }}
    >
      #{name}
    </button>
  );
}

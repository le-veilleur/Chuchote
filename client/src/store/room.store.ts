import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Room } from '../types/domain';

interface RoomState {
  rooms: Room[];
  activeRoomId: string | null;
  setRooms: (rooms: Room[]) => void;
  addRoom: (room: Room) => void;
  updateRoom: (room: Room) => void;
  setActiveRoom: (roomId: string | null) => void;
}

export const useRoomStore = create<RoomState>()(
  persist(
    (set) => ({
      rooms: [],
      activeRoomId: null,
      setRooms: (newRooms) => set((s) => ({
        // Merge: keep member data from WS (updateRoom) if already populated,
        // because HTTP response can arrive after room.joined and would wipe it.
        rooms: newRooms.map((r) => {
          const existing = s.rooms.find((e) => e.id === r.id);
          return existing && existing.members.length > 0 ? { ...r, members: existing.members } : r;
        }),
      })),
      addRoom: (room) => set((s) => ({ rooms: [...s.rooms, room] })),
      updateRoom: (room) =>
        set((s) => ({ rooms: s.rooms.map((r) => (r.id === room.id ? room : r)) })),
      setActiveRoom: (roomId) => set({ activeRoomId: roomId }),
    }),
    {
      name: 'chuchote-room',
      partialize: (s) => ({ activeRoomId: s.activeRoomId }), // persiste uniquement la room active, pas la liste
    }
  )
);

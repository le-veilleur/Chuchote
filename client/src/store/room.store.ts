import { create } from 'zustand';
import type { Room } from '../types/domain';

interface RoomState {
  rooms: Room[];
  activeRoomId: string | null;
  setRooms: (rooms: Room[]) => void;
  addRoom: (room: Room) => void;
  updateRoom: (room: Room) => void;
  setActiveRoom: (roomId: string | null) => void;
}

export const useRoomStore = create<RoomState>((set) => ({
  rooms: [],
  activeRoomId: null,
  setRooms: (rooms) => set({ rooms }),
  addRoom: (room) => set((s) => ({ rooms: [...s.rooms, room] })),
  updateRoom: (room) =>
    set((s) => ({ rooms: s.rooms.map((r) => (r.id === room.id ? room : r)) })),
  setActiveRoom: (roomId) => set({ activeRoomId: roomId }),
}));

import { useEffect } from 'react';
import { useRoomStore } from '../store/room.store';
import { wsService } from '../services/websocket.service';
import { http } from '../services/http.service';
import type { Room } from '../types/domain';

export function useRooms() {
  const { rooms, activeRoomId, setRooms, setActiveRoom } = useRoomStore();

  useEffect(() => {
    http.get<Room[]>('/rooms').then(setRooms).catch(() => {});
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const joinRoom = (roomId: string) => {
    wsService.send({ type: 'room.join', requestId: crypto.randomUUID(), roomId, payload: {} });
    setActiveRoom(roomId);
  };

  const leaveRoom = (roomId: string) => {
    wsService.send({ type: 'room.leave', requestId: crypto.randomUUID(), roomId, payload: {} });
    setActiveRoom(null);
  };

  const createRoom = async (name: string) => {
    const room = await http.post<Room>('/rooms', { name });
    useRoomStore.getState().addRoom(room);
    return room;
  };

  return { rooms, activeRoomId, joinRoom, leaveRoom, createRoom };
}

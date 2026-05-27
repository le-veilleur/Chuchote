import type { Room } from './domain';

export interface TokenResponse {
  token: string;  // matches Go dto.TokenView json:"token"
  userId: string;
  username: string;
}

export interface RoomsResponse {
  rooms: Room[];
}

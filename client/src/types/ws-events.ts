import type { Message, Room, Member } from './domain';

export interface WSFrame<T extends string, P> {
  type: T;
  requestId: string | null;
  roomId: string | null;
  payload: P;
}

// Client → Server
export type AuthConnectEvent = WSFrame<'auth.connect', { token: string }>;
export type RoomJoinEvent = WSFrame<'room.join', Record<string, never>>;
export type RoomLeaveEvent = WSFrame<'room.leave', Record<string, never>>;
export type MessageSendEvent = WSFrame<'message.send', { content: string; clientTempId: string }>;
export type TypingStartEvent = WSFrame<'typing.start', Record<string, never>>;
export type TypingStopEvent = WSFrame<'typing.stop', Record<string, never>>;

export type MessageEditEvent = WSFrame<'message.edit', { messageId: string; content: string }>;
export type MessageDeleteEvent = WSFrame<'message.delete', { messageId: string }>;

export type OutboundWSEvent =
  | AuthConnectEvent
  | RoomJoinEvent
  | RoomLeaveEvent
  | MessageSendEvent
  | MessageEditEvent
  | MessageDeleteEvent
  | TypingStartEvent
  | TypingStopEvent;

// Server → Client
export type AuthConnectedEvent = WSFrame<'auth.connected', { userId: string; username: string }>;
export type AuthErrorEvent = WSFrame<'auth.error', { code: string; message: string }>;
export type RoomJoinedEvent = WSFrame<'room.joined', { room: Room; members: Member[]; history: Message[] }>;
export type MessageNewEvent = WSFrame<'message.new', Message>;
export type MessageAckEvent = WSFrame<'message.ack', { messageId: string; clientTempId: string; createdAt: string }>;
export type TypingIndicatorEvent = WSFrame<'typing.indicator', { userId: string; username: string; isTyping: boolean }>;
export type UserPresenceEvent = WSFrame<'user.presence', { userId: string; username: string; status: 'online' | 'offline' }>;
export type ErrorEvent = WSFrame<'error', { code: string; message: string }>;

export type MessageEditedEvent = WSFrame<'message.edited', { messageId: string; content: string; editedAt: string }>;
export type MessageDeletedEvent = WSFrame<'message.deleted', { messageId: string }>;

export type InboundWSEvent =
  | AuthConnectedEvent
  | AuthErrorEvent
  | RoomJoinedEvent
  | MessageNewEvent
  | MessageAckEvent
  | MessageEditedEvent
  | MessageDeletedEvent
  | TypingIndicatorEvent
  | UserPresenceEvent
  | ErrorEvent;

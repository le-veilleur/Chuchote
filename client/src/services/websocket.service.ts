import type { InboundWSEvent, OutboundWSEvent } from '../types/ws-events';
import { useChatStore } from '../store/chat.store';
import { useRoomStore } from '../store/room.store';
import { useAuthStore } from '../store/auth.store';

type EventHandler<T extends InboundWSEvent> = (event: T) => void;
type AnyHandler = EventHandler<InboundWSEvent>;

class WebSocketService {
  private ws: WebSocket | null = null;
  private handlers = new Map<string, Set<AnyHandler>>();
  private queue: OutboundWSEvent[] = [];
  private currentUrl = '';

  connect(url: string): void {
    this.currentUrl = url;
    if (this.ws?.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      const { token } = useAuthStore.getState();
      if (token) {
        this.ws!.send(JSON.stringify({
          type: 'auth.connect',
          requestId: crypto.randomUUID(),
          roomId: null,
          payload: { token },
        }));
      }
      const pending = this.queue.splice(0);
      for (const event of pending) {
        this.ws!.send(JSON.stringify(event));
      }
    };

    this.ws.onmessage = (e) => {
      try {
        const frame = JSON.parse(e.data) as InboundWSEvent;
        this.dispatch(frame);
        this.handleStoreUpdates(frame);
      } catch {
        // ignore malformed frames
      }
    };

    this.ws.onclose = () => {
      setTimeout(() => this.connect(this.currentUrl), 2000);
    };
  }

  disconnect(): void {
    this.currentUrl = '';
    this.queue = [];
    this.ws?.close();
    this.ws = null;
  }

  send(event: OutboundWSEvent): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(event));
    } else {
      this.queue.push(event);
    }
  }

  on<T extends InboundWSEvent>(type: T['type'], handler: EventHandler<T>): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler as AnyHandler);
    return () => this.handlers.get(type)?.delete(handler as AnyHandler);
  }

  private dispatch(event: InboundWSEvent): void {
    this.handlers.get(event.type)?.forEach((h) => h(event));
  }

  private handleStoreUpdates(event: InboundWSEvent): void {
    const chat = useChatStore.getState();
    const rooms = useRoomStore.getState();

    switch (event.type) {
      case 'auth.connected': {
        const activeRoomId = useRoomStore.getState().activeRoomId;
        if (activeRoomId) {
          this.send({ type: 'room.join', requestId: crypto.randomUUID(), roomId: activeRoomId, payload: {} });
        }
        break;
      }

      case 'room.joined':
        chat.setHistory(event.roomId!, event.payload.history);
        rooms.updateRoom(event.payload.room);
        if (event.roomId) {
          rooms.setOnlineCount(event.roomId, event.payload.onlineCount);
        }
        break;

      case 'room.online_count':
        if (event.roomId) {
          rooms.setOnlineCount(event.roomId, event.payload.count);
        }
        break;

      case 'message.new': {
        const msg = event.payload;
        chat.addMessage(event.roomId!, msg);
        break;
      }

      case 'message.ack': {
        const { clientTempId, messageId, createdAt, replyToId, replyToSummary } = event.payload;
        if (event.roomId) {
          const existing = chat.messagesByRoom[event.roomId]?.find(
            (m) => m.clientTempId === clientTempId
          );
          if (existing) {
            chat.confirmMessage(event.roomId, clientTempId, {
              ...existing,
              id: messageId,
              createdAt,
              pending: false,
              replyToId,
              replyToSummary,
            });
          }
        }
        break;
      }

      case 'message.edited':
        if (event.roomId) {
          chat.editMessage(event.roomId, event.payload.messageId, event.payload.content, event.payload.editedAt);
        }
        break;

      case 'message.deleted':
        if (event.roomId) {
          chat.deleteMessage(event.roomId, event.payload.messageId);
        }
        break;

      case 'typing.indicator':
        if (event.roomId) {
          chat.setTyping(event.roomId, { userId: event.payload.userId, username: event.payload.username }, event.payload.isTyping);
        }
        break;

      case 'reaction.updated':
        if (event.roomId) {
          chat.setReactions(event.roomId, event.payload.messageId, event.payload.reactions);
        }
        break;
    }
  }
}

export const wsService = new WebSocketService();

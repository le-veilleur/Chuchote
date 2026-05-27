package hub

import (
	"sync"

	"github.com/maxime/chuchote/domain/model"
)

type client struct {
	conn model.Connection
	send chan<- []byte
}

type Hub struct {
	mu            sync.RWMutex
	clients       map[model.ConnID]client
	roomMembers   map[model.RoomID]map[model.ConnID]struct{}
	userConns     map[model.UserID]map[model.ConnID]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[model.ConnID]client),
		roomMembers: make(map[model.RoomID]map[model.ConnID]struct{}),
		userConns:   make(map[model.UserID]map[model.ConnID]struct{}),
	}
}

func (h *Hub) Register(conn model.Connection, send chan<- []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[conn.ID] = client{conn: conn, send: send}

	if h.userConns[conn.UserID] == nil {
		h.userConns[conn.UserID] = make(map[model.ConnID]struct{})
	}
	h.userConns[conn.UserID][conn.ID] = struct{}{}
}

func (h *Hub) Unregister(conn model.Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, conn.ID)

	if conns := h.userConns[conn.UserID]; conns != nil {
		delete(conns, conn.ID)
		if len(conns) == 0 {
			delete(h.userConns, conn.UserID)
		}
	}

	for roomID, members := range h.roomMembers {
		delete(members, conn.ID)
		if len(members) == 0 {
			delete(h.roomMembers, roomID)
		}
	}
}

func (h *Hub) SubscribeToRoom(conn model.Connection, roomID model.RoomID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.roomMembers[roomID] == nil {
		h.roomMembers[roomID] = make(map[model.ConnID]struct{})
	}
	h.roomMembers[roomID][conn.ID] = struct{}{}
}

func (h *Hub) UnsubscribeFromRoom(conn model.Connection, roomID model.RoomID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members := h.roomMembers[roomID]; members != nil {
		delete(members, conn.ID)
	}
}

func (h *Hub) BroadcastToRoom(roomID model.RoomID, payload []byte) {
	h.BroadcastToRoomExcept(roomID, "", payload)
}

func (h *Hub) BroadcastToRoomExcept(roomID model.RoomID, excludeUserID model.UserID, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for connID := range h.roomMembers[roomID] {
		if c, ok := h.clients[connID]; ok {
			if excludeUserID != "" && c.conn.UserID == excludeUserID {
				continue
			}
			select {
			case c.send <- payload:
			default:
			}
		}
	}
}

func (h *Hub) SendToUser(userID model.UserID, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for connID := range h.userConns[userID] {
		if c, ok := h.clients[connID]; ok {
			select {
			case c.send <- payload:
			default:
			}
		}
	}
}

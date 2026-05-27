package ws

import (
	"net/http"

	"github.com/maxime/chuchote/port/inbound"
	"github.com/maxime/chuchote/port/outbound"
	"github.com/coder/websocket"
)

type Handler struct {
	hub      outbound.BroadcastHub
	messages inbound.MessageUseCase
	rooms    inbound.RoomUseCase
	auth     inbound.AuthUseCase
}

func NewHandler(hub outbound.BroadcastHub, messages inbound.MessageUseCase, rooms inbound.RoomUseCase, auth inbound.AuthUseCase) *Handler {
	return &Handler{hub: hub, messages: messages, rooms: rooms, auth: auth}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // handled by CORS middleware in production
	})
	if err != nil {
		return
	}

	c := newClient(conn, h.hub, h.messages, h.rooms, h.auth)
	c.run(r.Context())
}

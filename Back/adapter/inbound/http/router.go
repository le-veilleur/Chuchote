package http

import (
	"net/http"

	"github.com/maxime/chuchote/adapter/inbound/http/middleware"
	"github.com/maxime/chuchote/port/inbound"
	ws "github.com/maxime/chuchote/adapter/inbound/ws"
	"github.com/maxime/chuchote/port/outbound"
)

func NewRouter(
	authHandler *AuthHandler,
	roomHandler *RoomHandler,
	authSvc inbound.AuthUseCase,
	hub outbound.BroadcastHub,
	messages inbound.MessageUseCase,
	rooms inbound.RoomUseCase,
) http.Handler {
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Room routes (protected)
	authMiddleware := middleware.Auth(authSvc)
	mux.Handle("POST /rooms", authMiddleware(http.HandlerFunc(roomHandler.Create)))
	mux.Handle("GET /rooms", authMiddleware(http.HandlerFunc(roomHandler.List)))
	mux.Handle("GET /rooms/{id}", authMiddleware(http.HandlerFunc(roomHandler.Get)))

	// WebSocket endpoint
	wsHandler := ws.NewHandler(hub, messages, rooms, authSvc)
	mux.Handle("/ws", wsHandler)

	return middleware.CORS(mux)
}

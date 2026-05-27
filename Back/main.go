package main

import (
	"log/slog"
	"os"

	httpadapter "github.com/maxime/chuchote/adapter/inbound/http"
	"github.com/maxime/chuchote/adapter/outbound/hub"
	"github.com/maxime/chuchote/adapter/outbound/memory"
	"github.com/maxime/chuchote/application/service"
	"github.com/maxime/chuchote/infrastructure/config"
	"github.com/maxime/chuchote/infrastructure/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.Load()

	// Outbound adapters
	userRepo := memory.NewUserRepo()
	roomRepo := memory.NewRoomRepo()
	messageRepo := memory.NewMessageRepo()
	broadcastHub := hub.NewHub()

	// Application services
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	roomSvc := service.NewRoomService(roomRepo, userRepo)
	messageSvc := service.NewMessageService(messageRepo, broadcastHub, userRepo)

	// Inbound adapters
	authHandler := httpadapter.NewAuthHandler(authSvc)
	roomHandler := httpadapter.NewRoomHandler(roomSvc)
	router := httpadapter.NewRouter(authHandler, roomHandler, authSvc, broadcastHub, messageSvc, roomSvc)

	if err := server.Run(cfg.Port, router); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

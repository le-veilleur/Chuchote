package http

import (
	"encoding/json"
	"net/http"

	"github.com/maxime/chuchote/adapter/inbound/http/middleware"
	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/domain/model"
	"github.com/maxime/chuchote/port/inbound"
)

type RoomHandler struct {
	rooms inbound.RoomUseCase
}

func NewRoomHandler(rooms inbound.RoomUseCase) *RoomHandler {
	return &RoomHandler{rooms: rooms}
}

func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var cmd dto.CreateRoomCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	cmd.CreatorID = claims.UserID
	view, err := h.rooms.CreateRoom(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, view)
}

func (h *RoomHandler) List(w http.ResponseWriter, r *http.Request) {
	views, err := h.rooms.ListRooms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, views)
}

func (h *RoomHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	view, err := h.rooms.GetRoom(r.Context(), model.RoomID(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

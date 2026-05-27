package http

import (
	"encoding/json"
	"net/http"

	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/port/inbound"
)

type AuthHandler struct {
	auth inbound.AuthUseCase
}

func NewAuthHandler(auth inbound.AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var cmd dto.RegisterCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	view, err := h.auth.Register(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, view)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var cmd dto.LoginCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	view, err := h.auth.Login(r.Context(), cmd)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

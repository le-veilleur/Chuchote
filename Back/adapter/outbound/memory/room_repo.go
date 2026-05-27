package memory

import (
	"context"
	"sync"

	domainerrors "github.com/maxime/chuchote/domain/errors"
	"github.com/maxime/chuchote/domain/model"
)

type RoomRepo struct {
	mu    sync.RWMutex
	rooms map[model.RoomID]model.Room
}

func NewRoomRepo() *RoomRepo {
	return &RoomRepo{rooms: make(map[model.RoomID]model.Room)}
}

func (r *RoomRepo) Save(_ context.Context, room model.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[room.ID] = room
	return nil
}

func (r *RoomRepo) FindByID(_ context.Context, id model.RoomID) (model.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	room, ok := r.rooms[id]
	if !ok {
		return model.Room{}, domainerrors.ErrRoomNotFound
	}
	return room, nil
}

func (r *RoomRepo) FindAll(_ context.Context) ([]model.Room, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		result = append(result, room)
	}
	return result, nil
}

func (r *RoomRepo) Update(_ context.Context, room model.Room) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.rooms[room.ID]; !ok {
		return domainerrors.ErrRoomNotFound
	}
	r.rooms[room.ID] = room
	return nil
}

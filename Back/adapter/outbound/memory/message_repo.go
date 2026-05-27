package memory

import (
	"context"
	"sort"
	"sync"

	domainerrors "github.com/maxime/chuchote/domain/errors"
	"github.com/maxime/chuchote/domain/model"
)

type MessageRepo struct {
	mu       sync.RWMutex
	messages []model.Message
}

func NewMessageRepo() *MessageRepo {
	return &MessageRepo{}
}

func (r *MessageRepo) Save(_ context.Context, msg model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, msg)
	return nil
}

func (r *MessageRepo) FindByID(_ context.Context, id model.MessageID) (model.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.messages {
		if m.ID == id {
			return m, nil
		}
	}
	return model.Message{}, domainerrors.ErrMessageNotFound
}

func (r *MessageRepo) FindByRoomID(_ context.Context, roomID model.RoomID, limit int) ([]model.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []model.Message
	for _, m := range r.messages {
		if m.RoomID == roomID {
			result = append(result, m)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	if limit > 0 && len(result) > limit {
		result = result[len(result)-limit:]
	}
	return result, nil
}

func (r *MessageRepo) Update(_ context.Context, msg model.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, m := range r.messages {
		if m.ID == msg.ID {
			r.messages[i] = msg
			return nil
		}
	}
	return domainerrors.ErrMessageNotFound
}

func (r *MessageRepo) Delete(_ context.Context, id model.MessageID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, m := range r.messages {
		if m.ID == id {
			r.messages = append(r.messages[:i], r.messages[i+1:]...)
			return nil
		}
	}
	return domainerrors.ErrMessageNotFound
}

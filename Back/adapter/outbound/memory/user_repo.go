package memory

import (
	"context"
	"sync"

	domainerrors "github.com/maxime/chuchote/domain/errors"
	"github.com/maxime/chuchote/domain/model"
)

type UserRepo struct {
	mu    sync.RWMutex
	byID  map[model.UserID]model.User
	byName map[string]model.UserID
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byID:   make(map[model.UserID]model.User),
		byName: make(map[string]model.UserID),
	}
}

func (r *UserRepo) Save(_ context.Context, user model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[user.ID] = user
	r.byName[user.Username] = user.ID
	return nil
}

func (r *UserRepo) FindByID(_ context.Context, id model.UserID) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byID[id]
	if !ok {
		return model.User{}, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (r *UserRepo) FindByUsername(_ context.Context, username string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byName[username]
	if !ok {
		return model.User{}, domainerrors.ErrUserNotFound
	}
	return r.byID[id], nil
}

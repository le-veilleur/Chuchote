package memory

import (
	"context"
	"sync"

	"github.com/maxime/chuchote/domain/model"
)

type ReactionRepo struct {
	mu        sync.RWMutex
	reactions []model.Reaction
}

func NewReactionRepo() *ReactionRepo {
	return &ReactionRepo{}
}

func (r *ReactionRepo) Toggle(_ context.Context, msgID model.MessageID, userID model.UserID, emoji string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rx := range r.reactions {
		if rx.MessageID == msgID && rx.UserID == userID && rx.Emoji == emoji {
			r.reactions = append(r.reactions[:i], r.reactions[i+1:]...)
			return false, nil
		}
	}
	r.reactions = append(r.reactions, model.Reaction{MessageID: msgID, UserID: userID, Emoji: emoji})
	return true, nil
}

func (r *ReactionRepo) FindByMessageIDs(_ context.Context, msgIDs []model.MessageID) (map[model.MessageID][]model.Reaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	idSet := make(map[model.MessageID]struct{}, len(msgIDs))
	for _, id := range msgIDs {
		idSet[id] = struct{}{}
	}
	result := make(map[model.MessageID][]model.Reaction)
	for _, rx := range r.reactions {
		if _, ok := idSet[rx.MessageID]; ok {
			result[rx.MessageID] = append(result[rx.MessageID], rx)
		}
	}
	return result, nil
}

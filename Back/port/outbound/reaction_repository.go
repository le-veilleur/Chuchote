package outbound

import (
	"context"

	"github.com/maxime/chuchote/domain/model"
)

type ReactionRepository interface {
	Toggle(ctx context.Context, msgID model.MessageID, userID model.UserID, emoji string) (added bool, err error)
	FindByMessageIDs(ctx context.Context, msgIDs []model.MessageID) (map[model.MessageID][]model.Reaction, error)
}

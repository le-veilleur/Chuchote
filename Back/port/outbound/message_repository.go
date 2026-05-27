package outbound

import (
	"context"

	"github.com/maxime/chuchote/domain/model"
)

type MessageRepository interface {
	Save(ctx context.Context, msg model.Message) error
	FindByID(ctx context.Context, id model.MessageID) (model.Message, error)
	FindByRoomID(ctx context.Context, roomID model.RoomID, limit int) ([]model.Message, error)
	Update(ctx context.Context, msg model.Message) error
	Delete(ctx context.Context, id model.MessageID) error
}

package outbound

import (
	"context"

	"github.com/maxime/chuchote/domain/model"
)

type MessageRepository interface {
	Save(ctx context.Context, msg model.Message) error
	FindByRoomID(ctx context.Context, roomID model.RoomID, limit int) ([]model.Message, error)
}

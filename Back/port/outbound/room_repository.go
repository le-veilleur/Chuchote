package outbound

import (
	"context"

	"github.com/maxime/chuchote/domain/model"
)

type RoomRepository interface {
	Save(ctx context.Context, room model.Room) error
	FindByID(ctx context.Context, id model.RoomID) (model.Room, error)
	FindAll(ctx context.Context) ([]model.Room, error)
	Update(ctx context.Context, room model.Room) error
}

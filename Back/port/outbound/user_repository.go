package outbound

import (
	"context"

	"github.com/maxime/chuchote/domain/model"
)

type UserRepository interface {
	Save(ctx context.Context, user model.User) error
	FindByID(ctx context.Context, id model.UserID) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
}

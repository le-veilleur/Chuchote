package inbound

import (
	"context"

	"github.com/maxime/chuchote/application/dto"
)

type AuthUseCase interface {
	Register(ctx context.Context, cmd dto.RegisterCommand) (dto.TokenView, error)
	Login(ctx context.Context, cmd dto.LoginCommand) (dto.TokenView, error)
	ValidateToken(ctx context.Context, token string) (dto.UserClaims, error)
}

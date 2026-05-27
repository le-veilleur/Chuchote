package inbound

import (
	"context"

	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/domain/model"
)

type RoomUseCase interface {
	CreateRoom(ctx context.Context, cmd dto.CreateRoomCommand) (dto.RoomView, error)
	JoinRoom(ctx context.Context, cmd dto.JoinRoomCommand) (dto.RoomView, error)
	LeaveRoom(ctx context.Context, userID model.UserID, roomID model.RoomID) error
	ListRooms(ctx context.Context) ([]dto.RoomView, error)
	GetRoom(ctx context.Context, roomID model.RoomID) (dto.RoomView, error)
}

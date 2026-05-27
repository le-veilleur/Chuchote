package inbound

import (
	"context"

	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/domain/model"
)

type MessageUseCase interface {
	SendMessage(ctx context.Context, cmd dto.SendMessageCommand) (dto.MessageView, error)
	GetRoomHistory(ctx context.Context, roomID model.RoomID, limit int) ([]dto.MessageView, error)
}

package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/domain/model"
	"github.com/maxime/chuchote/port/outbound"
)

type MessageService struct {
	messages outbound.MessageRepository
	hub      outbound.BroadcastHub
	users    outbound.UserRepository
}

func NewMessageService(messages outbound.MessageRepository, hub outbound.BroadcastHub, users outbound.UserRepository) *MessageService {
	return &MessageService{messages: messages, hub: hub, users: users}
}

func (s *MessageService) SendMessage(ctx context.Context, cmd dto.SendMessageCommand) (dto.MessageView, error) {
	msg, err := model.NewMessage(
		model.MessageID(uuid.NewString()),
		cmd.RoomID,
		cmd.AuthorID,
		cmd.Content,
		cmd.ClientTempID,
	)
	if err != nil {
		return dto.MessageView{}, err
	}

	if err := s.messages.Save(ctx, msg); err != nil {
		return dto.MessageView{}, err
	}

	view := dto.MessageView{
		ID:           msg.ID,
		RoomID:       msg.RoomID,
		AuthorID:     msg.AuthorID,
		AuthorName:   cmd.AuthorName,
		Content:      msg.Content,
		ClientTempID: msg.ClientTempID,
		CreatedAt:    msg.CreatedAt,
	}

	s.broadcastNewMessage(view)

	return view, nil
}

func (s *MessageService) GetRoomHistory(ctx context.Context, roomID model.RoomID, limit int) ([]dto.MessageView, error) {
	msgs, err := s.messages.FindByRoomID(ctx, roomID, limit)
	if err != nil {
		return nil, err
	}

	views := make([]dto.MessageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, dto.MessageView{
			ID:        m.ID,
			RoomID:    m.RoomID,
			AuthorID:  m.AuthorID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		})
	}
	return views, nil
}

type broadcastMessageNew struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"roomId"`
	Payload dto.MessageView `json:"payload"`
}

func (s *MessageService) broadcastNewMessage(view dto.MessageView) {
	frame := broadcastMessageNew{
		Type:    "message.new",
		RoomID:  string(view.RoomID),
		Payload: view,
	}
	data, _ := json.Marshal(frame)
	// Exclude the sender — they confirm via message.ack, not via message.new
	s.hub.BroadcastToRoomExcept(view.RoomID, view.AuthorID, data)
}

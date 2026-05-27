package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	domainerrors "github.com/maxime/chuchote/domain/errors"
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
		authorName := ""
		if u, err := s.users.FindByID(ctx, m.AuthorID); err == nil {
			authorName = u.Username
		}
		views = append(views, dto.MessageView{
			ID:         m.ID,
			RoomID:     m.RoomID,
			AuthorID:   m.AuthorID,
			AuthorName: authorName,
			Content:    m.Content,
			ClientTempID: m.ClientTempID,
			CreatedAt:  m.CreatedAt,
			EditedAt:   m.EditedAt,
		})
	}
	return views, nil
}

func (s *MessageService) EditMessage(ctx context.Context, cmd dto.EditMessageCommand) (dto.MessageView, error) {
	if cmd.Content == "" || len(cmd.Content) > 4000 {
		return dto.MessageView{}, domainerrors.ErrInvalidInput
	}

	msg, err := s.messages.FindByID(ctx, cmd.MessageID)
	if err != nil {
		return dto.MessageView{}, err
	}
	if msg.AuthorID != cmd.RequestorID {
		return dto.MessageView{}, domainerrors.ErrUnauthorized
	}

	now := time.Now().UTC()
	msg.Content = cmd.Content
	msg.EditedAt = &now

	if err := s.messages.Update(ctx, msg); err != nil {
		return dto.MessageView{}, err
	}

	user, _ := s.users.FindByID(ctx, msg.AuthorID)
	view := dto.MessageView{
		ID:        msg.ID,
		RoomID:    msg.RoomID,
		AuthorID:  msg.AuthorID,
		AuthorName: user.Username,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
		EditedAt:  msg.EditedAt,
	}

	data, _ := json.Marshal(map[string]any{
		"type":   "message.edited",
		"roomId": string(msg.RoomID),
		"payload": map[string]any{
			"messageId": msg.ID,
			"content":   msg.Content,
			"editedAt":  msg.EditedAt,
		},
	})
	s.hub.BroadcastToRoom(msg.RoomID, data)

	return view, nil
}

func (s *MessageService) DeleteMessage(ctx context.Context, cmd dto.DeleteMessageCommand) error {
	msg, err := s.messages.FindByID(ctx, cmd.MessageID)
	if err != nil {
		return err
	}
	if msg.AuthorID != cmd.RequestorID {
		return domainerrors.ErrUnauthorized
	}

	if err := s.messages.Delete(ctx, cmd.MessageID); err != nil {
		return err
	}

	data, _ := json.Marshal(map[string]any{
		"type":   "message.deleted",
		"roomId": string(msg.RoomID),
		"payload": map[string]any{
			"messageId": msg.ID,
		},
	})
	s.hub.BroadcastToRoom(msg.RoomID, data)

	return nil
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

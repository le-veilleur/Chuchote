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
	messages  outbound.MessageRepository
	reactions outbound.ReactionRepository
	hub       outbound.BroadcastHub
	users     outbound.UserRepository
}

func NewMessageService(messages outbound.MessageRepository, reactions outbound.ReactionRepository, hub outbound.BroadcastHub, users outbound.UserRepository) *MessageService {
	return &MessageService{messages: messages, reactions: reactions, hub: hub, users: users}
}

func (s *MessageService) SendMessage(ctx context.Context, cmd dto.SendMessageCommand) (dto.MessageView, error) {
	msg, err := model.NewMessage(
		model.MessageID(uuid.NewString()),
		cmd.RoomID,
		cmd.AuthorID,
		cmd.Content,
		cmd.ClientTempID,
		cmd.ReplyToID,
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
		ReplyToID:    msg.ReplyToID,
		Reactions:    []dto.ReactionView{},
	}

	if msg.ReplyToID != nil {
		if replied, err := s.messages.FindByID(ctx, *msg.ReplyToID); err == nil {
			author, _ := s.users.FindByID(ctx, replied.AuthorID)
			view.ReplyToSummary = &dto.ReplyToSummary{
				AuthorName: author.Username,
				Content:    truncate(replied.Content, 100),
			}
		}
	}

	s.broadcastNewMessage(view)
	return view, nil
}

func (s *MessageService) GetRoomHistory(ctx context.Context, roomID model.RoomID, limit int) ([]dto.MessageView, error) {
	msgs, err := s.messages.FindByRoomID(ctx, roomID, limit)
	if err != nil {
		return nil, err
	}

	msgIDs := make([]model.MessageID, len(msgs))
	for i, m := range msgs {
		msgIDs[i] = m.ID
	}
	reactionsByMsg, _ := s.reactions.FindByMessageIDs(ctx, msgIDs)

	views := make([]dto.MessageView, 0, len(msgs))
	for _, m := range msgs {
		authorName := ""
		if u, err := s.users.FindByID(ctx, m.AuthorID); err == nil {
			authorName = u.Username
		}
		view := dto.MessageView{
			ID:           m.ID,
			RoomID:       m.RoomID,
			AuthorID:     m.AuthorID,
			AuthorName:   authorName,
			Content:      m.Content,
			ClientTempID: m.ClientTempID,
			CreatedAt:    m.CreatedAt,
			EditedAt:     m.EditedAt,
			ReplyToID:    m.ReplyToID,
			Reactions:    aggregateReactions(reactionsByMsg[m.ID]),
		}
		if m.ReplyToID != nil {
			if replied, err := s.messages.FindByID(ctx, *m.ReplyToID); err == nil {
				author, _ := s.users.FindByID(ctx, replied.AuthorID)
				view.ReplyToSummary = &dto.ReplyToSummary{
					AuthorName: author.Username,
					Content:    truncate(replied.Content, 100),
				}
			}
		}
		views = append(views, view)
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
		ID:         msg.ID,
		RoomID:     msg.RoomID,
		AuthorID:   msg.AuthorID,
		AuthorName: user.Username,
		Content:    msg.Content,
		CreatedAt:  msg.CreatedAt,
		EditedAt:   msg.EditedAt,
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

func (s *MessageService) ToggleReaction(ctx context.Context, cmd dto.ToggleReactionCommand) ([]dto.ReactionView, error) {
	if cmd.Emoji == "" {
		return nil, domainerrors.ErrInvalidInput
	}

	if _, err := s.messages.FindByID(ctx, cmd.MessageID); err != nil {
		return nil, err
	}

	if _, err := s.reactions.Toggle(ctx, cmd.MessageID, cmd.UserID, cmd.Emoji); err != nil {
		return nil, err
	}

	allRx, err := s.reactions.FindByMessageIDs(ctx, []model.MessageID{cmd.MessageID})
	if err != nil {
		return nil, err
	}
	views := aggregateReactions(allRx[cmd.MessageID])

	data, _ := json.Marshal(map[string]any{
		"type":   "reaction.updated",
		"roomId": string(cmd.RoomID),
		"payload": map[string]any{
			"messageId": cmd.MessageID,
			"reactions": views,
		},
	})
	s.hub.BroadcastToRoom(cmd.RoomID, data)

	return views, nil
}

func aggregateReactions(rxs []model.Reaction) []dto.ReactionView {
	if len(rxs) == 0 {
		return []dto.ReactionView{}
	}
	byEmoji := make(map[string][]string)
	for _, rx := range rxs {
		byEmoji[rx.Emoji] = append(byEmoji[rx.Emoji], string(rx.UserID))
	}
	views := make([]dto.ReactionView, 0, len(byEmoji))
	for emoji, userIDs := range byEmoji {
		views = append(views, dto.ReactionView{Emoji: emoji, UserIDs: userIDs, Count: len(userIDs)})
	}
	return views
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
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
	s.hub.BroadcastToRoomExcept(view.RoomID, view.AuthorID, data)
}

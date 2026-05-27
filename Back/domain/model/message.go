package model

import (
	"time"

	domainerrors "github.com/maxime/chuchote/domain/errors"
)

type MessageID string

type Message struct {
	ID           MessageID
	RoomID       RoomID
	AuthorID     UserID
	Content      string
	ClientTempID string
	CreatedAt    time.Time
	EditedAt     *time.Time
}

func NewMessage(id MessageID, roomID RoomID, authorID UserID, content, clientTempID string) (Message, error) {
	if content == "" {
		return Message{}, domainerrors.ErrInvalidInput
	}
	if len(content) > 4000 {
		return Message{}, domainerrors.ErrInvalidInput
	}
	return Message{
		ID:           id,
		RoomID:       roomID,
		AuthorID:     authorID,
		Content:      content,
		ClientTempID: clientTempID,
		CreatedAt:    time.Now().UTC(),
	}, nil
}

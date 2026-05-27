package dto

import (
	"time"

	"github.com/maxime/chuchote/domain/model"
)

type SendMessageCommand struct {
	RoomID       model.RoomID
	AuthorID     model.UserID
	AuthorName   string
	Content      string
	ClientTempID string
}

type EditMessageCommand struct {
	MessageID model.MessageID
	RequestorID model.UserID
	Content   string
}

type DeleteMessageCommand struct {
	MessageID   model.MessageID
	RequestorID model.UserID
}

type MessageView struct {
	ID           model.MessageID `json:"id"`
	RoomID       model.RoomID    `json:"roomId"`
	AuthorID     model.UserID    `json:"authorId"`
	AuthorName   string          `json:"authorName"`
	Content      string          `json:"content"`
	ClientTempID string          `json:"clientTempId"`
	CreatedAt    time.Time       `json:"createdAt"`
	EditedAt     *time.Time      `json:"editedAt,omitempty"`
}

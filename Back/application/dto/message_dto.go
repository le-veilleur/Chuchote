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
	ReplyToID    *model.MessageID
}

type EditMessageCommand struct {
	MessageID   model.MessageID
	RequestorID model.UserID
	Content     string
}

type DeleteMessageCommand struct {
	MessageID   model.MessageID
	RequestorID model.UserID
}

type ToggleReactionCommand struct {
	MessageID model.MessageID
	UserID    model.UserID
	RoomID    model.RoomID
	Emoji     string
}

type ReplyToSummary struct {
	AuthorName string `json:"authorName"`
	Content    string `json:"content"`
}

type ReactionView struct {
	Emoji   string   `json:"emoji"`
	UserIDs []string `json:"userIds"`
	Count   int      `json:"count"`
}

type MessageView struct {
	ID             model.MessageID  `json:"id"`
	RoomID         model.RoomID     `json:"roomId"`
	AuthorID       model.UserID     `json:"authorId"`
	AuthorName     string           `json:"authorName"`
	Content        string           `json:"content"`
	ClientTempID   string           `json:"clientTempId"`
	CreatedAt      time.Time        `json:"createdAt"`
	EditedAt       *time.Time       `json:"editedAt,omitempty"`
	ReplyToID      *model.MessageID `json:"replyToId,omitempty"`
	ReplyToSummary *ReplyToSummary  `json:"replyToSummary,omitempty"`
	Reactions      []ReactionView   `json:"reactions"`
}

package dto

import "github.com/maxime/chuchote/domain/model"

type CreateRoomCommand struct {
	Name      string
	CreatorID model.UserID
}

type JoinRoomCommand struct {
	RoomID model.RoomID
	UserID model.UserID
}

type MemberView struct {
	UserID   model.UserID `json:"userId"`
	Username string       `json:"username"`
}

type RoomView struct {
	ID      model.RoomID `json:"id"`
	Name    string       `json:"name"`
	Members []MemberView `json:"members"`
}

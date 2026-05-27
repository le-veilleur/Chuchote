package dto

import "github.com/maxime/chuchote/domain/model"

type RegisterCommand struct {
	Username string
	Password string
}

type LoginCommand struct {
	Username string
	Password string
}

type TokenView struct {
	Token    string       `json:"token"`
	UserID   model.UserID `json:"userId"`
	Username string       `json:"username"`
}

type UserClaims struct {
	UserID   model.UserID `json:"userId"`
	Username string       `json:"username"`
}

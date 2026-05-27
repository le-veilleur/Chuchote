package errors

import "errors"

var (
	ErrRoomNotFound    = errors.New("room not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrMessageNotFound = errors.New("message not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token expired")
	ErrInvalidInput    = errors.New("invalid input")
	ErrUsernameExists  = errors.New("username already exists")
)

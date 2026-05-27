package model

import "time"

type UserID string

type User struct {
	ID           UserID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

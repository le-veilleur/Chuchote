package model

import "time"

type RoomID string

type Room struct {
	ID        RoomID
	Name      string
	Members   []UserID
	CreatedAt time.Time
}

func (r *Room) HasMember(userID UserID) bool {
	for _, id := range r.Members {
		if id == userID {
			return true
		}
	}
	return false
}

func (r *Room) AddMember(userID UserID) {
	if !r.HasMember(userID) {
		r.Members = append(r.Members, userID)
	}
}

func (r *Room) RemoveMember(userID UserID) {
	members := make([]UserID, 0, len(r.Members))
	for _, id := range r.Members {
		if id != userID {
			members = append(members, id)
		}
	}
	r.Members = members
}

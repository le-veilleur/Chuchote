package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/maxime/chuchote/application/dto"
	domainerrors "github.com/maxime/chuchote/domain/errors"
	"github.com/maxime/chuchote/domain/model"
	"github.com/maxime/chuchote/port/outbound"
)

type RoomService struct {
	rooms outbound.RoomRepository
	users outbound.UserRepository
}

func NewRoomService(rooms outbound.RoomRepository, users outbound.UserRepository) *RoomService {
	return &RoomService{rooms: rooms, users: users}
}

func (s *RoomService) CreateRoom(ctx context.Context, cmd dto.CreateRoomCommand) (dto.RoomView, error) {
	if cmd.Name == "" {
		return dto.RoomView{}, domainerrors.ErrInvalidInput
	}
	room := model.Room{
		ID:      model.RoomID(uuid.NewString()),
		Name:    cmd.Name,
		Members: []model.UserID{cmd.CreatorID},
	}
	if err := s.rooms.Save(ctx, room); err != nil {
		return dto.RoomView{}, err
	}
	return s.toView(ctx, room)
}

func (s *RoomService) JoinRoom(ctx context.Context, cmd dto.JoinRoomCommand) (dto.RoomView, error) {
	room, err := s.rooms.FindByID(ctx, cmd.RoomID)
	if err != nil {
		return dto.RoomView{}, err
	}
	room.AddMember(cmd.UserID)
	if err := s.rooms.Update(ctx, room); err != nil {
		return dto.RoomView{}, err
	}
	return s.toView(ctx, room)
}

func (s *RoomService) LeaveRoom(ctx context.Context, userID model.UserID, roomID model.RoomID) error {
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	room.RemoveMember(userID)
	return s.rooms.Update(ctx, room)
}

func (s *RoomService) ListRooms(ctx context.Context) ([]dto.RoomView, error) {
	rooms, err := s.rooms.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]dto.RoomView, 0, len(rooms))
	for _, r := range rooms {
		v, err := s.toView(ctx, r)
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, nil
}

func (s *RoomService) GetRoom(ctx context.Context, roomID model.RoomID) (dto.RoomView, error) {
	room, err := s.rooms.FindByID(ctx, roomID)
	if err != nil {
		return dto.RoomView{}, err
	}
	return s.toView(ctx, room)
}

func (s *RoomService) toView(ctx context.Context, room model.Room) (dto.RoomView, error) {
	members := make([]dto.MemberView, 0, len(room.Members))
	for _, uid := range room.Members {
		u, err := s.users.FindByID(ctx, uid)
		if err != nil {
			continue
		}
		members = append(members, dto.MemberView{UserID: u.ID, Username: u.Username})
	}
	return dto.RoomView{ID: room.ID, Name: room.Name, Members: members}, nil
}

package outbound

import "github.com/maxime/chuchote/domain/model"

type BroadcastHub interface {
	Register(conn model.Connection, send chan<- []byte)
	Unregister(conn model.Connection)
	BroadcastToRoom(roomID model.RoomID, payload []byte)
	BroadcastToRoomExcept(roomID model.RoomID, excludeUserID model.UserID, payload []byte)
	SendToUser(userID model.UserID, payload []byte)
	SubscribeToRoom(conn model.Connection, roomID model.RoomID)
	UnsubscribeFromRoom(conn model.Connection, roomID model.RoomID)
}

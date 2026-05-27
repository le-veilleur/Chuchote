package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/maxime/chuchote/application/dto"
	"github.com/maxime/chuchote/domain/model"
	"github.com/maxime/chuchote/port/inbound"
	"github.com/maxime/chuchote/port/outbound"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const sendBufSize = 256

type Client struct {
	conn            *websocket.Conn
	send            chan []byte
	hub             outbound.BroadcastHub
	messages        inbound.MessageUseCase
	rooms           inbound.RoomUseCase
	auth            inbound.AuthUseCase
	subscribedRooms []model.RoomID
}

func newClient(conn *websocket.Conn, hub outbound.BroadcastHub, messages inbound.MessageUseCase, rooms inbound.RoomUseCase, auth inbound.AuthUseCase) *Client {
	return &Client{
		conn:     conn,
		send:     make(chan []byte, sendBufSize),
		hub:      hub,
		messages: messages,
		rooms:    rooms,
		auth:     auth,
	}
}

func (c *Client) run(ctx context.Context) {
	connID := model.ConnID(uuid.NewString())
	var userClaims dto.UserClaims
	var authenticated bool

	defer func() {
		if authenticated {
			conn := model.Connection{ID: connID, UserID: userClaims.UserID}
			// Notify all subscribed rooms that this user went offline
			for _, roomID := range c.subscribedRooms {
				c.hub.UnsubscribeFromRoom(conn, roomID)
				c.broadcastRoomCount(roomID)
			}
			c.hub.Unregister(conn)
		}
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	go c.writePump(ctx)

	for {
		var raw json.RawMessage
		if err := wsjson.Read(ctx, c.conn, &raw); err != nil {
			return
		}

		frame, err := parseFrame(raw)
		if err != nil {
			c.sendError("", "", "PARSE_ERROR", "invalid frame")
			continue
		}

		if !authenticated && frame.Type != "auth.connect" {
			c.sendError(frame.RequestID, "", "NOT_AUTHENTICATED", "send auth.connect first")
			continue
		}

		switch frame.Type {
		case "auth.connect":
			var p AuthConnectPayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, "", "PARSE_ERROR", "invalid payload")
				continue
			}
			claims, err := c.auth.ValidateToken(ctx, p.Token)
			if err != nil {
				c.sendErrorAndClose(frame.RequestID, err.Error())
				return
			}
			userClaims = claims
			authenticated = true
			conn := model.Connection{ID: connID, UserID: claims.UserID}
			c.hub.Register(conn, c.send)
			c.sendJSON(map[string]any{
				"type":      "auth.connected",
				"requestId": frame.RequestID,
				"roomId":    nil,
				"payload": map[string]any{
					"userId":   claims.UserID,
					"username": claims.Username,
				},
			})

		case "room.join":
			roomView, err := c.rooms.JoinRoom(ctx, dto.JoinRoomCommand{
				RoomID: model.RoomID(frame.RoomID),
				UserID: userClaims.UserID,
			})
			if err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "JOIN_FAILED", err.Error())
				continue
			}
			conn := model.Connection{ID: connID, UserID: userClaims.UserID}
			c.hub.SubscribeToRoom(conn, model.RoomID(frame.RoomID))
			c.addSubscribedRoom(model.RoomID(frame.RoomID))

			history, _ := c.messages.GetRoomHistory(ctx, model.RoomID(frame.RoomID), 50)
			onlineCount := c.hub.CountRoomSubscribers(model.RoomID(frame.RoomID))
			c.sendJSON(map[string]any{
				"type":      "room.joined",
				"requestId": frame.RequestID,
				"roomId":    frame.RoomID,
				"payload": map[string]any{
					"room":        roomView,
					"history":     history,
					"onlineCount": onlineCount,
				},
			})
			// Notify everyone in the room of the new count
			c.broadcastRoomCount(model.RoomID(frame.RoomID))

		case "room.leave":
			if err := c.rooms.LeaveRoom(ctx, userClaims.UserID, model.RoomID(frame.RoomID)); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "LEAVE_FAILED", err.Error())
				continue
			}
			conn := model.Connection{ID: connID, UserID: userClaims.UserID}
			c.hub.UnsubscribeFromRoom(conn, model.RoomID(frame.RoomID))
			c.removeSubscribedRoom(model.RoomID(frame.RoomID))
			c.broadcastRoomCount(model.RoomID(frame.RoomID))

		case "message.send":
			var p MessageSendPayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "PARSE_ERROR", "invalid payload")
				continue
			}
			cmd := dto.SendMessageCommand{
				RoomID:       model.RoomID(frame.RoomID),
				AuthorID:     userClaims.UserID,
				AuthorName:   userClaims.Username,
				Content:      p.Content,
				ClientTempID: p.ClientTempID,
			}
			if p.ReplyToID != nil {
				msgID := model.MessageID(*p.ReplyToID)
				cmd.ReplyToID = &msgID
			}
			view, err := c.messages.SendMessage(ctx, cmd)
			if err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "SEND_FAILED", err.Error())
				continue
			}
			c.sendJSON(map[string]any{
				"type":      "message.ack",
				"requestId": frame.RequestID,
				"roomId":    frame.RoomID,
				"payload": map[string]any{
					"messageId":      view.ID,
					"clientTempId":   view.ClientTempID,
					"createdAt":      view.CreatedAt,
					"replyToId":      view.ReplyToID,
					"replyToSummary": view.ReplyToSummary,
				},
			})

		case "message.edit":
			var p MessageEditPayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "PARSE_ERROR", "invalid payload")
				continue
			}
			_, err := c.messages.EditMessage(ctx, dto.EditMessageCommand{
				MessageID:   model.MessageID(p.MessageID),
				RequestorID: userClaims.UserID,
				Content:     p.Content,
			})
			if err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "EDIT_FAILED", err.Error())
			}

		case "message.delete":
			var p MessageDeletePayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "PARSE_ERROR", "invalid payload")
				continue
			}
			if err := c.messages.DeleteMessage(ctx, dto.DeleteMessageCommand{
				MessageID:   model.MessageID(p.MessageID),
				RequestorID: userClaims.UserID,
			}); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "DELETE_FAILED", err.Error())
			}

		case "reaction.toggle":
			var p ReactionTogglePayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "PARSE_ERROR", "invalid payload")
				continue
			}
			if _, err := c.messages.ToggleReaction(ctx, dto.ToggleReactionCommand{
				MessageID: model.MessageID(p.MessageID),
				UserID:    userClaims.UserID,
				RoomID:    model.RoomID(frame.RoomID),
				Emoji:     p.Emoji,
			}); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "REACTION_FAILED", err.Error())
			}

		case "typing.start", "typing.stop":
			isTyping := frame.Type == "typing.start"
			data, _ := json.Marshal(map[string]any{
				"type":      "typing.indicator",
				"requestId": nil,
				"roomId":    frame.RoomID,
				"payload": map[string]any{
					"userId":   userClaims.UserID,
					"username": userClaims.Username,
					"isTyping": isTyping,
				},
			})
			room, err := c.rooms.GetRoom(ctx, model.RoomID(frame.RoomID))
			if err == nil {
				for _, m := range room.Members {
					if m.UserID != userClaims.UserID {
						c.hub.SendToUser(m.UserID, data)
					}
				}
			}

		default:
			slog.Warn("unknown ws event type", "type", frame.Type)
		}
	}
}

func (c *Client) broadcastRoomCount(roomID model.RoomID) {
	count := c.hub.CountRoomSubscribers(roomID)
	data, _ := json.Marshal(map[string]any{
		"type":   "room.online_count",
		"roomId": string(roomID),
		"payload": map[string]any{
			"count": count,
		},
	})
	c.hub.BroadcastToRoom(roomID, data)
}

func (c *Client) addSubscribedRoom(roomID model.RoomID) {
	for _, r := range c.subscribedRooms {
		if r == roomID {
			return
		}
	}
	c.subscribedRooms = append(c.subscribedRooms, roomID)
}

func (c *Client) removeSubscribedRoom(roomID model.RoomID) {
	for i, r := range c.subscribedRooms {
		if r == roomID {
			c.subscribedRooms = append(c.subscribedRooms[:i], c.subscribedRooms[i+1:]...)
			return
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.Write(ctx, websocket.MessageText, msg); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) sendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendError(requestID, roomID, code, message string) {
	c.sendJSON(map[string]any{
		"type":      "error",
		"requestId": requestID,
		"roomId":    roomID,
		"payload":   map[string]any{"code": code, "message": message},
	})
}

func (c *Client) sendErrorAndClose(requestID, message string) {
	c.sendJSON(map[string]any{
		"type":      "auth.error",
		"requestId": requestID,
		"roomId":    nil,
		"payload":   map[string]any{"code": "INVALID_TOKEN", "message": message},
	})
}

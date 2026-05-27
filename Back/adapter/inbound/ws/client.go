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
	conn     *websocket.Conn
	send     chan []byte
	hub      outbound.BroadcastHub
	messages inbound.MessageUseCase
	rooms    inbound.RoomUseCase
	auth     inbound.AuthUseCase
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
			c.hub.Unregister(model.Connection{ID: connID, UserID: userClaims.UserID})
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
			c.hub.SubscribeToRoom(model.Connection{ID: connID, UserID: userClaims.UserID}, model.RoomID(frame.RoomID))

			history, _ := c.messages.GetRoomHistory(ctx, model.RoomID(frame.RoomID), 50)
			c.sendJSON(map[string]any{
				"type":      "room.joined",
				"requestId": frame.RequestID,
				"roomId":    frame.RoomID,
				"payload": map[string]any{
					"room":    roomView,
					"history": history,
				},
			})

		case "room.leave":
			if err := c.rooms.LeaveRoom(ctx, userClaims.UserID, model.RoomID(frame.RoomID)); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "LEAVE_FAILED", err.Error())
				continue
			}
			c.hub.UnsubscribeFromRoom(model.Connection{ID: connID, UserID: userClaims.UserID}, model.RoomID(frame.RoomID))

		case "message.send":
			var p MessageSendPayload
			if err := json.Unmarshal(frame.Payload, &p); err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "PARSE_ERROR", "invalid payload")
				continue
			}
			view, err := c.messages.SendMessage(ctx, dto.SendMessageCommand{
				RoomID:       model.RoomID(frame.RoomID),
				AuthorID:     userClaims.UserID,
				AuthorName:   userClaims.Username,
				Content:      p.Content,
				ClientTempID: p.ClientTempID,
			})
			if err != nil {
				c.sendError(frame.RequestID, frame.RoomID, "SEND_FAILED", err.Error())
				continue
			}
			c.sendJSON(map[string]any{
				"type":      "message.ack",
				"requestId": frame.RequestID,
				"roomId":    frame.RoomID,
				"payload": map[string]any{
					"messageId":    view.ID,
					"clientTempId": view.ClientTempID,
					"createdAt":    view.CreatedAt,
				},
			})

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
			// Send to each room member except the sender
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

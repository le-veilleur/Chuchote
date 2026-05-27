package ws

import "encoding/json"

type WSFrame struct {
	Type      string          `json:"type"`
	RequestID string          `json:"requestId"`
	RoomID    string          `json:"roomId"`
	Payload   json.RawMessage `json:"payload"`
}

type AuthConnectPayload struct {
	Token string `json:"token"`
}

type RoomJoinPayload struct{}
type RoomLeavePayload struct{}

type MessageSendPayload struct {
	Content      string  `json:"content"`
	ClientTempID string  `json:"clientTempId"`
	ReplyToID    *string `json:"replyToId,omitempty"`
}

type MessageEditPayload struct {
	MessageID string `json:"messageId"`
	Content   string `json:"content"`
}

type MessageDeletePayload struct {
	MessageID string `json:"messageId"`
}

type ReactionTogglePayload struct {
	MessageID string `json:"messageId"`
	Emoji     string `json:"emoji"`
}

type TypingPayload struct{}

func parseFrame(data []byte) (WSFrame, error) {
	var f WSFrame
	return f, json.Unmarshal(data, &f)
}

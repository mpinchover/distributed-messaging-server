package requests

import (
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Websocket struct {
	Conn       *websocket.Conn
	DeviceUUID *string
	UserUUID   *string
	Outbound   chan interface{} // channel to send messages back to client socket
}

type Member struct {
	UserUUID string `json:"userUuid"`
	RoomUUID string `json:"roomUuid"`
}

type Room struct {
	UUID          string     `json:"uuid"`
	Members       []*Member  `json:"members"`
	Messages      []*Message `json:"messages"`
	CreatedAtNano float64    `json:"createdAtNano"`
}

type Message struct {
	UUID          string    `json:"uuid"`
	UserUUID      string    `json:"userUuid"`
	RoomUUID      string    `json:"roomUuid"`
	MessageText   string    `json:"messageText"`
	CreatedAtNano float64   `json:"createdAtNano"`
	MessageStatus string  `json:"messageStatus"`
}

type Connection struct {
	UserUUID    string
	Connections map[string]*websocket.Conn
}

// server specific room
type ServerChannel struct {
	Subscriber *redis.PubSub

	// just the participants that are on this server
	MembersOnServer map[string]bool
	UUID            string
}

type User struct {
	UserID string
}

type APIKey struct {
	Key string
}

// for ext service, not chat user
// type AuthProfile struct {
// 	UserID uint
// 	Email  string
// }

type ChatProfile struct {
	UserUUID string
}

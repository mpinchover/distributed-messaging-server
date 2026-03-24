package requests

type DeleteRoomEvent struct {
	EventType string `json:"eventType" validate:"required"`
	RoomUUID  string `json:"roomUuid" validate:"required"`
}

type LeaveRoomEvent struct {
	EventType string `json:"eventType" validate:"required"`
	RoomUUID  string `json:"roomUuid" validate:"required"`
	UserUUID  string `json:"userUuid"  validate:"required"`
	Token     string `json:"token"`
}

// sennd to clients room has been opened
type OpenRoomEvent struct {
	EventType string `json:"eventType" validate:"required"`
	Room      *Room  `json:"room" validate:"required"`
}

// subscrve the sever to a room
type SubscribeToRoomEvent struct {
	EventType string   `json:"eventType" validate:"required"`
	Channel   string   `json:"channel" validate:"required"`
	Members   []string `json:"members" validate:"gte=2,required"`
}

type SetClientConnectionEvent struct {
	EventType  string `json:"eventType" validate:"required"`
	UserUUID   string `json:"userUuid" validate:"required"`
	DeviceUUID string `json:"deviceUuid" validate:"required"`
	Token      string `json:"token"`
}

type TextMessageEvent struct {
	EventType  string   `json:"eventType" validate:"required"`
	UserUUID   string   `json:"userUuid" validate:"required"`
	DeviceUUID string   `json:"deviceUuid" validate:"required"`
	Message    *Message `json:"message" validate:"required"`
	Token      string   `json:"token"`
}

// type RoomsByUserUUIDEvent struct {
// 	EventType string          `json:"eventType"`
// 	UserUUID  string          `schema:"userUuid" validate:"required"`
// 	Offset    int             `schema:"offset"`
// 	Key       string          `schema:"key,-"`
// 	Rooms     []*records.Room `json:"rooms"`
// 	Token     string          `json:"token"`
// }

// type MessagesByRoomUUIDEvent struct {
// 	EventType string             `json:"eventType"`
// 	UserUUID  string             `schema:"userUuid" validate:"required"`
// 	RoomUUID  string             `schema:"roomUuid" validate:"required"`
// 	Offset    int                `schema:"offset"`
// 	Messages  []*records.Message `json:"messages"` // maybe make everything the actual record?
// 	Token     string             `json:"token"`
// }

package requests

type DeleteRoomEvent struct {
	EventType string `json:"eventType"`
	RoomUUID  string `json:"roomUuid"`
}

type LeaveRoomEvent struct {
	EventType string `json:"eventType"`
	RoomUUID  string `json:"roomUuid"`
	UserUUID  string `json:"userUuid"`
}

// sennd to clients room has been opened
type OpenRoomEvent struct {
	EventType string `json:"eventType"`
	Room      *Room  `json:"room"`
}

// subscrve the sever to a room
type SubscribeToRoomEvent struct {
	EventType string   `json:"eventType"`
	Channel   string   `json:"channel"`
	Members   []string `json:"members"`
}

type SetClientConnectionEvent struct {
	EventType      string `json:"eventType"`
	FromUUID       string `json:"fromUuid"`
	ConnectionUUID string `json:"connectionUuid"`
}

type TextMessageEvent struct {
	EventType      string `json:"eventType"`
	FromUUID       string `json:"fromUuid"`
	ConnectionUUID string `json:"connectionUuid"`
	RoomUUID       string `json:"roomUuid"`
	MessageText    string `json:"messageText"`
	CreatedAt      int64  `json:"createdAt"`
	MessageUUID    string `json:"messageUuid"`
}

// the recpt has read the message
// client will have the user uuid stored. If the message is opened
// by not owner user uuid, send out the event
type SeenMessageEvent struct {
	EventType   string `json:"eventType"`
	MessageUUID string `json:"messageUuid"`
	UserUUID    string `json:"userUuid"`
	RoomUUID    string `json:"roomUuid"`
}

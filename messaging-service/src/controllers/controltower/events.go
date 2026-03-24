package controltower

import (
	"encoding/json"
	"errors"

	"messaging-service/src/types/enums"
	"messaging-service/src/types/records"
	"messaging-service/src/types/requests"
	"time"

	"github.com/google/uuid"
)

// TODO – event should just have the message embedded within it
func (c *ControlTowerCtrlr) ProcessTextMessage(msg *requests.TextMessageEvent) (*records.Message, error) {
	// ensure room exists
	room, err := c.Repo.GetRoomByRoomUUID(msg.Message.RoomUUID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("room does not exist")
	}

	msgUUID := uuid.New().String()
	msg.Message.UUID = msgUUID

	createdAtNano := time.Now().UnixNano() // / 1e6

	repoMessage := &records.Message{
		UserUUID:      msg.UserUUID,
		RoomUUID:      msg.Message.RoomUUID,
		RoomID:        0,
		MessageText:   msg.Message.MessageText,
		UUID:          msgUUID,
		MessageStatus: enums.MESSAGE_STATUS_LIVE.String(),
		CreatedAtNano: float64(createdAtNano),
	}

	err = c.Repo.SaveMessage(repoMessage)
	if err != nil {
		return nil, err
	}

	// requestsMessage := mappers.FromRecordsMessageToRequestMessage(repoMessage)
	msg.Message.CreatedAtNano = float64(createdAtNano)

	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	err = c.RedisClient.PublishToRedisChannel(msg.Message.RoomUUID, bytes)
	if err != nil {
		return nil, err
	}
	return repoMessage, nil
}

package mappers

import (
	"messaging-service/src/types/records"
	"messaging-service/src/types/requests"
)

func ToRequestMessages(in []*records.Message) []*requests.Message {
	msgs := make([]*requests.Message, len(in))
	for i, val := range in {
		msgs[i] = ToRequestMessage(val)
	}
	return msgs
}

func ToRequestMessage(msg *records.Message) *requests.Message {
	return &requests.Message{
		UserUUID:      msg.UserUUID,
		UUID:          msg.UUID,
		RoomUUID:      msg.RoomUUID,
		MessageText:   msg.MessageText,
		MessageStatus: msg.MessageStatus,
		CreatedAtNano: msg.CreatedAtNano,
	}
}

func ToRecordMessages(in []*requests.Message) []*records.Message {
	msgs := make([]*records.Message, len(in))
	for i, val := range in {
		msgs[i] = ToRecordMessage(val)
	}
	return msgs
}

func ToRecordMessage(msg *requests.Message) *records.Message {
	return &records.Message{
		UserUUID:      msg.UserUUID,
		UUID:          msg.UUID,
		RoomUUID:      msg.RoomUUID,
		MessageText:   msg.MessageText,
		MessageStatus: msg.MessageStatus,
		CreatedAtNano: msg.CreatedAtNano,
	}
}

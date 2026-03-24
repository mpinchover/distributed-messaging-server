package handlers

import (
	"encoding/json"
	"messaging-service/src/types/connections"
	"messaging-service/src/types/requests"
	"messaging-service/src/utils"

	"github.com/gorilla/websocket"
)

func (h *Handler) handleClientEventTextMessage(conn *websocket.Conn, p []byte) error {
	msg := &requests.TextMessageEvent{}
	err := json.Unmarshal(p, msg)
	if err != nil {
		return err
	}
	// break this up into processTextMessage and SaveTextMessage
	_, err = h.ControlTowerCtrlr.ProcessTextMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) handleSetClientSocket(ws *requests.Websocket, p []byte) error {
	// TODO – have a new event that doesn't include the deviceUUID
	msg := &requests.SetClientConnectionEvent{}
	err := json.Unmarshal(p, msg)
	if err != nil {
		return err
	}
	resp, err := h.ControlTowerCtrlr.SetupClientConnectionV2(ws, msg)
	if err != nil {
		return err
	}

	userExistingRooms, err := h.ControlTowerCtrlr.GetRoomsByUserUUIDForSubscribing(msg.UserUUID)
	if err != nil {
		return err
	}

	for _, room := range userExistingRooms {
		_, ok := h.ControlTowerCtrlr.Channels[room.UUID]
		if !ok {
			// subscribe the room
			subscriber := utils.SetupChannel(h.RedisClient, room.UUID)
			go utils.SubscribeToChannel(subscriber, h.HandleRoomEvent)
			h.ControlTowerCtrlr.Channels[room.UUID] = &connections.Channel{
				UUID:       room.UUID,
				Users:      map[string]bool{},
				Subscriber: subscriber,
			}
		}
		h.ControlTowerCtrlr.Channels[room.UUID].Users[msg.UserUUID] = true
	}

	ws.DeviceUUID = &resp.DeviceUUID
	ws.UserUUID = &resp.UserUUID
	ws.Outbound <- resp
	return nil
}

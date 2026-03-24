package handlers

import (
	"errors"
	"sync"
)

// broadcast to all channel members excluding the client device
func (h *Handler) BroadcastEventToChannelSubscribersDeviceExclusive(channelUUID string, fromDeviceUUID string, msg interface{}) error {

	// get the room from the server
	_, ok := h.ControlTowerCtrlr.Channels[channelUUID]
	if !ok {
		return errors.New("room not found on server")
	}

	// if the user connection is on this server, blast it out.
	members, err := h.ControlTowerCtrlr.Repo.GetMembersByRoomUUID(channelUUID)
	if err != nil {
		return err
	}

	var mu = &sync.RWMutex{}

	for _, m := range members {
		mu.RLock()
		userConn, ok := h.ControlTowerCtrlr.UserConnections[m.UserUUID]
		mu.RUnlock()
		if !ok {
			continue
		}
		for deviceUUID, device := range userConn.Devices {
			if deviceUUID == fromDeviceUUID {
				continue
			}

			// device.WS.WriteJSON(msg)
			device.Outbound <- msg

		}
	}
	return nil
}

func (h *Handler) BroadcastEventToChannelSubscribers(channelUUID string, msg interface{}) error {

	var mu = &sync.RWMutex{}
	// get the room from the server
	mu.RLock()
	_, ok := h.ControlTowerCtrlr.UserConnections[channelUUID]
	mu.RUnlock()
	// room not on server
	if !ok {
		return errors.New("room not found on server")
	}

	// if the user connection is on this server, blast it out.
	members, err := h.ControlTowerCtrlr.Repo.GetMembersByRoomUUID(channelUUID)
	if err != nil {
		return err
	}

	for _, m := range members {
		mu.RLock()
		userConn, ok := h.ControlTowerCtrlr.UserConnections[m.UserUUID]
		mu.RUnlock()
		if !ok {
			continue
		}

		for _, device := range userConn.Devices {
			// device.WS.WriteJSON(msg)
			device.Outbound <- msg
		}
	}
	return nil
}

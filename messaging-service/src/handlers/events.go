package handlers

import (
	"encoding/json"
	"log"
	"messaging-service/src/types/connections"
	"messaging-service/src/types/enums"
	"messaging-service/src/types/requests"
	"messaging-service/src/utils"
)

func (h *Handler) SetupChannels() {
	subscriber := utils.SetupChannel(h.RedisClient, enums.CHANNEL_SERVER_EVENTS)
	go utils.SubscribeToChannel(subscriber, h.HandleServerEvent)
}

func (h *Handler) HandleServerEvent(event string) error {
	eventType, err := utils.GetEventType(event)
	if err != nil {
		log.Println(err)
		return err
	}

	if eventType == enums.EVENT_OPEN_ROOM.String() {
		openRoomEvent := &requests.OpenRoomEvent{}
		err = json.Unmarshal([]byte(event), openRoomEvent)
		if err != nil {
			log.Println(err)
			return err
		}
		err = h.handleOpenRoomEvent(openRoomEvent)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (h *Handler) HandleRoomEvent(event string) error {
	eventType, err := utils.GetEventType(event)
	if err != nil {
		log.Println(err)
		return err
	}

	if eventType == enums.EVENT_TEXT_MESSAGE.String() {
		textMessageEvent := &requests.TextMessageEvent{}
		err = json.Unmarshal([]byte(event), textMessageEvent)
		if err != nil {
			log.Println(err)
			return err
		}

		err = h.BroadcastEventToChannelSubscribersDeviceExclusive(
			textMessageEvent.Message.RoomUUID,
			textMessageEvent.DeviceUUID,
			textMessageEvent,
		)
		if err != nil {
			log.Println(err)
		}
		return err
	}

	if eventType == enums.EVENT_DELETE_ROOM.String() {
		deleteRoomEvent := &requests.DeleteRoomEvent{}
		err = json.Unmarshal([]byte(event), deleteRoomEvent)
		if err != nil {
			log.Println(err)
			return err
		}
		err = h.handleDeleteRoomEvent(deleteRoomEvent)
		if err != nil {
			log.Println(err)
		}
		return err
	}

	return nil
}

// this is an event that has been received by redis
func (h *Handler) handleDeleteRoomEvent(event *requests.DeleteRoomEvent) error {
	// get the room from the server
	channel := h.ControlTowerCtrlr.GetChannelFromServer(event.RoomUUID)

	// if channel not on server
	if channel == nil {
		return nil
	}

	h.ControlTowerCtrlr.DeleteChannelFromServer(event.RoomUUID)

	for userUUID := range channel.Users {
		userConn := h.ControlTowerCtrlr.GetUserConnection(userUUID)
		if userConn == nil {
			continue
		}

		// notify everyone that the channel has closed
		for _, device := range userConn.Devices {
			device.Outbound <- event
		}
	}
	return nil
}

func (h *Handler) handleOpenRoomEvent(event *requests.OpenRoomEvent) error {

	// for every member, check if they are on this server
	// if they are, then you need to subscribe the server to the channel
	members := event.Room.Members
	roomUUID := event.Room.UUID

	memberDevicesOnThisChannel := []*connections.Device{}
	// subscribe server to the room if members on are on this server
	for _, member := range members {
		userConn := h.ControlTowerCtrlr.GetUserConnection(member.UserUUID)
		if userConn == nil {
			continue
		}

		channel := h.ControlTowerCtrlr.GetChannelFromServer(roomUUID)

		// server contains a user who doesn't have the room subscribed
		if channel == nil {

			// Set up the room on this server
			// TODO - add a recover here
			// TODO - log out errors?
			subscriber := utils.SetupChannel(h.RedisClient, roomUUID)
			go utils.SubscribeToChannel(subscriber, h.HandleRoomEvent)

			newChannel := &connections.Channel{
				UUID:       roomUUID,
				Subscriber: subscriber,
				Users:      map[string]bool{},
			}
			err := h.ControlTowerCtrlr.SetChannelOnServer(roomUUID, newChannel)
			if err != nil {
				return err
			}
		}

		// TODO - get rid of member.UUID
		// add the member on this server to the channel on this server
		err := h.ControlTowerCtrlr.AddUserToChannel(member.UserUUID, roomUUID)
		if err != nil {
			return err
		}

		for _, device := range userConn.Devices {
			memberDevicesOnThisChannel = append(memberDevicesOnThisChannel, device)
		}
	}

	// write open room event to all member devices
	for _, d := range memberDevicesOnThisChannel {
		d.Outbound <- event
	}
	return nil
}

package controltower

import (
	"encoding/json"
	"errors"

	"messaging-service/controllers/channelscontroller"
	"messaging-service/controllers/connectionscontroller"
	redisClient "messaging-service/redis"
	"messaging-service/repo"
	"messaging-service/types/enums"
	"messaging-service/types/records"
	"messaging-service/types/requests"
	"messaging-service/utils"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ControlTowerCtrlr struct {
	RedisClient   *redisClient.RedisClient
	ConnCtrlr     *connectionscontroller.ConnectionsController
	ChannelsCtrlr *channelscontroller.ChannelsController
	// track active rooms/channels on this server
	ServerChannels map[string]*requests.ServerChannel

	MapLock *sync.Mutex

	Repo *repo.Repo
}

func New(
	redisClient *redisClient.RedisClient,
	repo *repo.Repo,
	connCtrlr *connectionscontroller.ConnectionsController,
	channelsCtrlr *channelscontroller.ChannelsController,
) *ControlTowerCtrlr {
	// connections := map[string]*requests.Connection{}
	// serverChannels := map[string]*requests.ServerChannel{}

	var mu sync.Mutex
	controlTower := &ControlTowerCtrlr{
		RedisClient:   redisClient,
		ConnCtrlr:     connCtrlr,
		ChannelsCtrlr: channelsCtrlr,

		Repo:    repo,
		MapLock: &mu,
	}

	return controlTower
}

func (c *ControlTowerCtrlr) GetMessagesByRoomUUID(roomUUID string, offset int) ([]*records.Message, error) {
	return c.Repo.GetMessagesByRoomUUID(roomUUID, offset)
}

func (c *ControlTowerCtrlr) CreateRoom(
	members []*requests.Member,
) (*requests.Room, error) {
	// build the room
	for _, member := range members {
		member.UUID = uuid.New().String()
	}

	roomUUID := uuid.New().String()
	repoMembers := make([]*records.Member, len(members))

	// TODO – if member is nil, make it a default type
	for i, member := range members {
		repoMembers[i] = &records.Member{
			UUID:     member.UUID,
			RoomUUID: roomUUID,
			UserUUID: member.UserUUID,
		}
	}

	repoRoom := &records.Room{
		UUID:    roomUUID,
		Members: repoMembers,
	}

	err := c.Repo.SaveRoom(repoRoom)
	if err != nil {
		return nil, err
	}

	newRoom := &requests.Room{
		Members: members,
		UUID:    roomUUID,
	}

	openRoomEvent := requests.OpenRoomEvent{
		EventType: enums.EVENT_OPEN_ROOM.String(),
		Room:      newRoom,
	}

	err = utils.PublishToRedisChannel(c.RedisClient, enums.CHANNEL_SERVER_EVENTS, openRoomEvent)
	if err != nil {
		return nil, err
	}
	return newRoom, nil
}

func (c *ControlTowerCtrlr) UpdateMessage(message *requests.Message) error {
	// first get the message
	existingMsg, err := c.Repo.GetMessageByUUID(message.UUID)
	if err != nil {
		return err
	}

	// if we haven't already deleted the message and want to delete it
	if existingMsg.MessageStatus != enums.MESSAGE_STATUS_DELETED.String() &&
		message.MessageStatus != existingMsg.MessageStatus {
		existingMsg.MessageStatus = message.MessageStatus
	}

	return c.Repo.UpdateMessage(existingMsg)
}

func (c *ControlTowerCtrlr) LeaveRoom(userUUID string, roomUUID string) error {
	room, err := c.Repo.GetRoomByRoomUUID(roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room does not exist")
	}

	// TODO – this is something the client should verify not the server

	// TODO - put in helper function
	// TODO – in the future add in fn to make this optional
	if len(room.Members) == 1 {
		err := c.Repo.DeleteRoom(roomUUID)
		if err != nil {
			return err
		}
		deleteRoomEvent := requests.DeleteRoomEvent{
			EventType: enums.EVENT_DELETE_ROOM.String(),
			RoomUUID:  roomUUID,
		}

		msgBytes, err := json.Marshal(deleteRoomEvent)
		if err != nil {
			return err
		}
		c.RedisClient.PublishToRedisChannel(roomUUID, msgBytes)
		return nil
	}

	err = c.Repo.LeaveRoom(userUUID, roomUUID)
	if err != nil {
		return err
	}
	leaveRoomEvent := requests.LeaveRoomEvent{
		EventType: enums.EVENT_LEAVE_ROOM.String(),
		RoomUUID:  roomUUID,
		UserUUID:  userUUID,
	}
	msgBytes, err := json.Marshal(leaveRoomEvent)
	if err != nil {
		return err
	}

	c.RedisClient.PublishToRedisChannel(roomUUID, msgBytes)
	return nil
}

func (c *ControlTowerCtrlr) DeleteRoom(roomUUID string) error {
	room, err := c.Repo.GetRoomByRoomUUID(roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("room not found")
	}

	// put in helper function
	membersInRoom := make([]string, len(room.Members))
	for i, mem := range room.Members {
		membersInRoom[i] = mem.UserUUID
	}

	err = c.Repo.DeleteRoom(roomUUID)
	if err != nil {
		return err
	}

	deleteRoomEvent := requests.DeleteRoomEvent{
		EventType: enums.EVENT_DELETE_ROOM.String(),
		RoomUUID:  roomUUID,
	}
	msgBytes, err := json.Marshal(deleteRoomEvent)
	if err != nil {
		return err
	}

	c.RedisClient.PublishToRedisChannel(roomUUID, msgBytes)
	return nil
}

func (c *ControlTowerCtrlr) SetupClientConnectionV2(
	conn *websocket.Conn,
	msg *requests.SetClientConnectionEvent) (*requests.SetClientConnectionEvent, error) {

	connectionUUID := uuid.New().String()
	msg.ConnectionUUID = connectionUUID
	userConnection := c.ConnCtrlr.GetConnection(msg.FromUUID)

	if userConnection == nil {
		userConnection = &requests.Connection{
			UserUUID:    msg.FromUUID,
			Connections: map[string]*websocket.Conn{},
		}
		c.ConnCtrlr.AddConnection(userConnection)
	}

	c.ConnCtrlr.AddClient(userConnection, connectionUUID, conn)
	return msg, nil
}

func (c *ControlTowerCtrlr) SaveSeenBy(msg *requests.SeenMessageEvent) error {
	existingMessage, err := c.Repo.GetMessageByUUID(msg.MessageUUID)
	if err != nil {
		return err
	}

	if existingMessage == nil {
		return errors.New("message not found")
	}

	seenBy := &records.SeenBy{
		UserUUID:    msg.UserUUID,
		MessageID:   int(existingMessage.Model.ID),
		MessageUUID: msg.MessageUUID,
	}

	err = c.Repo.SaveSeenBy(seenBy)
	if err != nil {
		return err
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.RedisClient.PublishToRedisChannel(msg.RoomUUID, bytes)
	return nil
}

func (c *ControlTowerCtrlr) GetRoomsByUserUUID(userUUID string, offset int) ([]*requests.Room, error) {
	rooms, err := c.Repo.GetRoomsByUserUUID(userUUID, offset)
	if err != nil {
		return nil, err
	}

	// TODO - put this all in the controller
	requestRooms := make([]*requests.Room, len(rooms))
	for i, room := range rooms {
		members := make([]*requests.Member, len(room.Members))
		messages := make([]*requests.Message, len(room.Messages))

		for j, member := range room.Members {
			members[j] = &requests.Member{
				UserUUID: member.UserUUID,
			}
		}

		for j, msg := range room.Messages {
			messages[j] = &requests.Message{
				UUID:        msg.UUID,
				FromUUID:    msg.FromUUID,
				RoomUUID:    msg.RoomUUID,
				MessageText: msg.MessageText,
			}
		}

		requestRooms[i] = &requests.Room{
			UUID:     room.UUID,
			Members:  members,
			Messages: messages,
		}
	}
	return requestRooms, nil
}

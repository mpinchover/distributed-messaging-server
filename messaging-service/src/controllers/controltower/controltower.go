package controltower

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	mappers "messaging-service/src/mappers/requests"
	redisClient "messaging-service/src/redis"
	"messaging-service/src/repo"
	"messaging-service/src/serrors"
	"messaging-service/src/types/connections"
	"messaging-service/src/types/enums"
	"messaging-service/src/types/records"
	"messaging-service/src/types/requests"

	"github.com/google/uuid"
)

type ControlTowerCtrlr struct {
	Mu          *sync.RWMutex
	RedisClient redisClient.RedisInterface
	Repo        repo.RepoInterface

	UserConnections map[string]*connections.UserConnection // user uuid to list of devices
	Channels        map[string]*connections.Channel        // room uuid to the list of users in the room
}

func New(
	redisClient *redisClient.RedisClient,
	repo *repo.Repo,
) *ControlTowerCtrlr {

	mu := &sync.RWMutex{}
	controlTower := &ControlTowerCtrlr{
		Mu:              mu,
		RedisClient:     redisClient,
		Repo:            repo,
		Channels:        map[string]*connections.Channel{},
		UserConnections: map[string]*connections.UserConnection{},
	}

	return controlTower
}

func (c *ControlTowerCtrlr) GetMessagesByRoomUUID(ctx context.Context, roomUUID string, offset int) ([]*records.Message, error) {
	return c.Repo.GetMessagesByRoomUUID(roomUUID, offset)

}

func (c *ControlTowerCtrlr) CreateRoom(
	ctx context.Context,
	members []*requests.Member,
) (*requests.Room, error) {
	roomUUID := uuid.New().String()

	for _, m := range members {
		m.RoomUUID = roomUUID
	}

	createdAtNano := float64(time.Now().UnixNano()) //  1e6
	room := &records.Room{
		UUID:          roomUUID,
		CreatedAtNano: createdAtNano,
		Members:       mappers.ToRecordMembers(members),
	}

	openRoomEvent := requests.OpenRoomEvent{
		EventType: enums.EVENT_OPEN_ROOM.String(),
		Room:      mappers.ToRequestRoom(room),
	}

	bytes, err := json.Marshal(openRoomEvent)
	if err != nil {
		return nil, err
	}

	err = c.RedisClient.PublishToRedisChannel(enums.CHANNEL_SERVER_EVENTS, bytes)
	if err != nil {
		return nil, err
	}

	// TODO - save in go routine?
	err = c.Repo.SaveRoom(room)
	if err != nil {
		return nil, err
	}

	newRoom := &requests.Room{
		Members:       members,
		UUID:          roomUUID,
		CreatedAtNano: createdAtNano,
	}

	return newRoom, nil
}

func (c *ControlTowerCtrlr) DeleteRoom(ctx context.Context, roomUUID string) error {
	room, err := c.Repo.GetRoomByRoomUUID(roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return serrors.InternalErrorf("room not found", nil)
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

	return c.RedisClient.PublishToRedisChannel(roomUUID, msgBytes)
}

func (c *ControlTowerCtrlr) SetupClientConnectionV2(
	ws *requests.Websocket,
	msg *requests.SetClientConnectionEvent) (*requests.SetClientConnectionEvent, error) {

	deviceUUID := uuid.New().String()
	msg.DeviceUUID = deviceUUID

	userConn := c.GetUserConnection(msg.UserUUID)
	if userConn == nil {
		userConn = &connections.UserConnection{
			UUID:    msg.UserUUID,
			Devices: map[string]*connections.Device{},
		}

		err := c.SetUserConnection(userConn)
		if err != nil {
			return nil, err
		}
	}
	newDeviceConnection := &connections.Device{
		WS:       ws.Conn,
		Outbound: ws.Outbound,
	}

	err := c.SetUserDevice(msg.UserUUID, deviceUUID, newDeviceConnection)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *ControlTowerCtrlr) GetRoomsByUserUUIDForSubscribing(userUUID string) ([]*records.Room, error) {
	rooms, err := c.Repo.GetRoomsByUserUUIDForSubscribing(userUUID)
	return rooms, err
}

func (c *ControlTowerCtrlr) GetRoomsByUserUUID(ctx context.Context, userUUID string, offset int) ([]*records.Room, error) {
	rooms, err := c.Repo.GetRoomsByUserUUID(userUUID, offset)
	if err != nil {
		return nil, err
	}

	// TODO - put this all in the controller
	// requestRooms := make([]*requests.Room, len(rooms))
	// for i, room := range rooms {
	// 	members := make([]*records.Member, len(room.Members))
	// 	messages := make([]*requests.Message, len(room.Messages))

	// 	for j, member := range room.Members {
	// 		members[j] = &records.Member{
	// 			UserUUID: member.UserUUID,
	// 		}
	// 	}

	// 	for j, msg := range room.Messages {
	// 		messages[j] = &records.Message{
	// 			UUID:        msg.UUID,
	// 			FromUUID:    msg.FromUUID,
	// 			RoomUUID:    msg.RoomUUID,
	// 			MessageText: msg.MessageText,
	// 		}
	// 	}

	// 	requestRooms[i] = &records.Room{
	// 		UUID:     room.UUID,
	// 		Members:  members,
	// 		Messages: messages,
	// 	}
	// }
	return rooms, nil
}

// maybe store the rooms each member is part of as memebersOnServer
// remove device from server
// todo, can ou just store the channel in mysql/redis?
func (c *ControlTowerCtrlr) RemoveClientDeviceFromServer(userUUID string, deviceUUID string) error {
	// remove the user from connections

	userConnection := c.GetUserConnection(userUUID)
	if userConnection == nil {
		return nil
	}

	err := c.DeleteDeviceFromServer(userUUID, deviceUUID)
	if err != nil {
		return err
	}

	userConnection = c.GetUserConnection(userUUID)
	if userConnection == nil {
		return nil
	}

	// user has no more devices attached to this connection, delete it
	if len(userConnection.Devices) == 0 {
		err := c.DeleteUserFromServer(userUUID)
		if err != nil {
			return err
		}
	}

	// get the channels for this user, possible optimization
	// rooms, err := c.GetRoomsByUserUUIDForSubscribing(userUUID)
	// if err != nil {
	// 	return err
	// }

	// for _, ch := range rooms {
	// 	_, ok := c.Channels[]
	// }

	// iterate over every channel
	// need to put this in a lock too
	channels := c.GetAllChannelsOnServerForUser(userUUID)
	for _, ch := range channels {

		userIsConnected := c.GetUserConnection(userUUID)

		// user has been deleted; delete user from the channel on this server
		if userIsConnected == nil {
			err := c.DeleteUserFromChannel(userUUID, ch.UUID)
			if err != nil {
				return err
			}
		}

		channel := c.GetChannelFromServer(ch.UUID)
		// if no one else is in channel, unsubscribe and delete channel
		if len(channel.Users) == 0 {
			c.DeleteChannelFromServer(ch.UUID)
			err := ch.Subscriber.Unsubscribe(context.Background())
			if err != nil {
				// TODO - remove the panic
				panic(err)
			}
		}
	}

	return nil
}

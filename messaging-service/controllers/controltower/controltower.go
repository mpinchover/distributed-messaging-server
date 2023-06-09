package controltower

import (
	"encoding/json"
	"errors"
	"fmt"
	redisClient "messaging-service/redis"
	"messaging-service/repo"
	"messaging-service/types/entities"
	"messaging-service/types/eventtypes"
	"messaging-service/types/records"
	"messaging-service/utils"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type MessageController struct {
	RedisClient *redisClient.RedisClient
	Connections map[string]*entities.Connection

	// map the user uuid to a list of the user's connections (different devices)
	UserConnections         map[string]map[string]bool
	OutboundMessagesChannel <-chan *redis.Message

	// track active rooms/channels on this server
	ActiveChannels map[string]*entities.Channel
	Repo           *repo.Repo
}

func New() *MessageController {
	// ctx := context.Background()
	redisClient := redisClient.New()

	repo, err := repo.New()
	if err != nil {
		panic(err)
	}

	connections := map[string]*entities.Connection{}
	userConnections := map[string]map[string]bool{}
	activeChannels := map[string]*entities.Channel{}

	msgController := &MessageController{
		RedisClient:     &redisClient,
		Connections:     connections,
		UserConnections: userConnections,
		ActiveChannels:  activeChannels,
		Repo:            repo,
	}

	serverEventsSubscriber := redisClient.SetupChannel(eventtypes.CHANNEL_SERVER_EVENTS)
	go msgController.subscribeToRedisChannel(serverEventsSubscriber, msgController.handleIncomingServerEventFromRedis)

	return msgController
}

func (c *MessageController) handleIncomingServerEventFromRedis(event string) error {
	eventType, err := getEventType(event)
	if err != nil {
		panic(err)
	}

	if eventType == eventtypes.EVENT_OPEN_ROOM.String() {
		openRoomEvent := entities.OpenRoomEvent{}
		err = json.Unmarshal([]byte(event), &openRoomEvent)
		if err != nil {
			panic(err)
		}

		var listOfFromConnections, listOfToConnections map[string]bool
		fromUUID := utils.ToStr(openRoomEvent.FromUUID)
		toUUID := utils.ToStr(openRoomEvent.ToUUID)

		_listOfFromConnections, ok := c.UserConnections[fromUUID]
		if ok {
			listOfFromConnections = _listOfFromConnections
		}
		_listOfToConnections, ok := c.UserConnections[toUUID]
		if ok {
			listOfToConnections = _listOfToConnections
		}

		roomUUID := utils.ToStr(openRoomEvent.Room.UUID)
		for connUUID := range listOfFromConnections {
			channel, ok := c.ActiveChannels[roomUUID]
			if !ok {
				roomSubscriber := c.RedisClient.SetupChannel(roomUUID)
				go c.subscribeToRedisChannel(roomSubscriber, c.handleIncomingTextMessageFromRedis)

				channel = &entities.Channel{
					Subscriber:           roomSubscriber,
					UUID:                 openRoomEvent.Room.UUID,
					ParticipantsOnServer: map[string]bool{},
				}
				c.ActiveChannels[roomUUID] = channel
			}
			channel.ParticipantsOnServer[toUUID] = true
			c.Connections[connUUID].Conn.WriteJSON(openRoomEvent)
		}

		for connUUID := range listOfToConnections {
			// TODO – use redis client to check if channel is already subscribed
			channel, ok := c.ActiveChannels[roomUUID]
			if !ok {
				roomSubscriber := c.RedisClient.SetupChannel(roomUUID)
				go c.subscribeToRedisChannel(roomSubscriber, c.handleIncomingTextMessageFromRedis)

				channel = &entities.Channel{
					Subscriber: roomSubscriber,
					UUID:       openRoomEvent.Room.UUID,
				}
				c.ActiveChannels[roomUUID] = channel
			}
			channel.ParticipantsOnServer[fromUUID] = true
			c.Connections[connUUID].Conn.WriteJSON(openRoomEvent)
		}

	}

	return nil
}

func (c *MessageController) subscribeToRedisChannel(subscriber *redis.PubSub, fn func(string) error) {
	for redisMsg := range subscriber.Channel() {
		err := fn(redisMsg.Payload)
		if err != nil {
			panic(err)
		}
	}
}

func (c *MessageController) handleIncomingTextMessageFromRedis(msg string) error {
	chatMessage := entities.ChatMessageEvent{}
	err := json.Unmarshal([]byte(msg), &chatMessage)
	if err != nil {
		panic(err)
	}

	roomUUID := utils.ToStr(chatMessage.RoomUUID)
	room, ok := c.ActiveChannels[roomUUID]
	if !ok {
		return nil
	}

	// get all the outbound connections we need to send the message
	outboundConnections := []*entities.Connection{}
	for participantUUID, _ := range room.ParticipantsOnServer {

		connections := c.UserConnections[participantUUID]
		for connUUID := range connections {
			if connUUID != utils.ToStr(chatMessage.FromConnectionUUID) {
				connection, ok := c.Connections[connUUID]
				if !ok {
					continue
				}
				outboundConnections = append(outboundConnections, connection)
			}
		}
	}

	for _, outboundConn := range outboundConnections {
		outboundConn.Conn.WriteJSON(chatMessage)
	}

	return nil
}

func getEventType(event string) (string, error) {
	e := map[string]interface{}{}
	err := json.Unmarshal([]byte(event), &e)
	if err != nil {
		return "", err
	}

	eType, ok := e["eventType"]
	if !ok {
		return "", errors.New("no event type present")
	}
	val, ok := eType.(string)
	if !ok {
		return "", errors.New("could not cast to event type")
	}
	return val, nil
}

func (c *MessageController) SetupClientConnection(conn *websocket.Conn) {

	var userUUID string
	var connectionUUID string
	var mu sync.Mutex
	conn.SetPongHandler(func(appData string) error {
		err := conn.WriteMessage(1, []byte("PONG"))
		if err != nil {
			panic(err)
		}
		return nil
	})

	defer func() {
		conn.Close()
		delete(c.UserConnections[userUUID], connectionUUID)
		if len(c.UserConnections) == 0 {
			delete(c.UserConnections, userUUID)
		}

		// log.Println("HI THERE!!")
		mu.Lock()
		// TODO - move this to a channel
		for roomUUID, channel := range c.ActiveChannels {
			_, ok := channel.ParticipantsOnServer[userUUID]
			if !ok {
				continue
			}

			// if two clients are attached to the same server, they will both try to delete from
			// the same map

			// delete this client from the participants of this room
			delete(channel.ParticipantsOnServer, userUUID)

			// if no one is left on this channel, unsubscribe from it
			if len(channel.ParticipantsOnServer) == 0 {
				err := c.ActiveChannels[roomUUID].Subscriber.Close()
				if err != nil {
					panic(err)
				}
				delete(c.ActiveChannels, roomUUID)
			}
			mu.Unlock()
		}
	}()

	for {
		// read in a message
		_, p, err := conn.ReadMessage()

		if err != nil && websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			break
		}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
				break
			}
		}

		msgType, err := getEventType(string(p))
		if err != nil {
			panic(err)
		}

		if msgType == eventtypes.EVENT_SET_CLIENT_SOCKET.String() {
			// set up the client here and send back a message to the client that everything is ready to go
			// client should be in a loading state until that happens

			msg := entities.SetClientConnectionEvent{}
			err := json.Unmarshal(p, &msg)
			if err != nil {
				panic(err)
			}

			userUUID = utils.ToStr(msg.FromUUID)
			connectionUUID := uuid.New().String()

			connection := &entities.Connection{
				Conn: conn,
				UUID: utils.ToStrPtr(connectionUUID),
			}

			// map the client uuid to a map of connection UUID's to the connection
			_, ok := c.UserConnections[userUUID]
			if !ok {
				c.UserConnections[userUUID] = map[string]bool{}
			}
			c.UserConnections[userUUID][connectionUUID] = true

			c.Connections[connectionUUID] = connection

			msg.ConnectionUUID = utils.ToStrPtr(connectionUUID)

			// send back to client the connection uuid so they can set it
			err = conn.WriteJSON(msg)
			if err != nil {
				panic(err)
			}

		}

		// client has sent out a text message
		if msgType == eventtypes.EVENT_CHAT_TEXT.String() {
			msg := entities.ChatMessageEvent{}
			err := json.Unmarshal(p, &msg)
			if err != nil {
				panic(err)
			}

			chatMessage := &records.ChatMessage{
				FromUUID:    *msg.FromUserUUID,
				MessageText: *msg.MessageText,
				RoomUUID:    *msg.RoomUUID,
				UUID:        uuid.New().String(),
			}

			err = c.Repo.SaveChatMessage(chatMessage)
			if err != nil {
				panic(err)
			}

			roomUUID := utils.ToStr(msg.RoomUUID)
			c.RedisClient.PublishToRedisChannel(roomUUID, p)
		}
	}
	// fmt.Println("CLOSING WEBSOCKET")
}

func (c *MessageController) GetRoomsByUserUUID(userUUID string) ([]*records.ChatRoom, error) {
	return c.Repo.GetHyrdatedRoomsByUserUUID(userUUID)
}

func (c *MessageController) SubscribeRoomsToServer(rooms []*records.ChatRoom, userUUID string) {
	for _, room := range rooms {
		roomUUID := room.UUID
		_, ok := c.ActiveChannels[roomUUID]
		if ok {
			continue
		}
		// if we are not already subscribed to the channel on this server, do so.

		roomSubscriber := c.RedisClient.SetupChannel(roomUUID)
		go c.subscribeToRedisChannel(roomSubscriber, c.handleIncomingTextMessageFromRedis)

		channel := &entities.Channel{
			Subscriber:           roomSubscriber,
			UUID:                 &roomUUID,
			ParticipantsOnServer: map[string]bool{},
		}
		c.ActiveChannels[roomUUID] = channel
		channel.ParticipantsOnServer[userUUID] = true
	}
}
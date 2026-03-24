package handlers

import (
	"log"
	"messaging-service/src/types/enums"
	"messaging-service/src/types/requests"
	"messaging-service/src/utils"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func (h *Handler) SetupWebsocketConnection(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	ws := &requests.Websocket{
		Conn:     conn,
		Outbound: make(chan interface{}),
	}
	// handle breaking the connection
	defer func() {
		conn.Close()
		close(ws.Outbound)
		if ws.UserUUID != nil && ws.DeviceUUID != nil {
			h.ControlTowerCtrlr.RemoveClientDeviceFromServer(*ws.UserUUID, *ws.DeviceUUID)
		}
	}()

	conn.SetPongHandler(func(appData string) error {

		var mu = &sync.RWMutex{}
		mu.Lock()
		err := conn.WriteMessage(1, []byte("PONG"))
		mu.Unlock()
		if err != nil {
			panic(err)
		}
		return nil
	})

	go h.handleOutboundMessages(ws)
	err = h.handleIncomingSocketEvents(ws)
	if err != nil {
		if ws.UserUUID != nil && ws.DeviceUUID != nil {
			// add this to the defer statement?
			h.ControlTowerCtrlr.RemoveClientDeviceFromServer(*ws.UserUUID, *ws.DeviceUUID)
		}
	}
}

func sendClientError(ws *requests.Websocket, err error) error {
	errResp := requests.ErrorResponse{
		Message: err.Error(),
	}
	ws.Outbound <- errResp
	return err
}

func (h *Handler) handleOutboundMessages(ws *requests.Websocket) error {
	// defer close(outbound)
	for msg := range ws.Outbound {
		// TODO - set read write deadline and if they haven't recvd it, remove them from the server
		var mu = &sync.RWMutex{}
		mu.Lock()
		err := ws.Conn.WriteJSON(msg)
		mu.Unlock()
		if err != nil {
			// TODO - remove the panic
			panic(err)
		}

	}
	return nil
}

func (h *Handler) handleIncomingSocketEvents(ws *requests.Websocket) error {

	for {
		// read in a message
		_, p, err := ws.Conn.ReadMessage()
		if err != nil {
			return err
		}

		// TODO – error message for websockets, don't just panic
		msgType, err := utils.GetEventType(string(p))
		if err != nil {
			log.Println(err)
			errResp := requests.ErrorResponse{
				Message: err.Error(),
			}
			ws.Outbound <- errResp
		}

		msgToken, err := utils.GetEventToken(string(p))
		if err != nil {
			sendClientError(ws, err)
		}

		var authErr error
		if msgType == enums.EVENT_SET_CLIENT_SOCKET.String() {
			_, authErr = utils.VerifyJWT(msgToken, true)
		} else {
			_, authErr = utils.VerifyJWT(msgToken, false)
		}

		if authErr != nil {
			log.Println(err)
			return sendClientError(ws, err)
		}

		if msgType == enums.EVENT_SET_CLIENT_SOCKET.String() {
			err := h.handleSetClientSocket(ws, p)
			if err != nil {
				log.Println(err)
				return sendClientError(ws, err)
			}
		}

		if msgType == enums.EVENT_TEXT_MESSAGE.String() {
			err := h.handleClientEventTextMessage(ws.Conn, p)
			if err != nil {
				log.Println(err)
				sendClientError(ws, err)
			}
		}

	}

	return nil
}

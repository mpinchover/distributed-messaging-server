package handlers

import (
	"encoding/json"
	"messaging-service/types/requests"
	"net/http"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()

func (h *Handler) TestHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return nil, nil
}

func (h *Handler) GetRoomsByUserUUID(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	req := &requests.GetRoomsByUserUUIDRequest{}
	err := decoder.Decode(req, r.URL.Query())
	if err != nil {
		return nil, err
	}
	ctx := r.Context()
	return h.getRoomsByUserUUID(ctx, req)
}

func (h *Handler) GetMessagesByRoomUUID(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	req := &requests.GetMessagesByRoomUUIDRequest{}
	err := decoder.Decode(req, r.URL.Query())
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	return h.getMessagesByRoomUUID(ctx, req)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	// run all the middleware here

	req := &requests.CreateRoomRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	ctx := r.Context()
	return h.createRoom(ctx, req)
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// TODO - validation
	req := &requests.DeleteRoomRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	ctx := r.Context()
	return h.deleteRoom(ctx, req)
}

func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	req := &requests.LeaveRoomRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	ctx := r.Context()
	return h.leaveRoom(ctx, req)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	req := &requests.LoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	ctx := r.Context()
	return h.login(ctx, req)
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation
	req := &requests.SignupRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}
	ctx := r.Context()
	return h.signup(ctx, req)
}

func (h *Handler) TestAuthProfileHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	// validation

	req := map[string]interface{}{}

	res, err := h.testAuthProfileHandler(r.Context(), req)
	return res, err
}

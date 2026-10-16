package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v5"
	"htqrcode/caro/api"
	"htqrcode/caro/app"
	"htqrcode/caro/domain"
	"htqrcode/common"
)

type Handler struct {
	service  *app.Service
	clientIP string
}

func NewHandler(service *app.Service) Handler { return Handler{service: service} }
func Register(_ context.Context, e common.EchoRouter, h Handler) error {
	RegisterHandlersWithOptions(e, ipAwareServer{ServerInterface: NewStrictHandler(h, nil), handler: h}, RegisterHandlersOptions{OperationMiddlewares: map[string][]echo.MiddlewareFunc{"createRoom": {validateCreateBody}}})
	return nil
}

func (h Handler) CreateRoom(ctx context.Context, request CreateRoomRequestObject) (CreateRoomResponseObject, error) {
	if request.Body == nil {
		return nil, common.NewInvalidInputError("request_body_required", "request body is required")
	}
	result, err := h.service.Create(ctx, app.CreateInput{DisplayName: request.Body.DisplayName, RoomType: string(request.Body.RoomType), MaxGames: int(request.Body.MaxGames), OwnerIP: h.clientIP, IdempotencyKey: request.Params.IdempotencyKey.String()})
	if err != nil {
		return nil, api.PublicError(err)
	}
	response := CreatedRoom{Room: roomResponse(result.Room)}
	response.HostReservation.PlayerId = result.HostID
	response.HostReservation.HostToken = result.HostToken
	response.HostReservation.Role = PLAYER1
	return CreateRoom201JSONResponse(response), nil
}

func (h Handler) GetRoom(ctx context.Context, request GetRoomRequestObject) (GetRoomResponseObject, error) {
	room, err := h.service.Get(ctx, request.Code)
	if err != nil {
		return nil, api.PublicError(err)
	}
	return GetRoom200JSONResponse(roomResponse(room)), nil
}

func (h Handler) ListRooms(ctx context.Context, _ ListRoomsRequestObject) (ListRoomsResponseObject, error) {
	rooms, err := h.service.List(ctx)
	if err != nil {
		return nil, api.PublicError(err)
	}
	response := ListRooms200JSONResponse{Rooms: make([]Room, 0, len(rooms))}
	for _, r := range rooms {
		response.Rooms = append(response.Rooms, roomResponse(r))
	}
	return response, nil
}

func roomResponse(r domain.Room) Room {
	result := Room{Id: r.ID(), Code: r.Code(), RoomType: RoomRoomType(r.Type()), MaxGames: r.MaxGames(), Status: RoomStatus(r.Status()), CreatedAt: r.CreatedAt(), ExpiresAt: r.ExpiresAt(), Occupancy: len(r.Players()), Full: len(r.Players()) == 2, Joinable: r.Status() == domain.Waiting && len(r.Players()) < 2}
	result.Players = make([]struct {
		Connected   bool   `json:"connected"`
		DisplayName string `json:"displayName"`
		Role        string `json:"role"`
	}, 0, len(r.Players()))
	for _, p := range r.Players() {
		result.Players = append(result.Players, struct {
			Connected   bool   `json:"connected"`
			DisplayName string `json:"displayName"`
			Role        string `json:"role"`
		}{p.Socket() != "", p.Name(), p.Role()})
	}
	return result
}

var _ StrictServerInterface = Handler{}

func validateCreateBody(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if c.Request().Body == nil {
			return common.NewInvalidInputError("invalid_caro_input", "request body is required")
		}
		body, err := io.ReadAll(http.MaxBytesReader(c.Response(), c.Request().Body, 4096))
		if err != nil {
			return common.NewInvalidInputError("invalid_caro_input", "invalid room request body")
		}
		var input *CreateRoom
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil || input == nil {
			return common.NewInvalidInputError("invalid_caro_input", "invalid room request body")
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return common.NewInvalidInputError("invalid_caro_input", "invalid room request body")
		}
		c.Request().Body = io.NopCloser(bytes.NewReader(body))
		c.Response().Header().Set("Cache-Control", "no-store")
		return next(c)
	}
}

// A copy of the handler carries trusted request metadata into the strict interface.
// This keeps IP identity out of the JSON payload and application context values.
type ipAwareServer struct {
	ServerInterface
	handler Handler
}

func (s ipAwareServer) CreateRoom(c *echo.Context, params CreateRoomParams) error {
	h := s.handler
	h.clientIP = c.RealIP()
	return NewStrictHandler(h, nil).CreateRoom(c, params)
}

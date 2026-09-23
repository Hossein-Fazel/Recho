package websocket

import (
	"net/http"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	CtxUserID = "user_id"
)

type WSHandler struct {
	hub             *Hub
	MessageDelivery *application.MessageDelivery
	upgrader        websocket.Upgrader
}

func NewWSHandler(hub *Hub, messageDelivery *application.MessageDelivery) *WSHandler {
	return &WSHandler{
		hub:             hub,
		MessageDelivery: messageDelivery,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,

			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WSHandler) RegsiterRoutes(g *echo.Group) {
	g.GET("/", h.Handler)
}

func (h *WSHandler) Handler(c echo.Context) error {
	userID, ok := c.Get(CtxUserID).(uuid.UUID)
	if !ok {
		return echo.ErrUnauthorized
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := NewClient(
		userID,
		conn,
		h.hub,
		h.MessageDelivery,
	)

	select {
	case h.hub.ClientRegisterChannel() <- client:
	case <-h.hub.ctx.Done():
		_ = conn.Close()
		return nil
	}

	go client.WritePump()
	go client.ReadPump()

	return nil
}

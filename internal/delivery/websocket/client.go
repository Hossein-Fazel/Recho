package websocket

import (
	"encoding/json"
	"time"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub

	MessageDelivery *application.MessageDelivery
}

func NewClient(userID uuid.UUID, conn *websocket.Conn, hub *Hub, msgDelivery *application.MessageDelivery) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,

		MessageDelivery: msgDelivery,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		pkg.Logger.Debug().
			Str("user_id", c.UserID.String()).
			Msg("websocket read pump stopped")

		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			pkg.Logger.Debug().
				Str("user_id", c.UserID.String()).
				Err(err).
				Msg("websocket read failed")

			return
		}

		var income dto.Incomming

		if err := json.Unmarshal(message, &income); err != nil {
			continue
		}

		switch income.Type {
		case string(model.MessageCreateEvent):
			var msg dto.MessageCreateRequest
			_ = json.Unmarshal(income.Payload, &msg)

			c.MessageDelivery.HandleCreateMessage(income.RequestID, model.Message{
				ConversationID: msg.ConversationID,
				SenderID:       c.UserID,
				Type:           msg.Type,
				Text:           (*model.TextMessage)(msg.Text),
			})

		case string(model.MessageEditEvent):
			var msg dto.MessageUpdateRequest
			_ = json.Unmarshal(income.Payload, &msg)

			c.MessageDelivery.HandleEditMessage(income.RequestID, model.Message{
				ID:             msg.MessageID,
				ConversationID: msg.ConversationID,
				SenderID:       c.UserID,
				Type:           msg.Type,
				Text:           (*model.TextMessage)(msg.Text),
			})

		case string(model.MessageDeleteEvent):
			var msg dto.MessageDeleteRequest
			_ = json.Unmarshal(income.Payload, &msg)
			c.MessageDelivery.HandleDeleteMessage(income.RequestID, model.Message{
				ID:             msg.MessageID,
				ConversationID: msg.ConversationID,
				SenderID:       c.UserID,
			})

		default:
			continue
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		pkg.Logger.Debug().
			Str("user_id", c.UserID.String()).
			Msg("websocket write pump stopped")

		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				pkg.Logger.Debug().
					Str("user_id", c.UserID.String()).
					Msg("websocket send channel closed")

				_ = c.Conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)
				return
			}

			if err := c.Conn.WriteMessage(
				websocket.TextMessage,
				message,
			); err != nil {
				pkg.Logger.Debug().
					Str("user_id", c.UserID.String()).
					Err(err).
					Msg("websocket write failed")

				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(
				websocket.PingMessage,
				nil,
			); err != nil {
				pkg.Logger.Debug().
					Str("user_id", c.UserID.String()).
					Err(err).
					Msg("websocket ping failed")

				return
			}
		}
	}
}

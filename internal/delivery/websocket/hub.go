package websocket

import (
	"encoding/json"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
	Deliver    chan application.SendItem
}

var _ application.Sender = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Deliver:    make(chan application.SendItem),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case deliver := <-h.Deliver:
			response := createResponse(deliver)
			for _, reciever := range deliver.Recievers {
				for _, client := range h.Clients[reciever] {
					client.Send <- response
				}
			}
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.Clients[client.UserID] = append(
		h.Clients[client.UserID],
		client,
	)
}

func (h *Hub) unregisterClient(client *Client) {
	clients := h.Clients[client.UserID]

	for i, c := range clients {
		if c == client {
			h.Clients[client.UserID] = append(
				clients[:i],
				clients[i+1:]...,
			)
			break
		}
	}

	if len(h.Clients[client.UserID]) == 0 {
		delete(h.Clients, client.UserID)
	}

	close(client.Send)
}

func (h *Hub) Broadcast(param application.SendItem) {
	h.Deliver <- param
}

func createResponse(v application.SendItem) []byte {
	var response dto.WSResponse
	response.RequestID = v.RequestID
	response.Type = string(v.Event)

	switch v.Event {
	case model.MessageCreateEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.ToMessageCreateResponse(msg)
		}

	case model.MessageEditEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.ToMessageEditResponse(msg)
		}

	case model.MessageDeleteEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.MessageDeleteResponse{
				ID:             msg.ID,
				ConversationID: msg.ConversationID,
			}
		}

	case model.ConversationDeleteEvent:
		if id, ok := v.Content.(uuid.UUID); ok {
			response.Data = dto.ConversationDeleteResponse{
				ConversationID: id,
			}
		}
	}

	res, _ := json.Marshal(response)
	return res
}

func toMessage(content any) (model.Message, bool) {
	switch msg := content.(type) {
	case model.Message:
		return msg, true
	case *model.Message:
		if msg == nil {
			return model.Message{}, false
		}
		return *msg, true
	default:
		return model.Message{}, false
	}
}

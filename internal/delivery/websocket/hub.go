package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
	Deliver    chan application.SendItems
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Deliver:    make(chan application.SendItems),
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
					fmt.Print("message sent to user ", client.UserID, deliver)
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

func (h *Hub) Send(params application.SendItems) {
	h.Deliver <- params
}

func createResponse(v application.SendItems) []byte {
	var response dto.WSResponse
	response.RequestID = v.RequestID
	response.Type = string(v.Event)

	switch v.Event {
	case model.MessageCreateEvent:
		if msg, ok := v.Content.(model.Message); ok {
			response.Data = dto.ToMessageCreatedResponse(msg)
		}
	}

	res, _ := json.Marshal(response)
	return res
}

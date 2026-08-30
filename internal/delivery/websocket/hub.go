package websocket

import (
	"encoding/json"

	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

type deliverModel struct {
	content   any
	receivers uuid.UUIDs
}

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
	Deliver    chan deliverModel
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Deliver:    make(chan deliverModel),
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
			response := createResponse(deliver.content)
			for _, receiver := range deliver.receivers {
				for _, client := range h.Clients[receiver] {
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

func (h *Hub) Send(v any, receivers uuid.UUIDs) {
	h.Deliver <- deliverModel{
		content:   v,
		receivers: receivers,
	}
}

func createResponse(v any) []byte {
	var response dto.WSResponse
	if msg, ok := v.(model.Message); ok {
		response.Type = "message.created"
		response.Data = dto.ToMessageCreatedResponse(msg)
	}

	res, _ := json.Marshal(response)
	return res
}

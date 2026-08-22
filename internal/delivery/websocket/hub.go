package websocket

import "github.com/google/uuid"

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)
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

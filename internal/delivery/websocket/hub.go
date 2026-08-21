package websocket

import "github.com/google/uuid"

type Hub struct {
	clients    map[uuid.UUID][]*Client
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client.UserID] = append(
		h.clients[client.UserID],
		client,
	)
}

func (h *Hub) unregisterClient(client *Client) {
	clients := h.clients[client.UserID]

	for i, c := range clients {
		if c == client {
			h.clients[client.UserID] = append(
				clients[:i],
				clients[i+1:]...,
			)
			break
		}
	}

	if len(h.clients[client.UserID]) == 0 {
		delete(h.clients, client.UserID)
	}

	close(client.Send)
}

package websocket

import (
	"sync"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
	Deliver    chan application.SendItem
	stop       chan struct{}
	stopOnce   sync.Once
}

var _ application.Sender = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Deliver:    make(chan application.SendItem),
		stop:       make(chan struct{}),
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

			for _, receiver := range deliver.Recievers {
				for _, client := range h.Clients[receiver] {
					select {
					case client.Send <- response:
					case <-h.stop:
						return
					}
				}
			}

		case <-h.stop:
			h.closeAllClients()
			return
		}
	}
}

func (h *Hub) Stop() {
	h.stopOnce.Do(func() {
		close(h.stop)
	})
}

func (h *Hub) closeAllClients() {
	for userID, clients := range h.Clients {
		for _, client := range clients {
			client.Close()
		}

		delete(h.Clients, userID)
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

	client.Close()
}

func (h *Hub) Broadcast(param application.SendItem) {
	select {
	case h.Deliver <- param:
	case <-h.stop:
	}
}
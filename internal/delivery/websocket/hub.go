package websocket

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/google/uuid"
)

type Hub struct {
	Clients    map[uuid.UUID][]*Client
	Register   chan *Client
	Unregister chan *Client
	Deliver    <-chan application.SendItem

	ctx  context.Context
	done chan struct{}
}

func NewHub(ctx context.Context, deliver <-chan application.SendItem) *Hub {
	return &Hub{
		Clients:    make(map[uuid.UUID][]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Deliver:    deliver,
		ctx:        ctx,
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	defer close(h.done)

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case deliver := <-h.Deliver:
			if stopped := h.broadcast(deliver); stopped {
				h.closeAllClients()
				return
			}

		case <-h.ctx.Done():
			h.closeAllClients()
			return
		}
	}
}

func (h *Hub) Shutdown(ctx context.Context) error {
	select {
	case <-h.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) broadcast(deliver application.SendItem) (stopped bool) {
	response := createResponse(deliver)

	for _, receiver := range deliver.Recievers {
		for _, client := range h.Clients[receiver] {
			select {
			case client.Send <- response:
			case <-h.ctx.Done():
				return true
			}
		}
	}

	return false
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

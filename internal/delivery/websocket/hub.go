package websocket

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/google/uuid"
)

type Hub struct {
	clients    map[uuid.UUID][]*Client
	register   chan *Client
	unregister chan *Client
	deliver    <-chan application.SendItem

	presenceServie *application.PresenceService

	ctx  context.Context
	done chan struct{}
}

func NewHub(ctx context.Context, deliver <-chan application.SendItem, presenceServie *application.PresenceService) *Hub {
	return &Hub{
		clients:        make(map[uuid.UUID][]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		deliver:        deliver,
		presenceServie: presenceServie,
		ctx:            ctx,
		done:           make(chan struct{}),
	}
}

func (h *Hub) ClientRegisterChannel() chan<- *Client {
	return h.register
}

func (h *Hub) ClientUnregisterChannel() chan<- *Client {
	return h.unregister
}

func (h *Hub) Run() {
	defer close(h.done)

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case deliver := <-h.deliver:
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
		for _, client := range h.clients[receiver] {
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
	for userID, clients := range h.clients {
		for _, client := range clients {
			client.Close()
		}

		delete(h.clients, userID)
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client.UserID] = append(
		h.clients[client.UserID],
		client,
	)
	h.presenceServie.Online(client.UserID)
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
		h.presenceServie.Offline(client.UserID)
	}

	client.Close()
}

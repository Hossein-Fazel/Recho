package memory

import (
	"sync"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/google/uuid"
)

type PresenceRepo struct {
	mu sync.RWMutex

	// target user -> subscriber user IDs
	watchers map[uuid.UUID]map[uuid.UUID]struct{}

	// subscriber user ID -> target user IDs
	subscriptions map[uuid.UUID]map[uuid.UUID]struct{}

	// users that currently have at least one active connection
	online map[uuid.UUID]struct{}
}

var _ application.PresenceRepository = (*PresenceRepo)(nil)

func NewPresenceRepo() *PresenceRepo {
	return &PresenceRepo{
		watchers:      make(map[uuid.UUID]map[uuid.UUID]struct{}),
		subscriptions: make(map[uuid.UUID]map[uuid.UUID]struct{}),
		online:        make(map[uuid.UUID]struct{}),
	}
}

func (p *PresenceRepo) Subscribe(subscriberID, targetUserID uuid.UUID) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if subscriberID == targetUserID {
		return false
	}

	if p.watchers[targetUserID] == nil {
		p.watchers[targetUserID] = make(map[uuid.UUID]struct{})
	}

	if p.subscriptions[subscriberID] == nil {
		p.subscriptions[subscriberID] = make(map[uuid.UUID]struct{})
	}

	if _, exists := p.subscriptions[subscriberID][targetUserID]; exists {
		_, online := p.online[targetUserID]
		return online
	}

	p.watchers[targetUserID][subscriberID] = struct{}{}
	p.subscriptions[subscriberID][targetUserID] = struct{}{}

	_, online := p.online[targetUserID]

	return online
}

func (p *PresenceRepo) Unsubscribe(subscriberID, targetUserID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if targets, ok := p.subscriptions[subscriberID]; ok {
		delete(targets, targetUserID)

		if len(targets) == 0 {
			delete(p.subscriptions, subscriberID)
		}
	}

	if subscribers, ok := p.watchers[targetUserID]; ok {
		delete(subscribers, subscriberID)

		if len(subscribers) == 0 {
			delete(p.watchers, targetUserID)
		}
	}
}

func (p *PresenceRepo) RemoveSubscriber(subscriberID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	targets, ok := p.subscriptions[subscriberID]
	if !ok {
		return
	}

	for targetUserID := range targets {
		if subscribers, ok := p.watchers[targetUserID]; ok {
			delete(subscribers, subscriberID)

			if len(subscribers) == 0 {
				delete(p.watchers, targetUserID)
			}
		}
	}

	delete(p.subscriptions, subscriberID)
}

func (p *PresenceRepo) SetOnline(userID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Already online.
	if _, exists := p.online[userID]; exists {
		return
	}

	p.online[userID] = struct{}{}
}

func (p *PresenceRepo) SetOffline(userID uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Already offline.
	if _, exists := p.online[userID]; !exists {
		return
	}

	delete(p.online, userID)
}

func (p *PresenceRepo) GetSubscribers(userID uuid.UUID) uuid.UUIDs {
	p.mu.Lock()
	defer p.mu.Unlock()

	subscribers := make([]uuid.UUID, 0, len(p.watchers[userID]))

	for subscriberID := range p.watchers[userID] {
		subscribers = append(subscribers, subscriberID)
	}

	return subscribers
}

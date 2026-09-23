package application

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/apperr"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/Hossein-Fazel/Recho/pkg"
	"github.com/google/uuid"
)

type PresenceService struct {
	presenceRepo PresenceRepo
	userRepo     UserRepo
	sender       Sender
}

func NewPresenceService(presenceRepo PresenceRepo, userRepo UserRepo, sender Sender) *PresenceService {
	pkg.Logger.Info().Msg("Initializing presence service")

	return &PresenceService{
		presenceRepo: presenceRepo,
		userRepo:     userRepo,
		sender:       sender,
	}
}

func (p *PresenceService) Online(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return apperr.InvalidInput("presence service", "invalid user id", nil)
	}
	p.presenceRepo.SetOnline(userID)
	subscribers := p.presenceRepo.GetSubscribers(userID)
	if len(subscribers) > 0 {
		p.sender.Broadcast(SendItem{
			RequestID: uuid.New(),
			Event:     model.UserOnline,
			Content: model.PresenceUpdate{
				UserID: userID,
				Online: true,
			},
			Recievers: subscribers,
		})
	}

	return nil
}

func (p *PresenceService) Offline(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return apperr.InvalidInput("presence service", "invalid user id", nil)
	}
	p.presenceRepo.SetOffline(userID)

	lastSeen, err := p.userRepo.UpdateLastSeen(context.Background(), userID)
	if err != nil {
		return err
	}

	subscribers := p.presenceRepo.GetSubscribers(userID)
	if len(subscribers) > 0 {
		p.sender.Broadcast(SendItem{
			RequestID: uuid.New(),
			Event:     model.UserOffline,
			Content: model.PresenceUpdate{
				UserID:   userID,
				Online:   false,
				LastSeen: &lastSeen,
			},
			Recievers: subscribers,
		})
	}
	p.presenceRepo.RemoveSubscriber(userID)

	return nil
}

func (p *PresenceService) Subscribe(subscriberID uuid.UUID, targetUserIDs uuid.UUIDs) error {
	if subscriberID == uuid.Nil {
		return apperr.InvalidInput("presence service", "user id is required", nil)
	}

	for _, targetUserID := range targetUserIDs {
		if subscriberID == targetUserID {
			return apperr.Conflict("presence service", "subscriber id and target userid should be different", nil)
		}
		p.presenceRepo.Subscribe(subscriberID, targetUserID)
	}

	return nil
}

func (p *PresenceService) Unsubscribe(subscriberID uuid.UUID, targetUserIDs uuid.UUIDs) error {
	if subscriberID == uuid.Nil {
		return apperr.InvalidInput("presence service", "user id is required", nil)
	}

	for _, targetUserID := range targetUserIDs {
		if subscriberID == targetUserID {
			return apperr.Conflict("presence service", "subscriber id and target userid should be different", nil)
		}

		p.presenceRepo.Unsubscribe(subscriberID, targetUserID)
	}

	return nil
}

func (p *PresenceService) GetStatuses(targetUserIDs uuid.UUIDs) ([]model.UserStatus, error) {
	if len(targetUserIDs) == 0 {
		return []model.UserStatus{}, apperr.InvalidInput("presence service", "user ids required", nil)
	}
	statuses := make([]model.UserStatus, 0, len(targetUserIDs))

	for _, target := range targetUserIDs {
		statuses = append(statuses, model.UserStatus{
			ID:     target,
			Online: p.presenceRepo.IsOnline(target),
		})
	}

	return statuses, nil
}

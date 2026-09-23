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
}

func NewPresenceService(presenceRepo PresenceRepo, userRepo UserRepo) *PresenceService {
	pkg.Logger.Info().Msg("Initializing presence service")

	return &PresenceService{
		presenceRepo: presenceRepo,
		userRepo:     userRepo,
	}
}

func (p *PresenceService) Online(userID uuid.UUID) (*SendItem, error) {
	if userID == uuid.Nil {
		return nil, apperr.InvalidInput("presence service", "invalid user id", nil)
	}
	p.presenceRepo.SetOnline(userID)

	subscribers := p.presenceRepo.GetSubscribers(userID)
	if len(subscribers) == 0 {
		return nil, nil
	}

	return &SendItem{
		RequestID: uuid.New(),
		Event:     model.UserOnline,
		Content: model.UserPresence{
			UserID: userID,
			Online: true,
		},
		Recievers: subscribers,
	}, nil
}

func (p *PresenceService) Offline(userID uuid.UUID) (*SendItem, error) {
	if userID == uuid.Nil {
		return nil, apperr.InvalidInput("presence service", "invalid user id", nil)
	}
	p.presenceRepo.SetOffline(userID)

	lastSeen, err := p.userRepo.UpdateLastSeen(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	subscribers := p.presenceRepo.GetSubscribers(userID)
	p.presenceRepo.RemoveSubscriber(userID)

	if len(subscribers) == 0 {
		return nil, nil
	}

	return &SendItem{
		RequestID: uuid.New(),
		Event:     model.UserOffline,
		Content: model.UserPresence{
			UserID:   userID,
			Online:   false,
			LastSeen: &lastSeen,
		},
		Recievers: subscribers,
	}, nil
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

func (p *PresenceService) GetStatuses(targetUserIDs uuid.UUIDs) ([]model.UserPresence, error) {
	if len(targetUserIDs) == 0 {
		return []model.UserPresence{}, apperr.InvalidInput("presence service", "user ids required", nil)
	}
	statuses := make([]model.UserPresence, 0, len(targetUserIDs))

	for _, target := range targetUserIDs {
		statuses = append(statuses, model.UserPresence{
			UserID: target,
			Online: p.presenceRepo.IsOnline(target),
		})
	}

	return statuses, nil
}

package application

import (
	"context"

	"github.com/Hossein-Fazel/Recho/internal/model"
)

type MessageDelivery struct {
	msgService  *MessageService
	convService *ConverasionService
	Sender      Sender
}

func NewMessageDelivery(msgService *MessageService) *MessageDelivery {
	return &MessageDelivery{
		msgService: msgService,
	}
}

func (d *MessageDelivery) HandleMessage(msg model.Message) error {
	ctx := context.Background()
	message, err := d.msgService.Create(ctx, msg)
	if err != nil {
		return err
	}

	userIDs, err := d.convService.GetConversationUsers(ctx, msg.SenderID, msg.ConversationID)

	d.Sender.Send(*message, userIDs)
	return nil
}

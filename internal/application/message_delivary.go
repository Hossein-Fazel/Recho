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

func NewMessageDelivery(msgService *MessageService, convService *ConverasionService, sender Sender) *MessageDelivery {
	return &MessageDelivery{
		msgService:  msgService,
		convService: convService,
		Sender:      sender,
	}
}

func (d *MessageDelivery) HandleMessage(msg model.Message) error {
	ctx := context.Background()
	message, err := d.msgService.Create(ctx, msg)
	if err != nil {
		return err
	}

	userIDs, err := d.convService.GetConversationUsers(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return err
	}

	d.Sender.Send(*message, userIDs)
	return nil
}

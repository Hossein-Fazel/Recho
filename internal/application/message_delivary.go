package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/Hossein-Fazel/Recho/internal/model"
)

type MessageDelivery struct {
	msgService  *MessageService
	convService *ConversationService
	Sender      Sender
}

func NewMessageDelivery(msgService *MessageService, convService *ConversationService, sender Sender) *MessageDelivery {
	return &MessageDelivery{
		msgService:  msgService,
		convService: convService,
		Sender:      sender,
	}
}

func (d *MessageDelivery) HandleCreateMessage(reqID uuid.UUID, msg model.Message) {
	ctx := context.Background()
	message, err := d.msgService.Create(ctx, msg)
	if err != nil {
		return
	}

	userIDs, err := d.convService.GetConversationUserIDs(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return
	}

	d.Sender.Send(SendItems{
		RequestID: reqID,
		Event:     model.MessageCreateEvent,
		Content:   message,
		Recievers: userIDs,
	})
}

func (d *MessageDelivery) HandleEditMessage(reqID uuid.UUID, msg model.Message) {
	ctx := context.Background()
	message, err := d.msgService.Update(ctx, msg)
	if err != nil {
		return
	}

	userIDs, err := d.convService.GetConversationUserIDs(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return
	}

	d.Sender.Send(SendItems{
		RequestID: reqID,
		Event:     model.MessageEditEvent,
		Content:   message,
		Recievers: userIDs,
	})
}

func (d *MessageDelivery) HandleDeleteMessage(reqID uuid.UUID, msg model.Message) {
	ctx := context.Background()
	err := d.msgService.Delete(ctx, msg)
	if err != nil {
		return
	}

	userIDs, err := d.convService.GetConversationUserIDs(ctx, msg.SenderID, msg.ConversationID)
	if err != nil {
		return
	}

	d.Sender.Send(SendItems{
		RequestID: reqID,
		Event:     model.MessageDeleteEvent,
		Content:   msg,
		Recievers: userIDs,
	})
}

package websocket

import (
	"encoding/json"

	"github.com/Hossein-Fazel/Recho/internal/application"
	"github.com/Hossein-Fazel/Recho/internal/delivery/websocket/dto"
	"github.com/Hossein-Fazel/Recho/internal/model"
	"github.com/google/uuid"
)

func createResponse(v application.SendItem) []byte {
	var response dto.WSResponse
	response.RequestID = v.RequestID
	response.Type = string(v.Event)

	switch v.Event {
	case model.MessageCreateEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.ToMessageCreateResponse(msg)
		}

	case model.MessageEditEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.ToMessageEditResponse(msg)
		}

	case model.MessageDeleteEvent:
		if msg, ok := toMessage(v.Content); ok {
			response.Data = dto.MessageDeleteResponse{
				ID:             msg.ID,
				ConversationID: msg.ConversationID,
			}
		}

	case model.ConversationDeleteEvent:
		if id, ok := v.Content.(uuid.UUID); ok {
			response.Data = dto.ConversationDeleteResponse{
				ConversationID: id,
			}
		}

	case model.GroupUpdateEvent:
		if group, ok := toGroup(v.Content); ok {
			response.Data = dto.GroupUpdateResponse{
				GroupID:   group.ID,
				Name:      group.Name,
				AvatarURL: group.AvatarKey,
				Bio:       group.Bio,
			}
		}
	}

	res, _ := json.Marshal(response)
	return res
}

func toGroup(content any) (model.Group, bool) {
	switch group := content.(type) {
	case model.Group:
		return group, true
	case *model.Group:
		if group == nil {
			return model.Group{}, false
		}
		return *group, true
	default:
		return model.Group{}, false
	}
}

func toMessage(content any) (model.Message, bool) {
	switch msg := content.(type) {
	case model.Message:
		return msg, true
	case *model.Message:
		if msg == nil {
			return model.Message{}, false
		}
		return *msg, true
	default:
		return model.Message{}, false
	}
}

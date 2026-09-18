package model

type Event string

const (
	MessageCreateEvent      Event = "message.create"
	MessageEditEvent        Event = "message.edit"
	MessageDeleteEvent      Event = "message.delete"
	ConversationDeleteEvent Event = "conversation.delete"
	GroupUpdateEvent        Event = "group.update"
)

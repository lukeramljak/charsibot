package server

type EventType string

const (
	EventTypeChatCommand        EventType = "chat_command"
	EventTypeCollectionDisplay  EventType = "blindbox_display"
	EventTypeBlindBoxRedemption EventType = "blindbox_redemption"
)

type OverlayEvent struct {
	Type EventType
	Data any
}

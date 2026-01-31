package server

type EventType string

const (
	EventTypeChatCommand        EventType = "chat_command"
	EventTypeRedemption         EventType = "redemption"
	EventTypeCollectionDisplay  EventType = "collection_display"
	EventTypeBlindBoxRedemption EventType = "blindbox_redemption"
	EventTypeConnected          EventType = "connected"
)

type OverlayEvent struct {
	Type      EventType `json:"type"`
	Message   string    `json:"message,omitempty"`
	Data      any       `json:"data,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
}

package bot

type EventType string

const (
	EventTypeChatCommand        EventType = "chat_command"
	EventTypeRedemption         EventType = "redemption"
	EventTypeCollectionDisplay  EventType = "collection_display"
	EventTypeBlindBoxRedemption EventType = "blindbox_redemption"
	EventTypeLeaderboard        EventType = "leaderboard"
	EventTypeConnected          EventType = "connected"
)

type OverlayEvent struct {
	Type      EventType `json:"type"`
	Message   string    `json:"message,omitempty"`
	Data      any       `json:"data,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
}

// Broadcaster is implemented by any type that can broadcast overlay events.
type Broadcaster interface {
	Broadcast(event OverlayEvent)
}

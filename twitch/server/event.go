package server

type EventType string

const (
	EventTypeChatCommand        EventType = "chat_command"
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

type CollectionDisplayData struct {
	UserID         string   `json:"userId"`
	Username       string   `json:"username"`
	Series         string   `json:"series"`
	Collection     []string `json:"collection"`
	CollectionSize int      `json:"collectionSize"`
}

type BlindBoxRedemptionData struct {
	UserID         string   `json:"userId"`
	Username       string   `json:"username"`
	Series         string   `json:"series"`
	SeriesName     string   `json:"seriesName"`
	Plushie        string   `json:"plushie"`
	IsNew          bool     `json:"isNew"`
	Collection     []string `json:"collection"`
	CollectionSize int      `json:"collectionSize"`
}

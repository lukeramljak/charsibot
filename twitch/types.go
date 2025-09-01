package main

type WebSocketMessage struct {
	Metadata WebSocketMetadata `json:"metadata"`
	Payload  WebSocketPayload  `json:"payload"`
}

type WebSocketMetadata struct {
	MessageID           string `json:"message_id"`
	MessageType         string `json:"message_type"`
	MessageTimestamp    string `json:"message_timestamp"`
	SubscriptionType    string `json:"subscription_type"`
	SubscriptionVersion string `json:"subscription_version"`
}

type WebSocketPayload struct {
	Session      *Session      `json:"session,omitempty"`
	Subscription *Subscription `json:"subscription,omitempty"`
	Event        *Event        `json:"event,omitempty"`
}

type Session struct {
	ID                      string  `json:"id"`
	Status                  string  `json:"status"`
	ConnectedAt             string  `json:"connected_at"`
	KeepaliveTimeoutSeconds *int    `json:"keepalive_timeout_seconds"`
	ReconnectURL            *string `json:"reconnect_url"`
}

type Subscription struct {
	ID        string         `json:"id"`
	Status    string         `json:"status"`
	Type      string         `json:"type"`
	Version   string         `json:"version"`
	Cost      int            `json:"cost"`
	Condition map[string]any `json:"condition"`
	Transport Transport      `json:"transport"`
	CreatedAt string         `json:"created_at"`
}

type Transport struct {
	Method    string `json:"method"`
	SessionID string `json:"session_id"`
}

type Event struct {
	// Common broadcaster fields
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`

	// Chat message fields
	ChatterUserID               string  `json:"chatter_user_id"`
	ChatterUserName             string  `json:"chatter_user_name"`
	ChatterUserLogin            string  `json:"chatter_user_login"`
	MessageID                   string  `json:"message_id"`
	Message                     Message `json:"message"`
	Color                       string  `json:"color"`
	Badges                      []Badge `json:"badges"`
	MessageType                 string  `json:"message_type"`
	Cheer                       *Cheer  `json:"cheer,omitempty"`
	Reply                       *Reply  `json:"reply,omitempty"`
	ChannelPointsCustomRewardID *string `json:"channel_points_custom_reward_id,omitempty"`

	// Channel points redemption fields
	ID         string               `json:"id"`
	UserID     string               `json:"user_id"`
	UserLogin  string               `json:"user_login"`
	UserName   string               `json:"user_name"`
	UserInput  string               `json:"user_input"`
	Status     string               `json:"status"`
	Reward     *ChannelPointsReward `json:"reward"`
	RedeemedAt string               `json:"redeemed_at"`
}

type Message struct {
	Text      string     `json:"text"`
	Fragments []Fragment `json:"fragments"`
}

type Fragment struct {
	Type      string     `json:"type"`
	Text      string     `json:"text"`
	Cheermote *Cheermote `json:"cheermote,omitempty"`
	Emote     *Emote     `json:"emote,omitempty"`
	Mention   *Mention   `json:"mention,omitempty"`
}

type Badge struct {
	SetID string `json:"set_id"`
	ID    string `json:"id"`
	Info  string `json:"info"`
}

type Cheer struct {
	Bits int `json:"bits"`
}

type Reply struct {
	ParentMessageID   string `json:"parent_message_id"`
	ParentMessageBody string `json:"parent_message_body"`
	ParentUserID      string `json:"parent_user_id"`
	ParentUserName    string `json:"parent_user_name"`
	ParentUserLogin   string `json:"parent_user_login"`
	ThreadMessageID   string `json:"thread_message_id"`
	ThreadUserID      string `json:"thread_user_id"`
	ThreadUserName    string `json:"thread_user_name"`
	ThreadUserLogin   string `json:"thread_user_login"`
}

type Cheermote struct {
	Prefix string `json:"prefix"`
	Bits   int    `json:"bits"`
	Tier   int    `json:"tier"`
}

type Emote struct {
	ID         string   `json:"id"`
	EmoteSetID string   `json:"emote_set_id"`
	OwnerID    string   `json:"owner_id"`
	Format     []string `json:"format"`
}

type Mention struct {
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	UserLogin string `json:"user_login"`
}

type ChannelPointsReward struct {
	ID                                string                           `json:"id"`
	BroadcasterID                     string                           `json:"broadcaster_id"`
	BroadcasterLogin                  string                           `json:"broadcaster_login"`
	BroadcasterName                   string                           `json:"broadcaster_name"`
	Title                             string                           `json:"title"`
	Prompt                            string                           `json:"prompt"`
	Cost                              int                              `json:"cost"`
	Image                             *ChannelPointsImage              `json:"image"`
	DefaultImage                      ChannelPointsImage               `json:"default_image"`
	BackgroundColor                   string                           `json:"background_color"`
	IsEnabled                         bool                             `json:"is_enabled"`
	IsUserInputRequired               bool                             `json:"is_user_input_required"`
	MaxPerStream                      ChannelPointsMaxPerStream        `json:"max_per_stream"`
	MaxPerUserPerStream               ChannelPointsMaxPerUserPerStream `json:"max_per_user_per_stream"`
	IsInStock                         bool                             `json:"is_in_stock"`
	ShouldRedemptionsSkipRequestQueue bool                             `json:"should_redemptions_skip_request_queue"`
	RedemptionsRedeemedCurrentStream  *int                             `json:"redemptions_redeemed_current_stream"`
	CooldownExpiresAt                 *string                          `json:"cooldown_expires_at"`
}

type ChannelPointsImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

type ChannelPointsMaxPerStream struct {
	IsEnabled    bool `json:"is_enabled"`
	MaxPerStream int  `json:"max_per_stream"`
}

type ChannelPointsMaxPerUserPerStream struct {
	IsEnabled           bool `json:"is_enabled"`
	MaxPerUserPerStream int  `json:"max_per_user_per_stream"`
}

type Stats struct {
	Username     string
	Strength     int
	Intelligence int
	Charisma     int
	Luck         int
	Dexterity    int
	Penis        int
}

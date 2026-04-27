package blindbox

import "github.com/lukeramljak/charsibot/twitch/db"

// BlindBoxDisplayData is the payload for a blindbox_display SSE event.
type BlindBoxDisplayData struct {
	Username   string       `json:"username"`
	Collection []string     `json:"collection"`
	Config     SeriesConfig `json:"config"`
}

// BlindBoxRedemptionData is the payload for a blindbox_redemption SSE event.
type BlindBoxRedemptionData struct {
	Username   string             `json:"username"`
	Plushie    db.BlindBoxPlushie `json:"plushie"`
	IsNew      bool               `json:"isNew"`
	Collection []string           `json:"collection"`
	Config     SeriesConfig       `json:"config"`
}

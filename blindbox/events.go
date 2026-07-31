package blindbox

// BlindBoxDisplayData is the payload for a blindbox_display SSE event.
type BlindBoxDisplayData struct {
	Username   string       `json:"username"`
	Collection []string     `json:"collection" nullable:"false"`
	Config     SeriesConfig `json:"config"`
}

// BlindBoxRedemptionData is the payload for a blindbox_redemption SSE event.
type BlindBoxRedemptionData struct {
	Username   string       `json:"username"`
	Plushie    Plushie      `json:"plushie"`
	IsNew      bool         `json:"isNew"`
	Collection []string     `json:"collection" nullable:"false"`
	Config     SeriesConfig `json:"config"`
}

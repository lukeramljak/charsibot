package blindbox

import (
	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
)

// BlindBoxRedemption handles channel point redemptions for blind boxes
type BlindBoxRedemption struct {
	config BlindBoxConfig
}

func NewBlindBoxRedemption(config BlindBoxConfig) *BlindBoxRedemption {
	return &BlindBoxRedemption{config: config}
}

func (r *BlindBoxRedemption) ShouldTrigger(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool {
	return event.Reward.Title == r.config.RewardTitle
}

func (r *BlindBoxRedemption) Execute(b *bot.Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	RedeemBlindBox(b, event.UserID, event.UserName, r.config)
}

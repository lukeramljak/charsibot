package bot

import (
	"github.com/joeyak/go-twitch-eventsub/v3"
)

type Redemption interface {
	ShouldTrigger(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool
	Execute(bot *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd)
}

type RedemptionHandler struct {
	Redemptions []Redemption
}

func NewRedemptionHandler(redemptions []Redemption) *RedemptionHandler {
	return &RedemptionHandler{
		Redemptions: redemptions,
	}
}

func (h *RedemptionHandler) Process(bot *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	for _, redemption := range h.Redemptions {
		if !redemption.ShouldTrigger(event) {
			continue
		}

		bot.logger.Info("executing redemption",
			"user", event.UserName,
			"reward", event.Reward.Title,
		)
		redemption.Execute(bot, event)
	}
}

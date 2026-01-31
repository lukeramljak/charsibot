package bot

import (
	"math/rand/v2"

	"github.com/joeyak/go-twitch-eventsub/v3"
)

type Trigger interface {
	TriggerChance() int
	ShouldTrigger(event twitch.EventChannelChatMessage) bool
	Execute(bot *Bot, event twitch.EventChannelChatMessage)
}

type TriggerHandler struct {
	Triggers []Trigger
}

func NewTriggerHandler(triggers []Trigger) *TriggerHandler {
	return &TriggerHandler{
		Triggers: triggers,
	}
}

func (h *TriggerHandler) Process(bot *Bot, event twitch.EventChannelChatMessage) {
	for _, trigger := range h.Triggers {
		if !trigger.ShouldTrigger(event) {
			continue
		}

		if chance := trigger.TriggerChance(); chance > 0 && chance < 100 {
			roll := rand.IntN(100) + 1
			if roll > chance {
				bot.logger.Debug("trigger failed chance roll",
					"roll", roll,
					"chance", chance,
				)
				continue
			}
		}

		bot.logger.Info("executing trigger",
			"user", event.ChatterUserName,
			"message", event.Message.Text,
		)
		trigger.Execute(bot, event)
	}
}

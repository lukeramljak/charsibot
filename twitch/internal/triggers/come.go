package triggers

import (
	"slices"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
)

type ComeTrigger struct {
	triggers []string
}

func NewComeTrigger() *ComeTrigger {
	return &ComeTrigger{
		triggers: []string{"come", "coming", "cum", "came"},
	}
}

func (t *ComeTrigger) TriggerChance() int {
	return 20
}

func (t *ComeTrigger) ShouldTrigger(event twitch.EventChannelChatMessage) bool {
	// Split by non-word characters and filter empty strings
	words := strings.FieldsFunc(strings.ToLower(event.Message.Text), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	for _, word := range words {
		if slices.Contains(t.triggers, word) {
			return true
		}
	}
	return false
}

func (t *ComeTrigger) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	b.SendMessage(bot.SendMessageParams{
		Message:              "no coming",
		ReplyParentMessageID: event.MessageId,
	})
}

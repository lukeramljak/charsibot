package bot

import (
	"slices"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
)

type Trigger struct {
	Chance        int
	ShouldTrigger func(event twitch.EventChannelChatMessage) bool
	Execute       func(b *Bot, event twitch.EventChannelChatMessage)
}

// Triggers returns all chat message triggers.
func Triggers() []Trigger {
	const comeTriggerChance = 20

	comeTriggerWords := []string{"come", "coming", "cum", "came"}

	return []Trigger{
		{
			Chance: comeTriggerChance,
			ShouldTrigger: func(event twitch.EventChannelChatMessage) bool {
				if strings.ToLower(event.Message.Text) == "no coming" {
					return false
				}

				words := strings.FieldsFunc(strings.ToLower(event.Message.Text), func(r rune) bool {
					return (r < 'a' || r > 'z') && (r < '0' || r > '9')
				})

				for _, word := range words {
					if slices.Contains(comeTriggerWords, word) {
						return true
					}
				}
				return false
			},
			Execute: func(b *Bot, event twitch.EventChannelChatMessage) {
				b.SendMessage(SendMessageParams{
					Message:              "no coming",
					ReplyParentMessageID: event.MessageId,
				})
			},
		},
	}
}

package bot

import (
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
)

type Command interface {
	ModeratorOnly() bool
	ShouldTrigger(command string) bool
	Execute(bot *Bot, event twitch.EventChannelChatMessage)
}

type CommandHandler struct {
	Commands []Command
}

func NewCommandHandler(commands []Command) *CommandHandler {
	return &CommandHandler{
		Commands: commands,
	}
}

func (h *CommandHandler) Process(bot *Bot, event twitch.EventChannelChatMessage) {
	if !strings.HasPrefix(event.Message.Text, "!") {
		return
	}

	fields := strings.Fields(strings.ToLower(event.Message.Text))
	if len(fields) == 0 {
		return
	}

	cmd := strings.TrimPrefix(fields[0], "!")
	isMod := IsModerator(event)

	bot.logger.Info("chat command received",
		"command", cmd,
		"user", event.ChatterUserName,
		"message", event.Message.Text,
	)

	for _, command := range h.Commands {
		if !command.ShouldTrigger(cmd) {
			continue
		}

		if command.ModeratorOnly() && !isMod {
			bot.logger.Warn("non-moderator attempted mod command",
				"user", event.ChatterUserName,
				"command", cmd,
			)
			bot.SendMessage(SendMessageParams{
				Message:              "You must be a moderator to use this command",
				ReplyParentMessageID: event.MessageId,
			})
			return
		}

		bot.logger.Info("executing command",
			"command", cmd,
			"user", event.ChatterUserName,
		)
		command.Execute(bot, event)
		return
	}
}

func IsModerator(event twitch.EventChannelChatMessage) bool {
	for _, badge := range event.Badges {
		switch badge.SetId {
		case "broadcaster", "moderator", "lead_moderator":
			return true
		}
	}
	return false
}

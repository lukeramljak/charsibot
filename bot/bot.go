package bot

import (
	"charsibot/bot/commands"
	"charsibot/bot/events"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
}

func NewBot(appID string, guildID string, token string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("Bot token is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("Error creating Discord session: %w", err)
	}

	for _, handler := range events.MessageHandlers {
		session.AddHandler(handler)
	}
	session.AddHandler(events.GuildMemberRemove)
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	fmt.Println("Registering commands...")

	_, err = session.ApplicationCommandBulkOverwrite(appID, guildID, commands.Commands)
	if err != nil {
		return nil, fmt.Errorf("Error creating commands: %w", err)
	}

	fmt.Println("Successfully registered " + fmt.Sprint(len(commands.Commands)) + " commands")

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers

	return &Bot{Session: session}, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return fmt.Errorf("Error opening Discord session: %w", err)
	}

	err = b.Session.UpdateListeningStatus("Big Chungus")
	if err != nil {
		return fmt.Errorf("Error setting listening status: %w", err)
	}

	fmt.Println("Ready! Logged in as", b.Session.State.User.Username+". Press CTRL-C to exit.")
	return nil
}

func (b *Bot) Close() error {
	err := b.Session.Close()
	if err != nil {
		return fmt.Errorf("Error closing Discord session: %w", err)
	}
	return nil
}

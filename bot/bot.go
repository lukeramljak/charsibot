package bot

import (
	"charsibot/bot/events"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
}

func NewBot(token string) (*Bot, error) {
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

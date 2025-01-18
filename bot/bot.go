package bot

import (
	"charsibot/bot/commands"
	"charsibot/bot/events"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
}

func NewBot(appID string, guildID string, token string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("Error creating Discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers

	addAllHandlers(session)
	registerCommands(session, appID, guildID)

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

	return nil
}

func (b *Bot) Close() error {
	err := b.Session.Close()
	if err != nil {
		return fmt.Errorf("Error closing Discord session: %w", err)
	}
	return nil
}

func addAllHandlers(s *discordgo.Session) {
	for _, handler := range events.MessageHandlers {
		s.AddHandler(handler)
	}
	s.AddHandler(events.GuildMemberRemove)
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Ready! Logged in as %s. Press CTRL-C to exit.", s.State.User.Username)
	})
}

func registerCommands(s *discordgo.Session, appID string, guildID string) error {
	fmt.Println("Registering commands...")
	_, err := s.ApplicationCommandBulkOverwrite(appID, guildID, commands.Commands)
	if err != nil {
		return fmt.Errorf("Error creating commands: %w", err)
	}
	fmt.Println("Successfully registered " + fmt.Sprint(len(commands.Commands)) + " commands")
	return nil
}

package bot

import (
	"charsibot/bot/commands"
	"charsibot/bot/events"
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

type Bot struct {
	appID   string
	guildID string
	session *discordgo.Session
}

func NewBot() (*Bot, error) {
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	token := os.Getenv("TOKEN")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentGuildMessageReactions

	return &Bot{session: session, appID: appID, guildID: guildID}, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}

	err = b.session.UpdateListeningStatus("Big Chungus")
	if err != nil {
		return fmt.Errorf("error setting listening status: %w", err)
	}

	return nil
}

func (b *Bot) RegisterHandlers() {
	for _, handler := range events.MessageHandlers {
		b.session.AddHandler(handler)
	}
	b.session.AddHandler(events.GuildMemberRemove)
	b.session.AddHandler(events.SinglePollReaction)
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Ready! Logged in as %s. Press CTRL-C to exit.", s.State.User.Username)
	})
}

func (b *Bot) RegisterCommands() error {
	log.Println("Registering commands...")
	createdCommands, err := b.session.ApplicationCommandBulkOverwrite(b.appID, b.guildID, commands.Commands)
	if err != nil {
		return fmt.Errorf("error registering commands: %w", err)
	}
	log.Printf("Successfully registered %d commands\n", len(createdCommands))
	return nil
}

func (b *Bot) Close() error {
	err := b.session.Close()
	if err != nil {
		return fmt.Errorf("error closing Discord session: %w", err)
	}
	return nil
}

package bot

import (
	"fmt"
	"log"
	"os"

	"github.com/lukeramljak/charsibot/discord/bot/commands"
	"github.com/lukeramljak/charsibot/discord/bot/events"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Bot struct {
	appID   string
	guildID string
	session *discordgo.Session
}

func NewBot() (*Bot, error) {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	appID := os.Getenv("DISCORD_APP_ID")
	guildID := os.Getenv("DISCORD_GUILD_ID")
	token := os.Getenv("DISCORD_TOKEN")

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

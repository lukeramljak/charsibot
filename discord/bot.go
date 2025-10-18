package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	appID   string
	guildID string
	session *discordgo.Session
}

func NewBot() (*Bot, error) {
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
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("error opening Discord session: %w", err)
	}

	if err := b.session.UpdateListeningStatus("Big Chungus"); err != nil {
		return fmt.Errorf("error setting listening status: %w", err)
	}

	return nil
}

func (b *Bot) RegisterHandlers() {
	for _, handler := range messageHandlers {
		b.session.AddHandler(handler)
	}
	b.session.AddHandler(handleGuildMemberRemove)
	b.session.AddHandler(handleSinglePollReaction)
	b.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})
	b.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Ready! Logged in as %s. Press CTRL-C to exit.", s.State.User.Username)
	})
}

func (b *Bot) RegisterCommands() error {
	log.Println("Registering commands...")
	createdCommands, err := b.session.ApplicationCommandBulkOverwrite(b.appID, b.guildID, commands)
	if err != nil {
		return fmt.Errorf("error registering commands: %w", err)
	}
	log.Printf("Successfully registered %d commands\n", len(createdCommands))
	return nil
}

func (b *Bot) Close() error {
	if err := b.session.Close(); err != nil {
		return fmt.Errorf("error closing Discord session: %w", err)
	}
	return nil
}

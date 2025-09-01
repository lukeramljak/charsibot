package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ClientID          string
	ClientSecret      string
	OAuthToken        string
	RefreshToken      string
	BotUserID         string
	ChatChannelUserID string
	DbURL             string
	DbAuthToken       string
}

func loadConfig() *Config {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	return &Config{
		ClientID:          os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret:      os.Getenv("TWITCH_CLIENT_SECRET"),
		OAuthToken:        os.Getenv("TWITCH_OAUTH_TOKEN"),
		RefreshToken:      os.Getenv("TWITCH_REFRESH_TOKEN"),
		BotUserID:         os.Getenv("TWITCH_BOT_USER_ID"),
		ChatChannelUserID: os.Getenv("TWITCH_CHANNEL_USER_ID"),
		DbURL:             os.Getenv("TURSO_DATABASE_URL"),
		DbAuthToken:       os.Getenv("TURSO_AUTH_TOKEN"),
	}
}

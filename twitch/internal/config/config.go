package config

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ClientID             string
	ClientSecret         string
	StreamerAccessToken  string
	StreamerRefreshToken string
	BotAccessToken       string
	BotRefreshToken      string
	BotUserID            string
	ChannelUserID        string
	DBPath               string
	UseMockServer        bool
	ServerPort           int
}

func Load() Config {
	serverPort := 8081
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			serverPort = p
		}
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "charsibot.db"
	}

	return Config{
		ClientID:             os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret:         os.Getenv("TWITCH_CLIENT_SECRET"),
		StreamerAccessToken:  os.Getenv("TWITCH_OAUTH_TOKEN"),
		StreamerRefreshToken: os.Getenv("TWITCH_REFRESH_TOKEN"),
		BotAccessToken:       os.Getenv("TWITCH_BOT_OAUTH_TOKEN"),
		BotRefreshToken:      os.Getenv("TWITCH_BOT_REFRESH_TOKEN"),
		BotUserID:            os.Getenv("TWITCH_BOT_USER_ID"),
		ChannelUserID:        os.Getenv("TWITCH_CHANNEL_USER_ID"),
		DBPath:               dbPath,
		UseMockServer:        os.Getenv("USE_MOCK_SERVER") == "true",
		ServerPort:           serverPort,
	}
}

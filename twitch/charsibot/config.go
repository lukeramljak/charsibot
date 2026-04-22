package charsibot

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	ClientID         string
	ClientSecret     string
	BotUserID        string
	ChannelUserID    string
	OAuthRedirectURI string
	DBPath           string
	UseMockServer    bool
	ServerPort       int
	LogLevel         slog.Level
}

func LoadConfig() Config {
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

	redirectURI := os.Getenv("TWITCH_OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("http://localhost:%d/oauth/callback", serverPort)
	}

	logLevel := slog.LevelInfo
	if raw := os.Getenv("LOG_LEVEL"); raw != "" {
		if err := logLevel.UnmarshalText([]byte(raw)); err != nil {
			logLevel = slog.LevelInfo
		}
	}

	return Config{
		ClientID:         os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret:     os.Getenv("TWITCH_CLIENT_SECRET"),
		BotUserID:        os.Getenv("TWITCH_BOT_USER_ID"),
		ChannelUserID:    os.Getenv("TWITCH_CHANNEL_USER_ID"),
		OAuthRedirectURI: redirectURI,
		DBPath:           dbPath,
		UseMockServer:    os.Getenv("USE_MOCK_SERVER") == "true",
		ServerPort:       serverPort,
		LogLevel:         logLevel,
	}
}

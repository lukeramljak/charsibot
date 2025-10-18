package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	devMode := flag.Bool("dev", false, "Enable development mode (use Twitch CLI local websocket)")
	flag.Parse()

	logLevel := slog.LevelInfo
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch levelStr {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		case "ERROR":
			logLevel = slog.LevelError
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	cfg := &Config{
		ClientID:          os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret:      os.Getenv("TWITCH_CLIENT_SECRET"),
		OAuthToken:        os.Getenv("TWITCH_OAUTH_TOKEN"),
		RefreshToken:      os.Getenv("TWITCH_REFRESH_TOKEN"),
		BotOAuthToken:     os.Getenv("TWITCH_BOT_OAUTH_TOKEN"),
		BotRefreshToken:   os.Getenv("TWITCH_BOT_REFRESH_TOKEN"),
		BotUserID:         os.Getenv("TWITCH_BOT_USER_ID"),
		ChatChannelUserID: os.Getenv("TWITCH_CHANNEL_USER_ID"),
		DbURL:             os.Getenv("TURSO_DATABASE_URL"),
		DbAuthToken:       os.Getenv("TURSO_AUTH_TOKEN"),
	}

	url := fmt.Sprintf("%s?authToken=%s", cfg.DbURL, cfg.DbAuthToken)
	db, err := sql.Open("libsql", url)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("connected to database")

	store := NewStore(db)
	bot := NewBot(store, cfg)

	websocketUrl := "wss://eventsub.wss.twitch.tv/ws"
	if *devMode {
		websocketUrl = "ws://127.0.0.1:8080/ws"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("starting Twitch bot", "websocket_url", websocketUrl)
	if err := bot.Connect(ctx, websocketUrl); err != nil && err != context.Canceled {
		slog.Error("bot error", "error", err)
		os.Exit(1)
	}
	slog.Info("bot exited cleanly")
}

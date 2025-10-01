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

	twitchapp "github.com/lukeramljak/charsibot/twitch/internal/twitchapp"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	devMode := flag.Bool("dev", false, "Enable development mode (use Twitch CLI local websocket)")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := &twitchapp.Config{
		ClientID:          os.Getenv("TWITCH_CLIENT_ID"),
		ClientSecret:      os.Getenv("TWITCH_CLIENT_SECRET"),
		OAuthToken:        os.Getenv("TWITCH_OAUTH_TOKEN"),
		RefreshToken:      os.Getenv("TWITCH_REFRESH_TOKEN"),
		BotUserID:         os.Getenv("TWITCH_BOT_USER_ID"),
		ChatChannelUserID: os.Getenv("TWITCH_CHANNEL_USER_ID"),
		DbURL:             os.Getenv("TURSO_DATABASE_URL"),
		DbAuthToken:       os.Getenv("TURSO_AUTH_TOKEN"),
	}

	url := fmt.Sprintf("%s?authToken=%s", cfg.DbURL, cfg.DbAuthToken)
	db, err := sql.Open("libsql", url)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Connected to database")

	client := twitchapp.New(db, cfg)

	websocketUrl := "wss://eventsub.wss.twitch.tv/ws"

	if *devMode {
		websocketUrl = "ws://127.0.0.1:8080/ws"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Starting Twitch client", "websocket_url", websocketUrl)
	err = client.Connect(ctx, websocketUrl)
	if err != nil {
		slog.Error("Failed to connect to Twitch", "error", err)
		os.Exit(1)
	}
	slog.Info("Twitch client exited cleanly")
}

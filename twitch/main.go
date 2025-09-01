package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	localTesting := flag.Bool("local", false, "Use local Twitch CLI WebSocket server for testing")
	flag.Parse()

	logLevel := slog.LevelInfo
	if *debug {
		logLevel = slog.LevelDebug
	} else if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch level {
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	config := loadConfig()

	url := fmt.Sprintf("%s?authToken=%s", config.DbURL, config.DbAuthToken)
	db, err := sql.Open("libsql", url)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("✅ Connected to database")

	twitchClient := NewTwitchClient(config, db)

	if err := twitchClient.ValidateAuth(); err != nil {
		slog.Error("Auth validation failed", "error", err)
		os.Exit(1)
	}

	wsClient := NewWebSocketClient(twitchClient)

	if *localTesting || os.Getenv("TWITCH_LOCAL_TESTING") == "true" {
		wsClient.EnableLocalTesting()
	}

	ctx := context.Background()
	if err := wsClient.Start(ctx); err != nil {
		slog.Error("WebSocket failed", "error", err)
		os.Exit(1)
	}
}

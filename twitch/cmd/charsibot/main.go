package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/twitch/charsibot"
	"github.com/lukeramljak/charsibot/twitch/db"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := charsibot.LoadConfig()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	sqlDB, err := db.Connect(context.Background(), cfg.DBPath, logger)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	queries := db.New(sqlDB)

	srv := charsibot.NewServer(cfg.ServerPort, cfg.ClientID, cfg.ClientSecret, cfg.OAuthRedirectURI, queries)
	if err = srv.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()

	bot, err := charsibot.New(cfg, queries, srv)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		<-sigChan
		slog.Info("received shutdown signal")
		bot.Shutdown()
		close(done)
	}()

	if err := bot.Start(); err != nil {
		return fmt.Errorf("bot: %w", err)
	}

	<-done
	slog.Info("bot shutdown complete")
	return nil
}

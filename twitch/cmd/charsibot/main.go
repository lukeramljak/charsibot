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

	"github.com/lukeramljak/charsibot/twitch/blindbox"
	"github.com/lukeramljak/charsibot/twitch/charsibot"
	"github.com/lukeramljak/charsibot/twitch/db"
	"github.com/lukeramljak/charsibot/twitch/server"
	"github.com/lukeramljak/charsibot/twitch/stats"
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

	blindboxService, err := blindbox.NewService(queries)
	if err != nil {
		return fmt.Errorf("blindbox service: %w", err)
	}
	statsService, err := stats.NewService(queries)
	if err != nil {
		return fmt.Errorf("stats service: %w", err)
	}

	srv := server.NewServer(server.ServerConfig{
		Port:             cfg.ServerPort,
		ClientID:         cfg.ClientID,
		ClientSecret:     cfg.ClientSecret,
		OAuthRedirectURI: cfg.OAuthRedirectURI,
	}, blindboxService)
	if err = srv.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()

	bot, err := charsibot.New(cfg, statsService, blindboxService, srv.Broadcast)
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
		return fmt.Errorf("run bot: %w", err)
	}

	<-done
	slog.Info("bot shutdown complete")
	return nil
}

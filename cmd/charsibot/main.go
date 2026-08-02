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

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/catalog"
	"github.com/lukeramljak/charsibot/charsibot"
	"github.com/lukeramljak/charsibot/db"
	"github.com/lukeramljak/charsibot/server"
	"github.com/lukeramljak/charsibot/stats"
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

	sqlDB, err := db.Connect(context.Background(), cfg.DBPath, logger)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	appCatalog, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("load catalog: %w", err)
	}

	queries := db.New(sqlDB)

	blindboxService, err := blindbox.NewService(queries, appCatalog.Series)
	if err != nil {
		return fmt.Errorf("blindbox service: %w", err)
	}
	statsService, err := stats.NewService(queries, appCatalog.Stats)
	if err != nil {
		return fmt.Errorf("stats service: %w", err)
	}

	srv := server.NewServer(server.ServerConfig{
		Port:             cfg.ServerPort,
		ClientID:         cfg.ClientID,
		ClientSecret:     cfg.ClientSecret,
		OAuthRedirectURI: cfg.OAuthRedirectURI,
		StatsService:     statsService,
		BlindBoxService:  blindboxService,
		Series:           appCatalog.Series,
	}, logger)
	if err = srv.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	defer srv.Stop()

	bot, err := charsibot.New(cfg, logger, statsService, blindboxService, appCatalog.Series, srv.Broadcast)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	srv.SetAdminChatMessage(func(message string) {
		bot.SendMessage(charsibot.SendMessageParams{Message: message})
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		bot.Shutdown()
		close(done)
	}()

	if err := bot.Start(); err != nil {
		return fmt.Errorf("run bot: %w", err)
	}

	<-done
	logger.Info("bot shutdown complete")
	return nil
}

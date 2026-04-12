package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/config"
	"github.com/lukeramljak/charsibot/internal/server"
	"github.com/lukeramljak/charsibot/internal/store"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(nil)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.Up(db, "internal/store/migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	queries := store.New(db)

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})))

	overlayServer := server.NewServer(cfg.ServerPort, cfg.ClientID, cfg.ClientSecret, cfg.OAuthRedirectURI, queries)
	if err := overlayServer.Start(); err != nil {
		return fmt.Errorf("start overlay server: %w", err)
	}
	defer overlayServer.Stop()

	twitchBot, err := bot.New(cfg, queries, overlayServer)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		<-sigChan
		slog.Info("received shutdown signal")
		twitchBot.Shutdown()
		close(done)
	}()

	if err := twitchBot.Start(); err != nil {
		return fmt.Errorf("bot: %w", err)
	}

	<-done
	slog.Info("bot shutdown complete")
	return nil
}

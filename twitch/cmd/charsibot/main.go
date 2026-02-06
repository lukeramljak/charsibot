package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lukeramljak/charsibot/internal/blindbox"
	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/config"
	"github.com/lukeramljak/charsibot/internal/stats"
	"github.com/lukeramljak/charsibot/internal/store"
	"github.com/lukeramljak/charsibot/internal/triggers"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	goose.SetBaseFS(nil)
	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "internal/store/migrations"); err != nil {
		log.Fatal(err)
	}

	queries := store.New(db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()
	if streamerTokens, err := queries.GetTokens(ctx, "streamer"); err == nil {
		logger.Info("loaded streamer tokens from database")
		cfg.StreamerAccessToken = streamerTokens.AccessToken
		cfg.StreamerRefreshToken = streamerTokens.RefreshToken
	} else {
		logger.Info("using streamer tokens from environment variables")
	}

	if botTokens, err := queries.GetTokens(ctx, "bot"); err == nil {
		logger.Info("loaded bot tokens from database")
		cfg.BotAccessToken = botTokens.AccessToken
		cfg.BotRefreshToken = botTokens.RefreshToken
	} else {
		logger.Info("using bot tokens from environment variables")
	}

	commands := []bot.Command{
		stats.NewStatsCommand(),
		stats.NewLeaderboardCommand(),
		stats.NewModifyStatCommand(),
		stats.NewExplodeCommand(),
		blindbox.NewCompletedCollectionsCommand(),
	}

	for _, bbConfig := range blindbox.BlindBoxConfigs {
		commands = append(commands,
			blindbox.NewRedeemCommand(bbConfig),
			blindbox.NewResetCommand(bbConfig),
			blindbox.NewShowCollectionCommand(bbConfig),
		)
	}

	redemptions := []bot.Redemption{
		stats.NewPotionRedemption(),
		stats.NewTemptDiceRedemption(),
	}

	for _, bbConfig := range blindbox.BlindBoxConfigs {
		redemptions = append(redemptions, blindbox.NewBlindBoxRedemption(bbConfig))
	}

	triggersList := []bot.Trigger{
		triggers.NewComeTrigger(),
	}

	commandHandler := bot.NewCommandHandler(commands)
	triggerHandler := bot.NewTriggerHandler(triggersList)
	redemptionHandler := bot.NewRedemptionHandler(redemptions)

	twitchBot, err := bot.New(cfg, queries, logger, commandHandler, triggerHandler, redemptionHandler)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("received shutdown signal")
		cancel()
		twitchBot.Shutdown()
	}()

	if err := twitchBot.Start(); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	logger.Info("bot shutdown complete")
}

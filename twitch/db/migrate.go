package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *sql.DB, migrationsDir string, logger *slog.Logger) error {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, os.DirFS(migrationsDir),
		goose.WithSlog(logger),
		goose.WithVerbose(true),
	)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	current, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("get current migration version: %w", err)
	}

	if _, err = provider.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	newVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	if newVersion > current {
		logger.Info("successfully migrated database", "previous_version", current, "new_version", newVersion)
	} else {
		logger.Info("database is already up to date", "version", current)
	}

	return nil
}

package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

const migrationsDir = "internal/store/migrations"

func Migrate(ctx context.Context, db *sql.DB, logger *slog.Logger) error {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, os.DirFS(migrationsDir),
		goose.WithSlog(logger),
		goose.WithVerbose(true),
	)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

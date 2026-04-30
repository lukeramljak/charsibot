package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var FS embed.FS

// Connect opens a SQLite database connection and runs migrations.
func Connect(ctx context.Context, dbPath string, logger *slog.Logger) (*sql.DB, error) {
	pragmas := map[string]string{
		"foreign_keys":  "ON",
		"journal_mode":  "WAL",
		"page_size":     "4096",
		"cache_size":    "-8000",
		"synchronous":   "NORMAL",
		"secure_delete": "ON",
		"busy_timeout":  "30000",
	}

	params := url.Values{}
	for name, value := range pragmas {
		params.Add("_pragma", fmt.Sprintf("%s(%s)", name, value))
	}

	dsn := fmt.Sprintf("file:%s?%s", dbPath, params.Encode())
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	migrations, err := fs.Sub(FS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("get migrations fs: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations,
		goose.WithSlog(logger),
		goose.WithVerbose(true),
	)
	if err != nil {
		return nil, fmt.Errorf("create migration provider: %w", err)
	}

	current, err := provider.GetDBVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("get current migration version: %w", err)
	}

	if _, err = provider.Up(ctx); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	newVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return nil, err
	}

	if newVersion > current {
		logger.Info("successfully migrated database", "previous_version", current, "new_version", newVersion)
	} else {
		logger.Info("database is already up to date", "version", current)
	}

	return db, nil
}

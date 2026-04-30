package db

import (
	"database/sql"
	"io/fs"
	"testing"

	"github.com/pressly/goose/v3"
)

// NewTestDB opens an in-memory SQLite database and runs all migrations.
func NewTestDB(t *testing.T) (*Queries, *sql.DB) {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	migrations, err := fs.Sub(FS, "migrations")
	if err != nil {
		t.Fatalf("failed to get migrations fs: %v", err)
	}

	provider, err := goose.NewProvider(goose.DialectSQLite3, sqlDB, migrations)
	if err != nil {
		t.Fatalf("failed to create migration provider: %v", err)
	}

	if _, err = provider.Up(t.Context()); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return New(sqlDB), sqlDB
}

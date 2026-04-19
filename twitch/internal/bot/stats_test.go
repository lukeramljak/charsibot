package bot

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/twitch/internal/store"
)

func setupStatsTestDB(t *testing.T) (*store.Queries, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
	CREATE TABLE stat_definitions (
		name          TEXT PRIMARY KEY,
		short_name    TEXT NOT NULL,
		long_name     TEXT NOT NULL,
		default_value INTEGER NOT NULL DEFAULT 3,
		sort_order    INTEGER NOT NULL
	);

	INSERT INTO stat_definitions (name, short_name, long_name, default_value, sort_order) VALUES
		('strength',     'STR',   'Strength',     3, 1),
		('intelligence', 'INT',   'Intelligence', 3, 2),
		('charisma',     'CHA',   'Charisma',     3, 3),
		('luck',         'LUCK',  'Luck',         3, 4),
		('dexterity',    'DEX',   'Dexterity',    3, 5),
		('penis',        'PENIS', 'Penis',        3, 6);

	CREATE TABLE user_stats (
		user_id   TEXT NOT NULL,
		username  TEXT NOT NULL,
		stat_name TEXT NOT NULL REFERENCES stat_definitions(name),
		value     INTEGER NOT NULL DEFAULT 3,
		PRIMARY KEY (user_id, stat_name)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return store.New(db), db
}

func TestGetOrCreateStats(t *testing.T) {
	queries, db := setupStatsTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("creates stat rows with defaults on first call", func(t *testing.T) {
		stats, err := GetOrCreateStats(ctx, queries, "user1", "alice")
		if err != nil {
			t.Fatalf("GetOrCreateStats failed: %v", err)
		}
		if len(stats) != 6 {
			t.Fatalf("expected 6 stat rows, got %d", len(stats))
		}
		if stats[0].ShortName != "STR" {
			t.Errorf("first stat = %q, want STR", stats[0].ShortName)
		}
		if stats[0].Value != 3 {
			t.Errorf("default value = %d, want 3", stats[0].Value)
		}
	})

	t.Run("is idempotent - second call does not duplicate rows", func(t *testing.T) {
		GetOrCreateStats(ctx, queries, "user2", "bob")
		stats, err := GetOrCreateStats(ctx, queries, "user2", "bob")
		if err != nil {
			t.Fatalf("second GetOrCreateStats failed: %v", err)
		}
		if len(stats) != 6 {
			t.Errorf("expected 6 stat rows after second call, got %d", len(stats))
		}
	})

	t.Run("updates username on subsequent calls", func(t *testing.T) {
		GetOrCreateStats(ctx, queries, "user3", "oldname")
		stats, err := GetOrCreateStats(ctx, queries, "user3", "newname")
		if err != nil {
			t.Fatalf("GetOrCreateStats failed: %v", err)
		}

		row := db.QueryRowContext(ctx, `SELECT username FROM user_stats WHERE user_id = ? LIMIT 1`, "user3")
		var username string
		if err := row.Scan(&username); err != nil {
			t.Fatalf("failed to query username: %v", err)
		}
		if username != "newname" {
			t.Errorf("username = %q, want %q", username, "newname")
		}
		_ = stats
	})
}

func TestFormatStats(t *testing.T) {
	stats := []store.GetUserStatsRow{
		{Name: "strength", ShortName: "STR", LongName: "Strength", Value: 5},
		{Name: "intelligence", ShortName: "INT", LongName: "Intelligence", Value: 5},
		{Name: "charisma", ShortName: "CHA", LongName: "Charisma", Value: 3},
		{Name: "luck", ShortName: "LUCK", LongName: "Luck", Value: 3},
		{Name: "dexterity", ShortName: "DEX", LongName: "Dexterity", Value: 3},
		{Name: "penis", ShortName: "PENIS", LongName: "Penis", Value: 3},
	}

	formatted := FormatStats("testuser", stats)
	expected := "testuser's stats: STR: 5 | INT: 5 | CHA: 3 | LUCK: 3 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

func TestFormatStatsNegative(t *testing.T) {
	stats := []store.GetUserStatsRow{
		{Name: "strength", ShortName: "STR", LongName: "Strength", Value: 3},
		{Name: "intelligence", ShortName: "INT", LongName: "Intelligence", Value: 3},
		{Name: "charisma", ShortName: "CHA", LongName: "Charisma", Value: 9},
		{Name: "luck", ShortName: "LUCK", LongName: "Luck", Value: -2},
		{Name: "dexterity", ShortName: "DEX", LongName: "Dexterity", Value: 3},
		{Name: "penis", ShortName: "PENIS", LongName: "Penis", Value: 3},
	}

	formatted := FormatStats("testuser", stats)
	expected := "testuser's stats: STR: 3 | INT: 3 | CHA: 9 | LUCK: -2 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

func TestFormatStatsEmpty(t *testing.T) {
	formatted := FormatStats("testuser", []store.GetUserStatsRow{})
	expected := "testuser's stats: "

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

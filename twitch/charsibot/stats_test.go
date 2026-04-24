package charsibot

import (
	"context"
	"testing"

	"github.com/lukeramljak/charsibot/twitch/db"
)

func TestGetOrCreateStats(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
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

		row := sqlDB.QueryRowContext(ctx, `SELECT username FROM user_stats WHERE user_id = ? LIMIT 1`, "user3")
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
	stats := []db.GetUserStatsRow{
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
	stats := []db.GetUserStatsRow{
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
	formatted := FormatStats("testuser", []db.GetUserStatsRow{})
	expected := "testuser's stats: "

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

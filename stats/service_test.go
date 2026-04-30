package stats_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/db"
	"github.com/lukeramljak/charsibot/stats"
)

func TestGetOrCreateStats(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	svc, err := stats.NewService(queries)
	if err != nil {
		t.Fatalf("failed to create stats service: %v", err)
	}

	t.Run("creates stat rows with defaults on first call", func(t *testing.T) {
		stats, err := svc.GetOrCreateStats(ctx, "user1", "alice")
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
		svc.GetOrCreateStats(ctx, "user2", "bob")
		stats, err := svc.GetOrCreateStats(ctx, "user2", "bob")
		if err != nil {
			t.Fatalf("second GetOrCreateStats failed: %v", err)
		}
		if len(stats) != 6 {
			t.Errorf("expected 6 stat rows after second call, got %d", len(stats))
		}
	})

	t.Run("updates username on subsequent calls", func(t *testing.T) {
		svc.GetOrCreateStats(ctx, "user3", "oldname")
		stats, err := svc.GetOrCreateStats(ctx, "user3", "newname")
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

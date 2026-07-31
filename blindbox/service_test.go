package blindbox_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/catalog"
	"github.com/lukeramljak/charsibot/db"
)

func TestResetCollection(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	_ = svc

	for _, key := range []string{"cutey", "blueberry", "secret"} {
		seedPlushie(t, queries, ctx, "reset1", "charlie", key)
	}

	err := queries.ResetUserPlushies(ctx, db.ResetUserPlushiesParams{
		UserID: "reset1",
		Series: "coobubu",
	})
	if err != nil {
		t.Fatalf("ResetUserPlushies failed: %v", err)
	}

	keys, err := queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: "reset1",
		Series: "coobubu",
	})
	if err != nil {
		t.Fatalf("GetCollectedPlushies failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty collection after reset, got %v", keys)
	}
}

func TestAddNewPlushieToCollection(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	seedPlushie(t, queries, ctx, "user1", "alice", "cutey")

	isNew, collection, err := svc.AddPlushieToCollection(ctx, "user1", "alice", "coobubu", "blueberry")
	if err != nil {
		t.Fatalf("AddPlushieToCollection failed: %v", err)
	}
	if !isNew {
		t.Error("expected isNew to be true for new key")
	}
	if len(collection) != 2 {
		t.Errorf("expected 2 collected keys, got %d: %v", len(collection), collection)
	}
}

func TestAddExistingPlushieToCollection(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	seedPlushie(t, queries, ctx, "user2", "bob", "cutey")

	isNew, _, err := svc.AddPlushieToCollection(ctx, "user2", "bob", "coobubu", "cutey")
	if err != nil {
		t.Fatalf("AddPlushieToCollection failed: %v", err)
	}
	if isNew {
		t.Error("expected isNew to be false for existing key")
	}
}

func TestAddExistingPlushieSyncsUsername(t *testing.T) {
	svc, queries, sqlDB, ctx := newBlindboxService(t)
	seedPlushie(t, queries, ctx, "user3", "oldname", "cutey")

	isNew, _, err := svc.AddPlushieToCollection(ctx, "user3", "newname", "coobubu", "cutey")
	if err != nil {
		t.Fatalf("AddPlushieToCollection failed: %v", err)
	}
	if isNew {
		t.Error("expected isNew to be false for existing key")
	}

	row := sqlDB.QueryRowContext(
		ctx,
		`SELECT username FROM user_plushies WHERE user_id = ? AND series = ? AND key = ?`,
		"user3",
		"coobubu",
		"cutey",
	)
	var username string
	if err := row.Scan(&username); err != nil {
		t.Fatalf("failed to query username: %v", err)
	}
	if username != "newname" {
		t.Errorf("username = %q, want %q", username, "newname")
	}
}

func TestCompletedCollectionsEmptyForPartialUser(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	seedPlushie(t, queries, ctx, "partial1", "alice", "cutey")
	seedPlushie(t, queries, ctx, "partial1", "alice", "blueberry")

	rows, err := svc.GetCompletedCollections(ctx)
	if err != nil {
		t.Fatalf("GetCompletedCollections failed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected no completed collections, got %v", rows)
	}
}

func TestCompletedCollectionsUsesSeriesName(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	seedCompleteCoobubu(t, queries, ctx, "complete1", "bob")

	rows, err := svc.GetCompletedCollections(ctx)
	if err != nil {
		t.Fatalf("GetCompletedCollections failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 completed collection, got %d", len(rows))
	}
	if rows[0].SeriesName != "Coobubus" {
		t.Errorf("SeriesName = %q, want %q", rows[0].SeriesName, "Coobubus")
	}
}

func TestCompletedCollectionsAggregatesUsers(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	seedCompleteCoobubu(t, queries, ctx, "complete2", "carol")
	seedCompleteCoobubu(t, queries, ctx, "complete3", "dave")

	rows, err := svc.GetCompletedCollections(ctx)
	if err != nil {
		t.Fatalf("GetCompletedCollections failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for series, got %d", len(rows))
	}
	if !strings.Contains(rows[0].Usernames, "carol") || !strings.Contains(rows[0].Usernames, "dave") {
		t.Errorf("Usernames = %q, want both carol and dave", rows[0].Usernames)
	}
}

func TestCompletedCollectionsGroupsByStableUserID(t *testing.T) {
	svc, queries, _, ctx := newBlindboxService(t)
	keys := []string{"cutey", "blueberry", "lemony", "bibi", "pinky", "minty", "cherry", "secret"}
	for i, key := range keys {
		username := "oldname"
		if i >= len(keys)/2 {
			username = "newname"
		}
		seedPlushie(t, queries, ctx, "renamed-user", username, key)
	}

	rows, err := svc.GetCompletedCollections(ctx)
	if err != nil {
		t.Fatalf("GetCompletedCollections failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for renamed user, got %d", len(rows))
	}
}

func newBlindboxService(t *testing.T) (*blindbox.Service, *db.Queries, *sql.DB, context.Context) {
	t.Helper()
	queries, sqlDB := db.NewTestDB(t)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	})
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	svc, err := blindbox.NewService(queries, cat.Series)
	if err != nil {
		t.Fatalf("failed to create blindbox service: %v", err)
	}
	return svc, queries, sqlDB, context.Background()
}

func seedCompleteCoobubu(t *testing.T, queries *db.Queries, ctx context.Context, userID, username string) {
	t.Helper()
	for _, key := range []string{"cutey", "blueberry", "lemony", "bibi", "pinky", "minty", "cherry", "secret"} {
		seedPlushie(t, queries, ctx, userID, username, key)
	}
}

func seedPlushie(t *testing.T, queries *db.Queries, ctx context.Context, userID, username, key string) {
	t.Helper()
	if err := queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
		UserID:   userID,
		Username: username,
		Series:   "coobubu",
		Key:      key,
	}); err != nil {
		t.Fatalf("UpsertUserPlushie failed: %v", err)
	}
}

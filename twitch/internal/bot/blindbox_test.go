package bot

import (
	"context"
	"database/sql"
	"testing"

	"github.com/lukeramljak/charsibot/internal/store"
	_ "modernc.org/sqlite"
)

func setupBlindBoxTestDB(t *testing.T) (*store.Queries, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	schema := `
	CREATE TABLE blind_box_series (
		series           TEXT PRIMARY KEY,
		redemption_title TEXT NOT NULL,
		name             TEXT NOT NULL DEFAULT '',
		reveal_sound     TEXT NOT NULL DEFAULT '',
		box_front_face   TEXT NOT NULL DEFAULT '',
		box_side_face    TEXT NOT NULL DEFAULT '',
		display_color    TEXT NOT NULL DEFAULT '',
		text_color       TEXT NOT NULL DEFAULT ''
	);

	CREATE TABLE blind_box_plushies (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		series      TEXT NOT NULL REFERENCES blind_box_series(series),
		key         TEXT NOT NULL,
		sort_order  INTEGER NOT NULL DEFAULT 0,
		weight      INTEGER NOT NULL DEFAULT 1,
		name        TEXT NOT NULL DEFAULT '',
		image       TEXT NOT NULL DEFAULT '',
		empty_image TEXT NOT NULL DEFAULT '',
		UNIQUE(series, key)
	);

	INSERT INTO blind_box_series (series, redemption_title, name) VALUES
		('coobubu', 'Cooper Series Blind Box', 'Coobubus');

	INSERT INTO blind_box_plushies (series, key, sort_order, weight) VALUES
		('coobubu', 'cutey',     1, 12),
		('coobubu', 'blueberry', 2, 12),
		('coobubu', 'secret',    3,  1);

	CREATE TABLE user_plushies (
		user_id  TEXT NOT NULL,
		username TEXT NOT NULL,
		series   TEXT NOT NULL REFERENCES blind_box_series(series),
		key      TEXT NOT NULL,
		PRIMARY KEY (user_id, series, key)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return store.New(db), db
}

func TestCollections(t *testing.T) {
	queries, db := setupBlindBoxTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("resets collection", func(t *testing.T) {
		for _, key := range []string{"cutey", "blueberry", "secret"} {
			queries.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
				UserID:   "reset1",
				Username: "charlie",
				Series:   "coobubu",
				Key:      key,
			})
		}

		err := queries.ResetUserPlushies(ctx, store.ResetUserPlushiesParams{
			UserID: "reset1",
			Series: "coobubu",
		})
		if err != nil {
			t.Fatalf("ResetUserPlushies failed: %v", err)
		}

		keys, err := queries.GetCollectedPlushies(ctx, store.GetCollectedPlushiesParams{
			UserID: "reset1",
			Series: "coobubu",
		})
		if err != nil {
			t.Fatalf("GetCollectedPlushies failed: %v", err)
		}

		if len(keys) != 0 {
			t.Errorf("expected empty collection after reset, got %v", keys)
		}
	})

	t.Run("adds plushie to collection - new reward", func(t *testing.T) {
		// Pre-seed one key so we can test isNew=false later
		queries.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
			UserID:   "user1",
			Username: "alice",
			Series:   "coobubu",
			Key:      "cutey",
		})

		isNew, collection, err := addPlushieToCollection(ctx, queries, "user1", "alice", "coobubu", "blueberry")
		if err != nil {
			t.Fatalf("AddPlushieToCollection failed: %v", err)
		}

		if !isNew {
			t.Error("expected isNew to be true for new key")
		}

		if len(collection) != 2 {
			t.Errorf("expected 2 collected keys, got %d: %v", len(collection), collection)
		}
	})

	t.Run("adds plushie to collection - existing reward", func(t *testing.T) {
		queries.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
			UserID:   "user2",
			Username: "bob",
			Series:   "coobubu",
			Key:      "cutey",
		})

		isNew, _, err := addPlushieToCollection(ctx, queries, "user2", "bob", "coobubu", "cutey")
		if err != nil {
			t.Fatalf("AddPlushieToCollection failed: %v", err)
		}

		if isNew {
			t.Error("expected isNew to be false for existing key")
		}
	})

	t.Run("syncs username when re-adding existing plushie", func(t *testing.T) {
		queries.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
			UserID:   "user3",
			Username: "oldname",
			Series:   "coobubu",
			Key:      "cutey",
		})

		isNew, _, err := addPlushieToCollection(ctx, queries, "user3", "newname", "coobubu", "cutey")
		if err != nil {
			t.Fatalf("addPlushieToCollection failed: %v", err)
		}
		if isNew {
			t.Error("expected isNew to be false for existing key")
		}

		row := db.QueryRowContext(ctx, `SELECT username FROM user_plushies WHERE user_id = ? AND series = ? AND key = ?`, "user3", "coobubu", "cutey")
		var username string
		if err := row.Scan(&username); err != nil {
			t.Fatalf("failed to query username: %v", err)
		}
		if username != "newname" {
			t.Errorf("username = %q, want %q", username, "newname")
		}
	})
}

func TestGetCompletedCollections(t *testing.T) {
	queries, db := setupBlindBoxTestDB(t)
	defer db.Close()
	ctx := context.Background()

	seed := func(userID, username, key string) {
		t.Helper()
		queries.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
			UserID: userID, Username: username, Series: "coobubu", Key: key,
		})
	}

	t.Run("returns empty when no one has completed a collection", func(t *testing.T) {
		seed("partial1", "alice", "cutey")
		seed("partial1", "alice", "blueberry")
		// missing "secret" — not complete

		rows, err := queries.GetCompletedCollections(ctx)
		if err != nil {
			t.Fatalf("GetCompletedCollections failed: %v", err)
		}
		if len(rows) != 0 {
			t.Errorf("expected no completed collections, got %v", rows)
		}
	})

	t.Run("returns series name not key", func(t *testing.T) {
		for _, key := range []string{"cutey", "blueberry", "secret"} {
			seed("complete1", "bob", key)
		}

		rows, err := queries.GetCompletedCollections(ctx)
		if err != nil {
			t.Fatalf("GetCompletedCollections failed: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 completed collection, got %d", len(rows))
		}
		if rows[0].SeriesName != "Coobubus" {
			t.Errorf("SeriesName = %q, want %q", rows[0].SeriesName, "Coobubus")
		}
	})

	t.Run("aggregates multiple completers into one row per series", func(t *testing.T) {
		for _, key := range []string{"cutey", "blueberry", "secret"} {
			seed("complete2", "carol", key)
			seed("complete3", "dave", key)
		}

		rows, err := queries.GetCompletedCollections(ctx)
		if err != nil {
			t.Fatalf("GetCompletedCollections failed: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row for series, got %d", len(rows))
		}
		if !containsAll(rows[0].Usernames, "carol", "dave") {
			t.Errorf("Usernames = %q, want both carol and dave", rows[0].Usernames)
		}
	})
}

func TestGetAllSeriesWithPlushies(t *testing.T) {
	queries, db := setupBlindBoxTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("returns plushies grouped under their series", func(t *testing.T) {
		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		if len(rows) != 3 {
			t.Fatalf("expected 3 rows (one per plushie), got %d", len(rows))
		}
		for _, r := range rows {
			if r.Series != "coobubu" {
				t.Errorf("Series = %q, want %q", r.Series, "coobubu")
			}
			if !r.PlushieKey.Valid {
				t.Error("expected PlushieKey to be valid")
			}
		}
	})

	t.Run("plushies are ordered by sort_order within a series", func(t *testing.T) {
		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		keys := make([]string, len(rows))
		for i, r := range rows {
			keys[i] = r.PlushieKey.String
		}
		want := []string{"cutey", "blueberry", "secret"}
		for i, k := range want {
			if keys[i] != k {
				t.Errorf("plushie[%d] = %q, want %q", i, keys[i], k)
			}
		}
	})

	t.Run("series with no plushies returns one row with null plushie fields", func(t *testing.T) {
		db.ExecContext(ctx, `INSERT INTO blind_box_series (series, redemption_title, name) VALUES ('empty', 'Empty Series', 'Empty')`)

		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		var emptyRow *store.GetAllSeriesWithPlushiesRow
		for i := range rows {
			if rows[i].Series == "empty" {
				emptyRow = &rows[i]
				break
			}
		}
		if emptyRow == nil {
			t.Fatal("expected a row for the 'empty' series")
		}
		if emptyRow.PlushieKey.Valid {
			t.Errorf("expected PlushieKey to be NULL for series with no plushies, got %q", emptyRow.PlushieKey.String)
		}
	})

	t.Run("multiple series are ordered by series key", func(t *testing.T) {
		// 'aardvark' sorts before 'coobubu'
		db.ExecContext(ctx, `INSERT INTO blind_box_series (series, redemption_title, name) VALUES ('aardvark', 'Aardvark Series', 'Aardvarks')`)
		db.ExecContext(ctx, `INSERT INTO blind_box_plushies (series, key, sort_order, weight) VALUES ('aardvark', 'andy', 1, 1)`)

		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		if rows[0].Series != "aardvark" {
			t.Errorf("first series = %q, want %q", rows[0].Series, "aardvark")
		}
	})
}

// containsAll reports whether s contains all the given substrings.
func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		found := false
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

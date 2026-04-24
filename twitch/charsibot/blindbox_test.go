package charsibot

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lukeramljak/charsibot/twitch/db"
)

func TestCollections(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	t.Run("resets collection", func(t *testing.T) {
		for _, key := range []string{"cutey", "blueberry", "secret"} {
			queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
				UserID:   "reset1",
				Username: "charlie",
				Series:   "coobubu",
				Key:      key,
			})
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
	})

	t.Run("adds plushie to collection - new reward", func(t *testing.T) {
		// Pre-seed one key so we can test isNew=false later
		queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
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
		queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
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
		queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
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
	})
}

func TestGetCompletedCollections(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	seed := func(userID, username, key string) {
		t.Helper()
		queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
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
		for _, key := range []string{"cutey", "blueberry", "lemony", "bibi", "pinky", "minty", "cherry", "secret"} {
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
		for _, key := range []string{"cutey", "blueberry", "lemony", "bibi", "pinky", "minty", "cherry", "secret"} {
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
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	t.Run("returns plushies grouped under their series", func(t *testing.T) {
		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		if len(rows) != 40 {
			t.Fatalf("expected 40 rows (one per plushie across all series), got %d", len(rows))
		}
		for _, r := range rows {
			if !r.PlushieKey.Valid {
				t.Errorf("expected PlushieKey to be valid for series %q", r.Series)
			}
		}
	})

	t.Run("plushies are ordered by sort_order within a series", func(t *testing.T) {
		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		// Filter to coobubu rows only
		var coobubuKeys []string
		for _, r := range rows {
			if r.Series == "coobubu" {
				coobubuKeys = append(coobubuKeys, r.PlushieKey.String)
			}
		}
		want := []string{"cutey", "blueberry", "lemony", "bibi", "pinky", "minty", "cherry", "secret"}
		for i, k := range want {
			if coobubuKeys[i] != k {
				t.Errorf("coobubu plushie[%d] = %q, want %q", i, coobubuKeys[i], k)
			}
		}
	})

	t.Run("series with no plushies returns one row with null plushie fields", func(t *testing.T) {
		sqlDB.ExecContext(
			ctx,
			`INSERT INTO blind_box_series (series, redemption_title, name) VALUES ('empty', 'Empty Series', 'Empty')`,
		)

		rows, err := queries.GetAllSeriesWithPlushies(ctx)
		if err != nil {
			t.Fatalf("GetAllSeriesWithPlushies failed: %v", err)
		}

		var emptyRow *db.GetAllSeriesWithPlushiesRow
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
		sqlDB.ExecContext(
			ctx,
			`INSERT INTO blind_box_series (series, redemption_title, name) VALUES ('aardvark', 'Aardvark Series', 'Aardvarks')`,
		)
		sqlDB.ExecContext(
			ctx,
			`INSERT INTO blind_box_plushies (series, key, sort_order, weight) VALUES ('aardvark', 'andy', 1, 1)`,
		)

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

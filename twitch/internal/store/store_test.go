package store

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*Queries, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create tables
	schema := `
	CREATE TABLE oauth_tokens (
		token_type TEXT PRIMARY KEY,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		updated_at REAL DEFAULT (unixepoch())
	);

	CREATE TABLE stats (
		id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		strength INTEGER NOT NULL DEFAULT 3,
		intelligence INTEGER NOT NULL DEFAULT 3,
		charisma INTEGER NOT NULL DEFAULT 3,
		luck INTEGER NOT NULL DEFAULT 3,
		dexterity INTEGER NOT NULL DEFAULT 3,
		penis INTEGER NOT NULL DEFAULT 3
	);

	CREATE TABLE user_collections (
		user_id TEXT,
		username TEXT NOT NULL,
		collection_type TEXT,
		reward1 INTEGER DEFAULT 0,
		reward2 INTEGER DEFAULT 0,
		reward3 INTEGER DEFAULT 0,
		reward4 INTEGER DEFAULT 0,
		reward5 INTEGER DEFAULT 0,
		reward6 INTEGER DEFAULT 0,
		reward7 INTEGER DEFAULT 0,
		reward8 INTEGER DEFAULT 0,
		PRIMARY KEY (user_id, collection_type)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return New(db), db
}

func TestTokens(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("saves and retrieves streamer tokens", func(t *testing.T) {
		err := queries.SaveTokens(ctx, SaveTokensParams{
			TokenType:    "streamer",
			AccessToken:  "access123",
			RefreshToken: "refresh456",
		})
		if err != nil {
			t.Fatalf("SaveTokens failed: %v", err)
		}

		tokens, err := queries.GetTokens(ctx, "streamer")
		if err != nil {
			t.Fatalf("GetTokens failed: %v", err)
		}

		if tokens.AccessToken != "access123" {
			t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, "access123")
		}
		if tokens.RefreshToken != "refresh456" {
			t.Errorf("RefreshToken = %q, want %q", tokens.RefreshToken, "refresh456")
		}
	})

	t.Run("updates existing tokens", func(t *testing.T) {
		queries.SaveTokens(ctx, SaveTokensParams{
			TokenType:    "bot",
			AccessToken:  "old_access",
			RefreshToken: "old_refresh",
		})

		queries.SaveTokens(ctx, SaveTokensParams{
			TokenType:    "bot",
			AccessToken:  "new_access",
			RefreshToken: "new_refresh",
		})

		tokens, err := queries.GetTokens(ctx, "bot")
		if err != nil {
			t.Fatalf("GetTokens failed: %v", err)
		}

		if tokens.AccessToken != "new_access" {
			t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, "new_access")
		}
	})
}

func TestStats(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("creates new user stats with defaults", func(t *testing.T) {
		err := queries.UpsertStatsUser(ctx, UpsertStatsUserParams{
			ID:       "user123",
			Username: "testuser",
		})
		if err != nil {
			t.Fatalf("UpsertStatsUser failed: %v", err)
		}

		stats, err := queries.GetStats(ctx, "user123")
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}

		if stats.Username != "testuser" {
			t.Errorf("Username = %q, want %q", stats.Username, "testuser")
		}
		if stats.Strength != 3 {
			t.Errorf("Strength = %d, want 3", stats.Strength)
		}
	})

	t.Run("modifies stats", func(t *testing.T) {
		queries.UpsertStatsUser(ctx, UpsertStatsUserParams{
			ID:       "user456",
			Username: "alice",
		})

		err := queries.ModifyStat(ctx, "user456", "alice", "strength", 5)
		if err != nil {
			t.Fatalf("ModifyStat failed: %v", err)
		}

		stats, err := queries.GetStats(ctx, "user456")
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}

		if stats.Strength != 8 {
			t.Errorf("Strength = %d, want 8", stats.Strength)
		}
	})

	t.Run("accumulates stat modifications", func(t *testing.T) {
		queries.UpsertStatsUser(ctx, UpsertStatsUserParams{
			ID:       "user789",
			Username: "bob",
		})

		queries.ModifyStat(ctx, "user789", "bob", "luck", 2)
		queries.ModifyStat(ctx, "user789", "bob", "luck", 3)

		stats, err := queries.GetStats(ctx, "user789")
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}

		if stats.Luck != 8 {
			t.Errorf("Luck = %d, want 8", stats.Luck)
		}
	})

	t.Run("handles negative modifications", func(t *testing.T) {
		queries.UpsertStatsUser(ctx, UpsertStatsUserParams{
			ID:       "user999",
			Username: "charlie",
		})

		queries.ModifyStat(ctx, "user999", "charlie", "charisma", 10)
		queries.ModifyStat(ctx, "user999", "charlie", "charisma", -3)

		stats, err := queries.GetStats(ctx, "user999")
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}

		if stats.Charisma != 10 {
			t.Errorf("Charisma = %d, want 10", stats.Charisma)
		}
	})
}

func TestCollections(t *testing.T) {
	queries, db := setupTestDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("adds plushie to collection - new reward", func(t *testing.T) {
		// Start with reward1
		queries.AddReward1(ctx, AddReward1Params{
			UserID:         sql.NullString{String: "user1", Valid: true},
			Username:       "alice",
			CollectionType: sql.NullString{String: "coobubu", Valid: true},
		})

		// Add reward2
		isNew, collection, err := queries.AddPlushieToCollection(ctx, "user1", "alice", "coobubu", 2)
		if err != nil {
			t.Fatalf("AddPlushieToCollection failed: %v", err)
		}

		if !isNew {
			t.Error("expected isNew to be true for new reward")
		}

		expected := []int{1, 2}
		if !equalIntSlices(collection, expected) {
			t.Errorf("collection = %v, want %v", collection, expected)
		}
	})

	t.Run("adds plushie to collection - existing reward", func(t *testing.T) {
		queries.AddReward1(ctx, AddReward1Params{
			UserID:         sql.NullString{String: "user2", Valid: true},
			Username:       "bob",
			CollectionType: sql.NullString{String: "coobubu", Valid: true},
		})

		// Try to add reward1 again
		isNew, _, err := queries.AddPlushieToCollection(ctx, "user2", "bob", "coobubu", 1)
		if err != nil {
			t.Fatalf("AddPlushieToCollection failed: %v", err)
		}

		if isNew {
			t.Error("expected isNew to be false for existing reward")
		}
	})

	t.Run("resets collection", func(t *testing.T) {
		// Add all rewards
		for i := 1; i <= 8; i++ {
			queries.AddPlushieToCollection(ctx, "user3", "charlie", "coobubu", i)
		}

		// Reset
		err := queries.ResetUserCollection(ctx, ResetUserCollectionParams{
			UserID:         sql.NullString{String: "user3", Valid: true},
			CollectionType: sql.NullString{String: "coobubu", Valid: true},
		})
		if err != nil {
			t.Fatalf("ResetUserCollection failed: %v", err)
		}

		// Check collection is empty
		uc, err := queries.GetUserCollectionRow(ctx, GetUserCollectionRowParams{
			UserID:         sql.NullString{String: "user3", Valid: true},
			CollectionType: sql.NullString{String: "coobubu", Valid: true},
		})
		if err != nil {
			t.Fatalf("GetUserCollectionRow failed: %v", err)
		}

		collection := GetUserCollection(uc)
		if len(collection) != 0 {
			t.Errorf("expected empty collection, got %v", collection)
		}
	})
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestFormatStats(t *testing.T) {
	stats := Stat{
		ID:           "user123",
		Username:     "testuser",
		Strength:     5,
		Intelligence: 5,
		Charisma:     3,
		Luck:         3,
		Dexterity:    3,
		Penis:        3,
	}

	formatted := FormatStats("testuser", stats)
	expected := "testuser's stats: STR: 5 | INT: 5 | CHA: 3 | LUCK: 3 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

func TestFormatStatsNegative(t *testing.T) {
	stats := Stat{
		ID:           "user123",
		Username:     "testuser",
		Strength:     3,
		Intelligence: 3,
		Charisma:     9,
		Luck:         -2,
		Dexterity:    3,
		Penis:        3,
	}

	formatted := FormatStats("testuser", stats)
	expected := "testuser's stats: STR: 3 | INT: 3 | CHA: 9 | LUCK: -2 | DEX: 3 | PENIS: 3"

	if formatted != expected {
		t.Errorf("FormatStats() = %q, want %q", formatted, expected)
	}
}

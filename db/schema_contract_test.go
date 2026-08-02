package db

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"slices"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigratedSchemaContract(t *testing.T) {
	_, sqlDB := NewTestDB(t)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	assertSchemaContract(t, sqlDB)
}

func TestNodeCreatedDatabaseCompatibility(t *testing.T) {
	dbPath := os.Getenv("CHARSIBOT_NODE_DB_PATH")
	if dbPath == "" {
		t.Skip("CHARSIBOT_NODE_DB_PATH is not set")
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat Node-created database: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("Node-created database path %q is a directory", dbPath)
	}

	logger := slog.New(slog.DiscardHandler)
	sqlDB, err := Connect(t.Context(), dbPath, logger)
	if err != nil {
		t.Fatalf("open Node-created database with Go Connect: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close Node-created database: %v", err)
		}
	})

	assertSchemaContract(t, sqlDB)
}

func TestCreateGoDatabaseCompatibilityFixture(t *testing.T) {
	dbPath := os.Getenv("CHARSIBOT_GO_DB_PATH")
	if dbPath == "" {
		t.Skip("CHARSIBOT_GO_DB_PATH is not set")
	}
	if _, err := os.Stat(dbPath); !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			t.Fatalf("stat Go fixture path: %v", err)
		}
		t.Fatalf("refusing to overwrite existing Go fixture path %q", dbPath)
	}

	logger := slog.New(slog.DiscardHandler)
	sqlDB, err := Connect(t.Context(), dbPath, logger)
	if err != nil {
		t.Fatalf("create Go database fixture: %v", err)
	}
	assertSchemaContract(t, sqlDB)
	if _, err := sqlDB.ExecContext(t.Context(), `
INSERT INTO user_stats (user_id, username, stat_name, value)
VALUES ('go-viewer', 'GoViewer', 'strength', 17);
INSERT INTO user_plushies (user_id, username, series, key)
VALUES ('go-viewer', 'GoViewer', 'coobubu', 'cutey');
INSERT INTO viewer_activity (user_id, username, last_active_at)
VALUES ('go-viewer', 'GoViewer', '2026-08-02T01:02:03.456Z');`); err != nil {
		t.Fatalf("seed Go database fixture: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close Go database fixture: %v", err)
	}
}

func assertSchemaContract(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	assertApplicationTables(t, sqlDB)
	assertColumns(t, sqlDB, "user_stats", []columnContract{
		{name: "user_id", declaredType: "TEXT", notNull: true},
		{name: "username", declaredType: "TEXT", notNull: true},
		{name: "stat_name", declaredType: "TEXT", notNull: true},
		{name: "value", declaredType: "INTEGER", notNull: true, hasDefault: true, defaultValue: "3"},
	})
	assertColumns(t, sqlDB, "user_plushies", []columnContract{
		{name: "user_id", declaredType: "TEXT", notNull: true},
		{name: "username", declaredType: "TEXT", notNull: true},
		{name: "series", declaredType: "TEXT", notNull: true},
		{name: "key", declaredType: "TEXT", notNull: true},
	})
	assertColumns(t, sqlDB, "viewer_activity", []columnContract{
		{name: "user_id", declaredType: "TEXT"},
		{name: "username", declaredType: "TEXT", notNull: true},
		{name: "last_active_at", declaredType: "TEXT", notNull: true},
	})
	assertColumns(t, sqlDB, "goose_db_version", []columnContract{
		{name: "id", declaredType: "INTEGER"},
		{name: "version_id", declaredType: "INTEGER", notNull: true},
		{name: "is_applied", declaredType: "INTEGER", notNull: true},
		{name: "tstamp", declaredType: "TIMESTAMP", hasDefault: true, defaultValue: "datetime('now')"},
	})
	assertPrimaryKey(t, sqlDB, "user_stats", []string{"user_id", "stat_name"})
	assertPrimaryKey(t, sqlDB, "user_plushies", []string{"user_id", "series", "key"})
	assertPrimaryKey(t, sqlDB, "viewer_activity", []string{"user_id"})
	assertViewerActivityIndex(t, sqlDB)
	assertGooseMetadata(t, sqlDB)
}

type columnContract struct {
	name         string
	declaredType string
	notNull      bool
	hasDefault   bool
	defaultValue string
}

func assertColumns(t *testing.T, sqlDB *sql.DB, table string, expected []columnContract) {
	t.Helper()

	rows, err := sqlDB.QueryContext(t.Context(), `
SELECT name, type, "notnull", dflt_value
FROM pragma_table_info(?)
ORDER BY cid`, table)
	if err != nil {
		t.Fatalf("list %s columns: %v", table, err)
	}
	defer rows.Close()

	var actual []columnContract
	for rows.Next() {
		var column columnContract
		var notNull int
		var defaultValue sql.NullString
		if err := rows.Scan(&column.name, &column.declaredType, &notNull, &defaultValue); err != nil {
			t.Fatalf("scan %s column: %v", table, err)
		}
		column.notNull = notNull != 0
		column.hasDefault = defaultValue.Valid
		column.defaultValue = defaultValue.String
		actual = append(actual, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s columns: %v", table, err)
	}

	if len(actual) != len(expected) {
		t.Fatalf("%s has %d columns, want %d: %#v", table, len(actual), len(expected), actual)
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("%s column %d = %#v, want %#v", table, i, actual[i], expected[i])
		}
	}
}

func assertApplicationTables(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	rows, err := sqlDB.QueryContext(t.Context(), `
SELECT name
FROM sqlite_schema
WHERE type = 'table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		t.Fatalf("list application tables: %v", err)
	}
	defer rows.Close()

	var actual []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan application table: %v", err)
		}
		actual = append(actual, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate application tables: %v", err)
	}

	expected := []string{"goose_db_version", "user_plushies", "user_stats", "viewer_activity"}
	slices.Sort(actual)
	slices.Sort(expected)
	if !slices.Equal(actual, expected) {
		t.Fatalf("application tables = %v, want %v", actual, expected)
	}

	obsolete := []string{
		"oauth_tokens",
		"bot_token",
		"stats",
		"user_collections",
		"blind_box_series",
		"blind_box_plushies",
		"stat_definitions",
	}
	for _, name := range obsolete {
		var count int
		if err := sqlDB.QueryRowContext(t.Context(), `
SELECT COUNT(*) FROM sqlite_schema WHERE type = 'table' AND name = ?`, name).Scan(&count); err != nil {
			t.Fatalf("check obsolete table %q: %v", name, err)
		}
		if count != 0 {
			t.Errorf("obsolete table %q is present", name)
		}
	}
}

func assertPrimaryKey(t *testing.T, sqlDB *sql.DB, table string, expected []string) {
	t.Helper()

	rows, err := sqlDB.QueryContext(t.Context(), `
SELECT name
FROM pragma_table_info(?)
WHERE pk > 0
ORDER BY pk`, table)
	if err != nil {
		t.Fatalf("list %s primary key: %v", table, err)
	}
	defer rows.Close()

	var actual []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan %s primary key: %v", table, err)
		}
		actual = append(actual, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s primary key: %v", table, err)
	}
	if !slices.Equal(actual, expected) {
		t.Errorf("%s primary key = %v, want %v", table, actual, expected)
	}
}

func assertViewerActivityIndex(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	const indexName = "viewer_activity_last_active_at_idx"
	var table string
	if err := sqlDB.QueryRowContext(t.Context(), `
SELECT tbl_name
FROM sqlite_schema
WHERE type = 'index' AND name = ?`, indexName).Scan(&table); err != nil {
		t.Fatalf("find %s: %v", indexName, err)
	}
	if table != "viewer_activity" {
		t.Errorf("%s table = %q, want viewer_activity", indexName, table)
	}

	rows, err := sqlDB.QueryContext(t.Context(), `
SELECT name
FROM pragma_index_info(?)
ORDER BY seqno`, indexName)
	if err != nil {
		t.Fatalf("list %s columns: %v", indexName, err)
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan %s column: %v", indexName, err)
		}
		columns = append(columns, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate %s columns: %v", indexName, err)
	}
	if expected := []string{"last_active_at"}; !slices.Equal(columns, expected) {
		t.Errorf("%s columns = %v, want %v", indexName, columns, expected)
	}
}

func assertGooseMetadata(t *testing.T, sqlDB *sql.DB) {
	t.Helper()

	assertPrimaryKey(t, sqlDB, "goose_db_version", []string{"id"})

	rows, err := sqlDB.QueryContext(t.Context(), `
SELECT version_id, is_applied FROM goose_db_version`)
	if err != nil {
		t.Fatalf("list Goose migration history: %v", err)
	}
	defer rows.Close()

	actual := make(map[int64]bool)
	rowCount := 0
	for rows.Next() {
		var version int64
		var applied bool
		if err := rows.Scan(&version, &applied); err != nil {
			t.Fatalf("scan Goose migration history: %v", err)
		}
		if _, duplicate := actual[version]; duplicate {
			t.Fatalf("Goose migration history contains duplicate version %d", version)
		}
		actual[version] = applied
		rowCount++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate Goose migration history: %v", err)
	}

	expected := make(map[int64]bool, 8)
	for version := range int64(8) {
		expected[version] = true
	}
	if rowCount != len(expected) {
		t.Fatalf("Goose migration history has %d rows, want %d: %v", rowCount, len(expected), actual)
	}
	for version, wantApplied := range expected {
		applied, found := actual[version]
		if !found {
			t.Errorf("Goose migration history is missing version %d", version)
			continue
		}
		if applied != wantApplied {
			t.Errorf("Goose migration %d applied = %t, want %t", version, applied, wantApplied)
		}
	}

	var currentVersion int64
	if err := sqlDB.QueryRowContext(t.Context(), `
SELECT MAX(version_id) FROM goose_db_version WHERE is_applied = 1`).Scan(&currentVersion); err != nil {
		t.Fatalf("read current Goose version: %v", err)
	}
	if currentVersion != 7 {
		t.Errorf("current Goose version = %d, want 7", currentVersion)
	}
}

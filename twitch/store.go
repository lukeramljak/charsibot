package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func (s *Store) InitTokenSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS oauth_tokens (
			token_type TEXT PRIMARY KEY,
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *Store) GetTokens(ctx context.Context, tokenType string) (*Tokens, error) {
	var tokens Tokens
	err := s.db.QueryRowContext(ctx,
		"SELECT access_token, refresh_token FROM oauth_tokens WHERE token_type = ?",
		tokenType,
	).Scan(&tokens.AccessToken, &tokens.RefreshToken)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &tokens, err
}

func (s *Store) SaveTokens(ctx context.Context, tokenType, accessToken, refreshToken string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO oauth_tokens (token_type, access_token, refresh_token, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(token_type) DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			updated_at = CURRENT_TIMESTAMP
	`, tokenType, accessToken, refreshToken)
	return err
}

type Stats struct {
	ID           string
	Username     string
	Strength     int
	Intelligence int
	Charisma     int
	Luck         int
	Dexterity    int
	Penis        int
}

type Stat struct {
	Display string
	Column  string
}

var statList = []Stat{
	{Display: "Strength", Column: "strength"},
	{Display: "Intelligence", Column: "intelligence"},
	{Display: "Charisma", Column: "charisma"},
	{Display: "Luck", Column: "luck"},
	{Display: "Dexterity", Column: "dexterity"},
	{Display: "Penis", Column: "penis"},
}

var validStatColumns = map[string]bool{
	"strength":     true,
	"intelligence": true,
	"charisma":     true,
	"luck":         true,
	"dexterity":    true,
	"penis":        true,
}

func (s *Store) GetStats(ctx context.Context, userID, username string) (*Stats, error) {
	var st Stats
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO stats (id, username) VALUES (?, ?)
		ON CONFLICT(id) DO UPDATE SET username=excluded.username
		RETURNING id, username, strength, intelligence, charisma, luck, dexterity, penis
	`, userID, username).Scan(&st.ID, &st.Username, &st.Strength, &st.Intelligence, &st.Charisma, &st.Luck, &st.Dexterity, &st.Penis)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *Store) ModifyStat(ctx context.Context, userID, username, column string, delta int) error {
	if !validStatColumns[column] {
		return fmt.Errorf("invalid stat column: %s", column)
	}

	res, err := s.db.ExecContext(ctx, "UPDATE stats SET "+column+" = "+column+" + ? WHERE id = ?", delta, userID)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// Create the user first
		if _, err := s.db.ExecContext(ctx, "INSERT INTO stats (id, username) VALUES (?, ?)", userID, username); err != nil {
			return err
		}
		// Try the update again
		_, err = s.db.ExecContext(ctx, "UPDATE stats SET "+column+" = "+column+" + ? WHERE id = ?", delta, userID)
	}
	return err
}

func FormatStats(username string, st *Stats) string {
	return fmt.Sprintf("%s's stats: STR: %d | INT: %d | CHA: %d | LUCK: %d | DEX: %d | PENIS: %d",
		username, st.Strength, st.Intelligence, st.Charisma, st.Luck, st.Dexterity, st.Penis)
}

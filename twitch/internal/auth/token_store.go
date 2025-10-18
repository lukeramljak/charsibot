package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

type TokenType string

const (
	TokenTypeStreamer TokenType = "streamer"
	TokenTypeBot      TokenType = "bot"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type TokenStore struct {
	db *sql.DB
}

func NewTokenStore(db *sql.DB) *TokenStore {
	return &TokenStore{db: db}
}

// InitSchema creates the tokens table if it doesn't exist
func (s *TokenStore) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS oauth_tokens (
			token_type TEXT PRIMARY KEY,
			access_token TEXT NOT NULL,
			refresh_token TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create oauth_tokens table: %w", err)
	}
	return nil
}

// GetTokens retrieves stored tokens for the given type
func (s *TokenStore) GetTokens(ctx context.Context, tokenType TokenType) (*Tokens, error) {
	var tokens Tokens
	err := s.db.QueryRowContext(ctx,
		"SELECT access_token, refresh_token FROM oauth_tokens WHERE token_type = ?",
		string(tokenType),
	).Scan(&tokens.AccessToken, &tokens.RefreshToken)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tokens: %w", err)
	}

	return &tokens, nil
}

// SaveTokens stores or updates tokens for the given type
func (s *TokenStore) SaveTokens(ctx context.Context, tokenType TokenType, accessToken, refreshToken string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO oauth_tokens (token_type, access_token, refresh_token, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(token_type) DO UPDATE SET
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			updated_at = CURRENT_TIMESTAMP
	`, string(tokenType), accessToken, refreshToken)

	if err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}

	slog.Info("Tokens saved to database", "token_type", tokenType)
	return nil
}

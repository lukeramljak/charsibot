package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

func (tc *TwitchClient) RegisterChatCommand(command string, handler func(username, message string) error) {
	tc.handlers[command] = handler
}

func (tc *TwitchClient) handleChatCommand(username, text string) {
	if handler, exists := tc.handlers[text]; exists {
		if err := handler(username, text); err != nil {
			slog.Error("Handler error", "command", text, "error", err)
		}
	}
}

func (tc *TwitchClient) handleStats(username, message string) error {
	var stats Stats
	err := tc.db.QueryRow("SELECT * FROM stats WHERE username = ?", username).
		Scan(
			&stats.Username,
			&stats.Strength,
			&stats.Intelligence,
			&stats.Charisma,
			&stats.Luck,
			&stats.Dexterity,
			&stats.Penis)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = tc.db.QueryRow("INSERT INTO stats (username) VALUES (?) RETURNING *", username).
				Scan(
					&stats.Username,
					&stats.Strength,
					&stats.Intelligence,
					&stats.Charisma,
					&stats.Luck,
					&stats.Dexterity,
					&stats.Penis)
			if err != nil {
				return fmt.Errorf("failed to create stats for user %s: %w", username, err)
			}
			slog.Info("Created new stats for user", "username", username)
		} else {
			return fmt.Errorf("failed to query stats for user %s: %w", username, err)
		}
	}

	statsMessage := fmt.Sprintf(
		"%s's stats: STR: %d | INT: %d | CHA: %d | LUCK: %d | DEX: %d | PENIS: %d",
		username,
		stats.Strength,
		stats.Intelligence,
		stats.Charisma,
		stats.Luck,
		stats.Dexterity,
		stats.Penis)

	return tc.SendChatMessage(statsMessage)
}

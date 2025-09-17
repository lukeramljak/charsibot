package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

func (tc *TwitchClient) RegisterChatCommand(command string, handler func(event *Event) error) {
	tc.handlers[command] = handler
}

func (tc *TwitchClient) handleChatCommand(event *Event) {
	if handler, exists := tc.handlers[event.Message.Text]; exists {
		if err := handler(event); err != nil {
			slog.Error("Handler error", "command", event.Message.Text, "error", err)
		}
	}
}

func (tc *TwitchClient) handleStats(event *Event) error {
	var stats Stats
	err := tc.db.QueryRow("SELECT * FROM stats WHERE id = ?", event.ChatterUserID).
		Scan(
			&stats.ID,
			&stats.Username,
			&stats.Strength,
			&stats.Intelligence,
			&stats.Charisma,
			&stats.Luck,
			&stats.Dexterity,
			&stats.Penis)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = tc.db.QueryRow("INSERT INTO stats (id, username) VALUES (?, ?) RETURNING *", event.ChatterUserID, event.ChatterUserLogin).
				Scan(
					&stats.ID,
					&stats.Username,
					&stats.Strength,
					&stats.Intelligence,
					&stats.Charisma,
					&stats.Luck,
					&stats.Dexterity,
					&stats.Penis)
			if err != nil {
				return fmt.Errorf("failed to create stats for user %s: %w", event.ChatterUserID, err)
			}
			slog.Info("Created new stats for user", "username", event.ChatterUserLogin)
		} else {
			return fmt.Errorf("failed to query stats for user %s: %w", event.ChatterUserLogin, err)
		}
	}

	statsMessage := fmt.Sprintf(
		"%s's stats: STR: %d | INT: %d | CHA: %d | LUCK: %d | DEX: %d | PENIS: %d",
		event.ChatterUserLogin,
		stats.Strength,
		stats.Intelligence,
		stats.Charisma,
		stats.Luck,
		stats.Dexterity,
		stats.Penis)

	return tc.SendChatMessage(statsMessage)
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Viewer struct {
	UserID       string
	Username     string
	LastActiveAt *time.Time
}

func (q *Queries) RecordViewerActivity(ctx context.Context, userID, username string, at time.Time) error {
	_, err := q.db.ExecContext(ctx, `
INSERT INTO viewer_activity (user_id, username, last_active_at)
VALUES (?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET username = excluded.username, last_active_at = excluded.last_active_at`, userID, username, at.UTC().Format(time.RFC3339Nano))
	return err
}

func (q *Queries) ListViewers(ctx context.Context) ([]Viewer, error) {
	rows, err := q.db.QueryContext(ctx, `
SELECT user_id, MAX(username), MAX(last_active_at) FROM (
  SELECT user_id, username, last_active_at FROM viewer_activity
  UNION ALL SELECT user_id, username, NULL FROM user_stats
  UNION ALL SELECT user_id, username, NULL FROM user_plushies
) GROUP BY user_id ORDER BY MAX(username) COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	viewers := []Viewer{}
	for rows.Next() {
		var viewer Viewer
		var lastActive sql.NullString
		if err := rows.Scan(&viewer.UserID, &viewer.Username, &lastActive); err != nil {
			return nil, err
		}
		if lastActive.Valid {
			value, err := time.Parse(time.RFC3339Nano, lastActive.String)
			if err != nil {
				return nil, fmt.Errorf("parse viewer activity: %w", err)
			}
			viewer.LastActiveAt = &value
		}
		viewers = append(viewers, viewer)
	}
	return viewers, rows.Err()
}

func (q *Queries) GetViewer(ctx context.Context, userID string) (Viewer, error) {
	var viewer Viewer
	var lastActive sql.NullString
	err := q.db.QueryRowContext(ctx, `
SELECT user_id, username, last_active_at FROM viewer_activity WHERE user_id = ?
UNION
SELECT user_id, username, NULL FROM user_stats WHERE user_id = ?
UNION
SELECT user_id, username, NULL FROM user_plushies WHERE user_id = ?
LIMIT 1`, userID, userID, userID).Scan(&viewer.UserID, &viewer.Username, &lastActive)
	if err != nil {
		return Viewer{}, err
	}
	if lastActive.Valid {
		value, err := time.Parse(time.RFC3339Nano, lastActive.String)
		if err != nil {
			return Viewer{}, fmt.Errorf("parse viewer activity: %w", err)
		}
		viewer.LastActiveAt = &value
	}
	return viewer, nil
}

func (q *Queries) DeleteViewer(ctx context.Context, userID string) error {
	for _, query := range []string{"DELETE FROM viewer_activity WHERE user_id = ?", "DELETE FROM user_stats WHERE user_id = ?", "DELETE FROM user_plushies WHERE user_id = ?"} {
		if _, err := q.db.ExecContext(ctx, query, userID); err != nil {
			return err
		}
	}
	return nil
}

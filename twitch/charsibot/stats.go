package charsibot

import (
	"context"
	"fmt"
	"strings"

	"github.com/lukeramljak/charsibot/twitch/db"
)

// GetOrCreateStats ensures stat rows exist for a user then returns their stats.
func GetOrCreateStats(ctx context.Context, q *db.Queries, userID, username string) ([]db.GetUserStatsRow, error) {
	if err := q.EnsureUserStats(ctx, db.EnsureUserStatsParams{
		UserID:   userID,
		Username: username,
		UserID_2: userID,
	}); err != nil {
		return nil, fmt.Errorf("ensure stats: %w", err)
	}
	if err := q.UpdateUsername(ctx, db.UpdateUsernameParams{
		Username: username,
		UserID:   userID,
	}); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}
	return q.GetUserStats(ctx, userID)
}

func FormatStats(username string, stats []db.GetUserStatsRow) string {
	parts := make([]string, len(stats))
	for i, s := range stats {
		parts[i] = fmt.Sprintf("%s: %d", s.ShortName, s.Value)
	}
	return fmt.Sprintf("%s's stats: %s", username, strings.Join(parts, " | "))
}

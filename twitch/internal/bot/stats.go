package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/lukeramljak/charsibot/twitch/internal/store"
)

// GetOrCreateStats ensures stat rows exist for a user then returns their stats.
func GetOrCreateStats(ctx context.Context, q *store.Queries, userID, username string) ([]store.GetUserStatsRow, error) {
	if err := q.EnsureUserStats(ctx, store.EnsureUserStatsParams{
		UserID:   userID,
		Username: username,
		UserID_2: userID,
	}); err != nil {
		return nil, fmt.Errorf("ensure stats: %w", err)
	}
	if err := q.UpdateUsername(ctx, store.UpdateUsernameParams{
		Username: username,
		UserID:   userID,
	}); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}
	return q.GetUserStats(ctx, userID)
}

func FormatStats(username string, stats []store.GetUserStatsRow) string {
	parts := make([]string, len(stats))
	for i, s := range stats {
		parts[i] = fmt.Sprintf("%s: %d", s.ShortName, s.Value)
	}
	return fmt.Sprintf("%s's stats: %s", username, strings.Join(parts, " | "))
}

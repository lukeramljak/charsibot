package stats

import (
	"fmt"
	"strings"

	"github.com/lukeramljak/charsibot/twitch/db"
)

// FormatStats formats a user's stats as a human-readable chat message.
func FormatStats(username string, stats []db.GetUserStatsRow) string {
	parts := make([]string, len(stats))
	for i, s := range stats {
		parts[i] = fmt.Sprintf("%s: %d", s.ShortName, s.Value)
	}
	return fmt.Sprintf("%s's stats: %s", username, strings.Join(parts, " | "))
}

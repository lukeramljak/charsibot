package stats

import (
	"fmt"
	"strings"

	"github.com/lukeramljak/charsibot/db"
)

// FormatStats formats a user's stats as a human-readable chat message.
func FormatStats(username string, stats []db.GetUserStatsRow) string {
	if len(stats) == 0 {
		return fmt.Sprintf("No stats found for %s", username)
	}

	parts := make([]string, len(stats))
	for i, s := range stats {
		parts[i] = fmt.Sprintf("%s: %d", s.ShortName, s.Value)
	}

	return fmt.Sprintf("%s's stats: %s", username, strings.Join(parts, " | "))
}

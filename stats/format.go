package stats

import (
	"fmt"
	"strings"
)

// FormatStats formats a user's stats as a human-readable chat message.
func FormatStats(username string, stats []UserStat) string {
	if len(stats) == 0 {
		return fmt.Sprintf("No stats found for %s", username)
	}

	parts := make([]string, 0, len(stats))
	for _, stat := range stats {
		parts = append(parts, fmt.Sprintf("%s: %d", stat.ShortName, stat.Value))
	}

	return fmt.Sprintf("%s's stats: %s", username, strings.Join(parts, " | "))
}

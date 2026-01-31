package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/store"
	"github.com/nicklaw5/helix/v2"
)

// StatsCommand displays a user's stats
type StatsCommand struct{}

func NewStatsCommand() *StatsCommand {
	return &StatsCommand{}
}

func (c *StatsCommand) ModeratorOnly() bool {
	return false
}

func (c *StatsCommand) ShouldTrigger(command string) bool {
	return command == "stats"
}

func (c *StatsCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	id := event.ChatterUserId
	username := event.ChatterUserName

	stats, err := b.Store().GetStats(b.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// User doesn't have stats yet, create default
			if err := b.Store().UpsertStatsUser(b.Context(), store.UpsertStatsUserParams{
				ID:       id,
				Username: username,
			}); err != nil {
				slog.Error("failed to create stats for user", "err", err, "user", username)
				return
			}
			// Fetch again
			stats, err = b.Store().GetStats(b.Context(), id)
			if err != nil {
				slog.Error("failed to get stats after creation", "err", err, "user", username)
				return
			}
		} else {
			slog.Error("failed to get stats", "err", err, "user", username)
			return
		}
	}

	message := store.FormatStats(username, stats)
	if err := b.SendMessage(bot.SendMessageParams{
		Message:              message,
		ReplyParentMessageID: event.MessageId,
	}); err != nil {
		slog.Error("failed to send stats message", "err", err)
	}
}

// LeaderboardCommand displays the stats leaderboard
type LeaderboardCommand struct{}

func NewLeaderboardCommand() *LeaderboardCommand {
	return &LeaderboardCommand{}
}

func (c *LeaderboardCommand) ModeratorOnly() bool {
	return false
}

func (c *LeaderboardCommand) ShouldTrigger(command string) bool {
	return command == "leaderboard"
}

func (c *LeaderboardCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	stats, err := b.Store().GetStatLeaderboard(b.Context())
	if err != nil {
		slog.Error("failed to get leaderboard", "err", err)
		return
	}

	statMap := map[string]struct {
		username string
		value    int64
	}{
		"STR":   {username: stats.TopStrengthUsername, value: stats.TopStrengthValue},
		"INT":   {username: stats.TopIntelligenceUsername, value: stats.TopIntelligenceValue},
		"CHA":   {username: stats.TopCharismaUsername, value: stats.TopCharismaValue},
		"LUCK":  {username: stats.TopLuckUsername, value: stats.TopLuckValue},
		"DEX":   {username: stats.TopDexterityUsername, value: stats.TopDexterityValue},
		"PENIS": {username: stats.TopPenisUsername, value: stats.TopPenisValue},
	}

	emojiMap := map[string]string{
		"STR":   "💪",
		"INT":   "🧠",
		"CHA":   "✨",
		"LUCK":  "🍀",
		"DEX":   "🤸",
		"PENIS": "🍆",
	}

	order := []string{"STR", "INT", "CHA", "LUCK", "DEX", "PENIS"}
	parts := []string{}
	for _, label := range order {
		emoji := emojiMap[label]
		stat := statMap[label]
		parts = append(parts, fmt.Sprintf("%s %s(%d)", emoji, stat.username, stat.value))
	}

	message := "Stats leaderboard: "
	for i, part := range parts {
		if i > 0 {
			message += " | "
		}
		message += part
	}

	if err := b.SendMessage(bot.SendMessageParams{
		Message: message,
	}); err != nil {
		slog.Error("failed to send leaderboard message", "err", err)
	}
}

// ModifyStatCommand allows moderators to modify user stats
type ModifyStatCommand struct{}

func NewModifyStatCommand() *ModifyStatCommand {
	return &ModifyStatCommand{}
}

func (c *ModifyStatCommand) ModeratorOnly() bool {
	return true
}

func (c *ModifyStatCommand) ShouldTrigger(command string) bool {
	return command == "addstat" || command == "rmstat"
}

func (c *ModifyStatCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	parts := strings.Fields(strings.ToLower(event.Message.Text))
	if len(parts) < 1 {
		return
	}

	command := parts[0]
	isRemove := command == "!rmstat"

	mentionedLogin, statColumn, amount, err := parseModifyStatCommand(event.Message.Text)
	if err != nil {
		slog.Warn("invalid modify stat command", "err", err, "msg", event.Message.Text)
		if err := b.SendMessage(bot.SendMessageParams{
			Message: err.Error(),
		}); err != nil {
			slog.Error("failed to send error message", "err", err)
		}
		return
	}

	// Look up the mentioned user using the bot's helix client
	mentionedUser, err := b.HelixClient().GetUsers(&helix.UsersParams{
		Logins: []string{mentionedLogin},
	})
	if err != nil || len(mentionedUser.Data.Users) == 0 {
		slog.Error("failed to find user", "login", mentionedLogin, "err", err)
		if err := b.SendMessage(bot.SendMessageParams{
			Message: "Failed to find user",
		}); err != nil {
			slog.Error("failed to send error message", "err", err)
		}
		return
	}

	user := mentionedUser.Data.Users[0]

	// Modify the stat
	finalAmount := amount
	if isRemove {
		finalAmount = -amount
	}

	if err := b.Store().ModifyStat(b.Context(), user.ID, user.Login, statColumn, finalAmount); err != nil {
		slog.Error("failed to modify stat", "err", err, "user", user.Login)
		return
	}

	// Get updated stats
	stats, err := b.Store().GetStats(b.Context(), user.ID)
	if err != nil {
		slog.Error("failed to get stats", "err", err, "user", user.Login)
		return
	}

	message := store.FormatStats(user.Login, stats)
	if err := b.SendMessage(bot.SendMessageParams{
		Message: message,
	}); err != nil {
		slog.Error("failed to send stats message", "err", err)
	}
}

var mentionRegex = regexp.MustCompile(`@(\w+)`)

// parseModifyStatCommand parses commands like: !addstat @user stat amount OR !rmstat @user stat amount
func parseModifyStatCommand(text string) (string, string, int64, error) {
	parts := strings.Fields(text)
	if len(parts) < 4 {
		return "", "", 0, fmt.Errorf("expected format: !addstat/!rmstat @user stat amount")
	}

	// Find mention
	matches := mentionRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return "", "", 0, fmt.Errorf("no user mention found")
	}

	mentionedLogin := strings.ToLower(matches[1])
	statColumn := parts[2]
	amountStr := parts[3]

	if statColumn == "" || amountStr == "" {
		return "", "", 0, fmt.Errorf("expected 'stat amount'")
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid number")
	}

	return mentionedLogin, statColumn, amount, nil
}

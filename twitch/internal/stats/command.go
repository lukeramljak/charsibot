package stats

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/server"
	"github.com/lukeramljak/charsibot/internal/store"
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
	b.SendMessage(bot.SendMessageParams{
		Message:              message,
		ReplyParentMessageID: event.MessageId,
	})
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
		Username string `json:"username"`
		Value    int64  `json:"value"`
	}{
		"STR":   {Username: stats.TopStrengthUsername, Value: stats.TopStrengthValue},
		"INT":   {Username: stats.TopIntelligenceUsername, Value: stats.TopIntelligenceValue},
		"CHA":   {Username: stats.TopCharismaUsername, Value: stats.TopCharismaValue},
		"LUCK":  {Username: stats.TopLuckUsername, Value: stats.TopLuckValue},
		"DEX":   {Username: stats.TopDexterityUsername, Value: stats.TopDexterityValue},
		"PENIS": {Username: stats.TopPenisUsername, Value: stats.TopPenisValue},
	}

	b.BroadcastOverlayEvent(server.OverlayEvent{
		Type: server.EventTypeLeaderboard,
		Data: statMap,
	})
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

	mentionedUser, err := extractMentionedUserFromFragments(event.Message.Fragments)
	if err != nil {
		b.SendMessage(bot.SendMessageParams{
			Message:              "Missing user mention",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	statColumn, amount, err := parseModifyStatCommand(event.Message.Text)
	if err != nil {
		slog.Warn("invalid modify stat command", "err", err, "msg", event.Message.Text)
		b.SendMessage(bot.SendMessageParams{
			Message: err.Error(),
		})
		return
	}

	// Modify the stat
	finalAmount := amount
	if isRemove {
		finalAmount = -amount
	}

	if err := b.Store().ModifyStat(b.Context(), mentionedUser.UserID, mentionedUser.UserLogin, statColumn, finalAmount); err != nil {
		slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
		b.SendMessage(bot.SendMessageParams{
			Message:              "Failed to update stats",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	// Get updated stats
	stats, err := b.Store().GetStats(b.Context(), mentionedUser.UserID)
	if err != nil {
		slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
		b.SendMessage(bot.SendMessageParams{
			Message:              "Failed to get stats",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	message := store.FormatStats(mentionedUser.UserLogin, stats)
	b.SendMessage(bot.SendMessageParams{
		Message: message,
	})
}

// ExplodeCommand sets a user's penis stat to -1000.
type ExplodeCommand struct{}

func NewExplodeCommand() *ExplodeCommand {
	return &ExplodeCommand{}
}

func (c *ExplodeCommand) ModeratorOnly() bool {
	return true
}

func (c *ExplodeCommand) ShouldTrigger(command string) bool {
	return command == "explode"
}

func (c *ExplodeCommand) Execute(b *bot.Bot, event twitch.EventChannelChatMessage) {
	fragments := event.Message.Fragments

	mentionedUser, err := extractMentionedUserFromFragments(fragments)
	if err != nil {
		b.SendMessage(bot.SendMessageParams{
			Message:              "Missing user mention",
			ReplyParentMessageID: event.MessageId,
		})
	}

	if err := b.Store().ModifyStat(b.Context(), mentionedUser.UserID, mentionedUser.UserLogin, "penis", -1003); err != nil {
		slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
		b.SendMessage(bot.SendMessageParams{
			Message:              "Failed to update stats",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	stats, err := b.Store().GetStats(b.Context(), mentionedUser.UserID)
	if err != nil {
		slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
		b.SendMessage(bot.SendMessageParams{
			Message:              "Failed to get updated stats",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	message := store.FormatStats(mentionedUser.UserLogin, stats)
	b.SendMessage(bot.SendMessageParams{
		Message:              message,
		ReplyParentMessageID: event.MessageId,
	})
}

// extractMentionedUserFromFragments extracts the user of the first mention from message fragments.
func extractMentionedUserFromFragments(fragments []twitch.ChatMessageFragment) (*twitch.ChatMessageFragmentMention, error) {
	for _, fragment := range fragments {
		if fragment.Type == "mention" && fragment.Mention != nil {
			return fragment.Mention, nil
		}
	}
	return nil, fmt.Errorf("no user mention found")
}

// parseModifyStatCommand parses commands like: !addstat @user stat amount OR !rmstat @user stat amount
func parseModifyStatCommand(text string) (string, int64, error) {
	parts := strings.Fields(text)
	if len(parts) < 4 {
		return "", 0, fmt.Errorf("expected format: !addstat/!rmstat @user stat amount")
	}

	statColumn := parts[2]
	amountStr := parts[3]

	if statColumn == "" || amountStr == "" {
		return "", 0, fmt.Errorf("expected 'stat amount'")
	}

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid number")
	}

	return statColumn, amount, nil
}

package stats

import (
	"database/sql"
	"errors"
	"log/slog"
	"math/rand/v2"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/store"
)

type statInfo struct {
	Display string
	Column  string
}

var statList = []statInfo{
	{Display: "Strength", Column: "strength"},
	{Display: "Intelligence", Column: "intelligence"},
	{Display: "Charisma", Column: "charisma"},
	{Display: "Luck", Column: "luck"},
	{Display: "Dexterity", Column: "dexterity"},
	{Display: "Penis", Column: "penis"},
}

func getRandomStat() statInfo {
	return statList[rand.IntN(len(statList))]
}

func getRandomStatDelta() int64 {
	if rand.Float64() < 0.05 {
		return -1
	}
	return 1
}

// PotionRedemption handles the "Drink a Potion" channel point redemption
type PotionRedemption struct{}

func NewPotionRedemption() *PotionRedemption {
	return &PotionRedemption{}
}

func (r *PotionRedemption) ShouldTrigger(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool {
	return event.Reward.Title == "Drink a Potion"
}

func (r *PotionRedemption) Execute(b *bot.Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	stat := getRandomStat()
	delta := getRandomStatDelta()
	outcome := "lost"
	if delta > 0 {
		outcome = "gained"
	}

	userId := event.UserID
	username := event.UserName

	if err := b.Store().ModifyStat(b.Context(), userId, username, stat.Column, delta); err != nil {
		slog.Error("failed to modify stat", "err", err, "user", username)
		return
	}

	message := "A shifty looking merchant hands " + username + " a glittering potion. Without hesitation, they sink the whole drink. " +
		username + " " + outcome + " " + stat.Display

	if err := b.SendMessage(bot.SendMessageParams{
		Message: message,
	}); err != nil {
		slog.Error("failed to send potion message", "err", err)
		return
	}

	stats, err := b.Store().GetStats(b.Context(), userId)
	if err != nil {
		slog.Error("failed to get stats", "err", err, "user", username)
		return
	}

	if err := b.SendMessage(bot.SendMessageParams{
		Message: store.FormatStats(username, stats),
	}); err != nil {
		slog.Error("failed to send stats message", "err", err)
	}
}

// TemptDiceRedemption handles the "Tempt the Dice" channel point redemption
type TemptDiceRedemption struct{}

func NewTemptDiceRedemption() *TemptDiceRedemption {
	return &TemptDiceRedemption{}
}

func (r *TemptDiceRedemption) ShouldTrigger(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool {
	return event.Reward.Title == "Tempt the Dice"
}

func (r *TemptDiceRedemption) Execute(b *bot.Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	userId := event.UserID
	username := event.UserName

	// Send initiative message
	if err := b.SendMessage(bot.SendMessageParams{
		Message: username + " has rolled with initiative.",
	}); err != nil {
		slog.Error("failed to send initiative message", "err", err)
		return
	}

	// Get stats
	stats, err := b.Store().GetStats(b.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create default stats
			if err := b.Store().UpsertStatsUser(b.Context(), store.UpsertStatsUserParams{
				ID:       userId,
				Username: username,
			}); err != nil {
				slog.Error("failed to create stats", "err", err, "user", username)
				return
			}
			stats, err = b.Store().GetStats(b.Context(), userId)
			if err != nil {
				slog.Error("failed to get stats after creation", "err", err, "user", username)
				return
			}
		} else {
			slog.Error("failed to get stats", "err", err, "user", username)
			return
		}
	}

	// Send stats
	if err := b.SendMessage(bot.SendMessageParams{
		Message: store.FormatStats(username, stats),
	}); err != nil {
		slog.Error("failed to send stats message", "err", err)
	}
}

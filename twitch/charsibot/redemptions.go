package charsibot

import (
	"log/slog"
	"math/rand/v2"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/twitch/db"
)

// RedemptionFunc handles a channel point redemption event.
type RedemptionFunc func(b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd)

// Redemptions returns the full map of channel point redemptions keyed by reward title.
func Redemptions(seriesConfigs []SeriesConfig) map[string]RedemptionFunc {
	redemptions := map[string]RedemptionFunc{
		"Drink a Potion": func(b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			const (
				negativePercent = 5
				percentMax      = 100
			)

			userID := event.UserID
			username := event.UserName

			if _, err := GetOrCreateStats(b.ctx, b.store, userID, username); err != nil {
				slog.Error("failed to get or create stats", "err", err, "user", username)
				return
			}

			stat, err := b.store.GetRandomStatDefinition(b.ctx)
			if err != nil {
				slog.Error("failed to get random stat definition", "err", err)
				return
			}

			delta := int64(1)
			outcome := "gained"
			roll := rand.IntN(percentMax)
			if roll < negativePercent {
				delta = -1
				outcome = "lost"
			}

			if err = b.store.ModifyStatValue(b.ctx, db.ModifyStatValueParams{
				Value:    delta,
				UserID:   userID,
				StatName: stat.Name,
			}); err != nil {
				slog.Error("failed to modify stat", "err", err, "user", username)
				return
			}

			b.SendMessage(SendMessageParams{
				Message: "A shifty looking merchant hands " + username + " a glittering potion. Without hesitation, they sink the whole drink. " +
					username + " " + outcome + " " + stat.LongName,
			})

			stats, err := b.store.GetUserStats(b.ctx, userID)
			if err != nil {
				slog.Error("failed to get stats", "err", err, "user", username)
				return
			}
			b.SendMessage(SendMessageParams{Message: FormatStats(username, stats)})
		},

		"Tempt the Dice": func(b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			userID := event.UserID
			username := event.UserName

			b.SendMessage(SendMessageParams{
				Message: username + " has rolled with initiative.",
			})

			stats, err := GetOrCreateStats(b.ctx, b.store, userID, username)
			if err != nil {
				slog.Error("failed to get stats", "err", err, "user", username)
				return
			}
			b.SendMessage(SendMessageParams{Message: FormatStats(username, stats)})
		},
	}

	for _, cfg := range seriesConfigs {
		redemptions[cfg.RedemptionTitle] = func(b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			RedeemBlindBox(b, event.UserID, event.UserName, cfg)
		}
	}

	return redemptions
}

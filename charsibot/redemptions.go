package charsibot

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/server"
	"github.com/lukeramljak/charsibot/stats"
)

// RedemptionFunc handles a channel point redemption event.
type RedemptionFunc func(ctx context.Context, b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd)

// Redemptions returns the full map of channel point redemptions keyed by reward title.
func Redemptions(seriesConfigs []blindbox.SeriesConfig) map[string]RedemptionFunc {
	const rewardDrinkAPotion = "Drink a Potion"

	redemptions := map[string]RedemptionFunc{
		rewardDrinkAPotion: func(ctx context.Context, b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			const (
				negativePercent = 5
				percentMax      = 100
			)

			userID := event.UserID
			username := event.UserName

			if _, err := b.statsService.GetOrCreateStats(ctx, userID, username); err != nil {
				b.logger.Error("failed to get or create stats", "err", err, "user", username)
				return
			}

			stat, err := b.statsService.GetRandomStatDefinition(ctx)
			if err != nil {
				b.logger.Error("failed to get random stat definition", "err", err)
				return
			}

			delta := int64(1)
			outcome := "gained"
			roll := rand.IntN(percentMax)
			if roll < negativePercent {
				delta = -1
				outcome = "lost"
			}

			if err = b.statsService.ModifyStatValue(ctx, userID, stat.Name, delta); err != nil {
				b.logger.Error("failed to modify stat", "err", err, "user", username)
				return
			}

			b.SendMessage(SendMessageParams{
				Message: "A shifty looking merchant hands " + username + " a glittering potion. Without hesitation, they sink the whole drink. " +
					username + " " + outcome + " " + stat.LongName,
			})

			userStats, err := b.statsService.GetUserStats(ctx, userID)
			if err != nil {
				b.logger.Error("failed to get stats", "err", err, "user", username)
				return
			}
			b.SendMessage(SendMessageParams{Message: stats.FormatStats(username, userStats)})
		},

		"Tempt the Dice": func(ctx context.Context, b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			userID := event.UserID
			username := event.UserName

			b.SendMessage(SendMessageParams{
				Message: username + " has rolled with initiative.",
			})

			userStats, err := b.statsService.GetOrCreateStats(ctx, userID, username)
			if err != nil {
				b.logger.Error("failed to get stats", "err", err, "user", username)
				return
			}
			b.SendMessage(SendMessageParams{Message: stats.FormatStats(username, userStats)})
		},
	}

	for _, cfg := range seriesConfigs {
		redemptions[cfg.RedemptionTitle] = func(ctx context.Context, b *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			redeemBlindBox(ctx, b, event.UserID, event.UserName, cfg)
		}
	}

	return redemptions
}

// redeemBlindBox picks a random plushie, records it, and broadcasts the SSE event.
func redeemBlindBox(ctx context.Context, b *Bot, userID, username string, cfg blindbox.SeriesConfig) {
	plushie, err := blindbox.PickPlushie(cfg.Plushies)
	if err != nil {
		b.logger.Error("failed to pick plushie", "err", err, "series", cfg.Series)
		b.SendMessage(
			SendMessageParams{
				Message: fmt.Sprintf("@%s sorry, the redemption failed. Please ping @modservo.", username),
			},
		)
		return
	}

	result, err := b.blindboxService.Redeem(ctx, userID, username, cfg.Series, plushie.Key)
	if err != nil {
		b.logger.Error("failed to redeem blind box", "err", err, "user", username)
		b.SendMessage(
			SendMessageParams{
				Message: fmt.Sprintf("@%s sorry, the redemption failed. Please ping @modservo.", username),
			},
		)
		return
	}

	b.broadcast(server.OverlayEvent{
		Type: server.EventTypeBlindBoxRedemption,
		Data: blindbox.BlindBoxRedemptionData{
			Username:   result.Username,
			Plushie:    plushie,
			IsNew:      result.IsNew,
			Collection: result.Collection,
			Config:     cfg,
		},
	})
	b.logger.Info(
		"blind box redeemed",
		"user",
		username,
		"series",
		cfg.Series,
		"plushie",
		plushie.Key,
		"is_new",
		result.IsNew,
	)
}

package charsibot

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/twitch/blindbox"
	"github.com/lukeramljak/charsibot/twitch/server"
	"github.com/lukeramljak/charsibot/twitch/stats"
)

const (
	statsCommandMinParts    = 5
	statsSubCommandAdd      = "add"
	statsSubCommandSet      = "set"
	statsSubCommandRm       = "rm"
	blindBoxCommandMinParts = 2
)

type Command struct {
	ModeratorOnly bool
	Execute       func(b *Bot, event twitch.EventChannelChatMessage)
}

// Commands returns the full map of chat commands keyed by trigger word.
//
//nolint:cyclop // Commands is a registry mapping every command.
func Commands(seriesConfigs []blindbox.SeriesConfig) map[string]Command {
	cmds := map[string]Command{
		"collections": {
			Execute: func(b *Bot, _ twitch.EventChannelChatMessage) {
				collections, err := b.blindboxService.GetCompletedCollections(b.ctx)
				if err != nil {
					slog.Error("failed to get completed collections", "err", err)
					return
				}

				b.SendMessage(SendMessageParams{
					Message: "The following chatters have completed the below blind box collections:",
				})
				for _, row := range collections {
					b.SendMessage(SendMessageParams{
						Message: fmt.Sprintf("%s: %s", row.SeriesName, row.Usernames),
					})
				}
			},
		},

		"explode": {
			ModeratorOnly: true,
			Execute: func(b *Bot, event twitch.EventChannelChatMessage) {
				mentionedUser, err := extractMentionedUserFromFragments(event.Message.Fragments)
				if err != nil {
					b.SendMessage(SendMessageParams{
						Message:              "Missing user mention",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if _, err = b.statsService.GetOrCreateStats(
					b.ctx,
					mentionedUser.UserID,
					mentionedUser.UserLogin,
				); err != nil {
					slog.Error("failed to ensure stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to update stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if err = b.statsService.ModifyStatValue(b.ctx, mentionedUser.UserID, "penis", -1003); err != nil {
					slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to update stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				userStats, err := b.statsService.GetUserStats(b.ctx, mentionedUser.UserID)
				if err != nil {
					slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to get updated stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}
				b.SendMessage(SendMessageParams{
					Message:              stats.FormatStats(mentionedUser.UserLogin, userStats),
					ReplyParentMessageID: event.MessageId,
				})
			},
		},

		"leaderboard": {
			Execute: func(b *Bot, _ twitch.EventChannelChatMessage) {
				rows, err := b.statsService.GetStatLeaderboard(b.ctx)
				if err != nil {
					slog.Error("failed to get leaderboard", "err", err)
					b.SendMessage(SendMessageParams{
						Message: "Failed to get leaderboard",
					})
					return
				}

				parts := make([]string, len(rows))
				for i, r := range rows {
					parts[i] = fmt.Sprintf("%s %s (%d)", r.Emoji, r.Username, r.Value)
				}
				b.SendMessage(SendMessageParams{
					Message: strings.Join(parts, " | "),
				})
			},
		},

		"stats": {
			Execute: func(b *Bot, event twitch.EventChannelChatMessage) {
				parts := strings.Fields(event.Message.Text)

				if len(parts) == 1 || !IsModerator(event) {
					userStats, err := b.statsService.GetOrCreateStats(b.ctx, event.ChatterUserId, event.ChatterUserName)
					if err != nil {
						slog.Error("failed to get stats", "err", err, "user", event.ChatterUserName)
						return
					}
					b.SendMessage(SendMessageParams{
						Message:              stats.FormatStats(event.ChatterUserName, userStats),
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if len(parts) < statsCommandMinParts {
					b.SendMessage(SendMessageParams{
						Message:              "Usage: !stats <add|set|rm> <@user> <stat> <amount>",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				validSubcommands := []string{statsSubCommandAdd, statsSubCommandSet, statsSubCommandRm}

				subcommand := strings.ToLower(parts[1])
				if !slices.Contains(validSubcommands, subcommand) {
					b.SendMessage(SendMessageParams{
						Message:              "Usage: !stats <add|set|rm> <@user> <stat> <amount>",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				mentionedUser, err := extractMentionedUserFromFragments(event.Message.Fragments)
				if err != nil {
					b.SendMessage(SendMessageParams{
						Message:              "Missing user mention",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				statColumn := parts[3]
				amount, err := strconv.ParseInt(parts[4], 10, 64)
				if err != nil {
					b.SendMessage(SendMessageParams{
						Message:              "Invalid amount",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if subcommand == statsSubCommandRm {
					amount = -amount
				}

				if _, err = b.statsService.GetOrCreateStats(
					b.ctx,
					mentionedUser.UserID,
					mentionedUser.UserLogin,
				); err != nil {
					slog.Error("failed to ensure stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to update stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if subcommand == statsSubCommandSet {
					if err = b.statsService.SetStatValue(b.ctx, mentionedUser.UserID, statColumn, amount); err != nil {
						slog.Error("failed to set stat", "err", err, "user", mentionedUser.UserLogin)
						b.SendMessage(SendMessageParams{
							Message:              "Failed to update stats",
							ReplyParentMessageID: event.MessageId,
						})
						return
					}
				} else {
					if err = b.statsService.ModifyStatValue(
						b.ctx,
						mentionedUser.UserID,
						statColumn,
						amount,
					); err != nil {
						slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
						b.SendMessage(SendMessageParams{
							Message:              "Failed to update stats",
							ReplyParentMessageID: event.MessageId,
						})
						return
					}
				}

				userStats, err := b.statsService.GetUserStats(b.ctx, mentionedUser.UserID)
				if err != nil {
					slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to get stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}
				b.SendMessage(SendMessageParams{Message: stats.FormatStats(mentionedUser.UserLogin, userStats)})
			},
		},
	}

	for _, cfg := range seriesConfigs {
		cmds[cfg.Series] = Command{
			Execute: func(b *Bot, event twitch.EventChannelChatMessage) {
				parts := strings.Fields(event.Message.Text)
				subcommand := ""
				if len(parts) >= blindBoxCommandMinParts {
					subcommand = strings.ToLower(parts[1])
				}

				switch subcommand {
				case "redeem":
					if !IsModerator(event) {
						b.SendMessage(SendMessageParams{
							Message:              "You must be a moderator to use this subcommand",
							ReplyParentMessageID: event.MessageId,
						})
						return
					}
					redeemBlindBox(b, event.ChatterUserId, event.ChatterUserName, cfg)
				case "reset":
					if !IsModerator(event) {
						b.SendMessage(SendMessageParams{
							Message:              "You must be a moderator to use this subcommand",
							ReplyParentMessageID: event.MessageId,
						})
						return
					}
					if err := b.blindboxService.ResetCollection(b.ctx, event.ChatterUserId, cfg.Series); err != nil {
						slog.Error("failed to reset collection", "err", err, "user", event.ChatterUserName)
						return
					}
					slog.Info("collection reset", "user", event.ChatterUserName, "series", cfg.Series)
				default:
					userID := event.ChatterUserId
					username := event.ChatterUserName

					slots, err := b.blindboxService.GetCollection(b.ctx, userID, cfg.Series)
					if err != nil {
						slog.Error("failed to get collection", "err", err, "user", username)
						b.SendMessage(SendMessageParams{
							Message: fmt.Sprintf("Failed to get %s's collection", username),
						})
						return
					}

					b.broadcast(server.OverlayEvent{
						Type: server.EventTypeCollectionDisplay,
						Data: server.CollectionDisplayData{
							UserID:         userID,
							Username:       username,
							Series:         cfg.Series,
							Collection:     slots,
							CollectionSize: len(slots),
						},
					})

					slog.Info("displaying collection", "user", username, "series", cfg.Series, "size", len(slots))
				}
			},
		}
	}

	return cmds
}

// redeemBlindBox picks a random plushie, records it, and broadcasts the SSE event.
func redeemBlindBox(b *Bot, userID, username string, cfg blindbox.SeriesConfig) {
	key := blindbox.PickPlushie(cfg.Plushies)

	result, err := b.blindboxService.Redeem(b.ctx, userID, username, cfg.Series, key)
	if err != nil {
		slog.Error("failed to redeem blind box", "err", err, "user", username)
		return
	}

	b.broadcast(server.OverlayEvent{
		Type: server.EventTypeBlindBoxRedemption,
		Data: server.BlindBoxRedemptionData{
			UserID:         result.UserID,
			Username:       result.Username,
			Series:         result.Series,
			SeriesName:     cfg.RedemptionTitle,
			Plushie:        result.Plushie,
			IsNew:          result.IsNew,
			Collection:     result.Collection,
			CollectionSize: len(result.Collection),
		},
	})

	slog.Info("blind box redeemed",
		"user", username,
		"series", cfg.Series,
		"plushie", key,
		"isNew", result.IsNew,
	)
}

func extractMentionedUserFromFragments(
	fragments []twitch.ChatMessageFragment,
) (*twitch.ChatMessageFragmentMention, error) {
	for _, fragment := range fragments {
		if fragment.Type == "mention" && fragment.Mention != nil {
			return fragment.Mention, nil
		}
	}
	return nil, errors.New("no user mention found")
}

// IsModerator reports whether the event sender has broadcaster, moderator, or
// lead_moderator badge.
func IsModerator(event twitch.EventChannelChatMessage) bool {
	for _, badge := range event.Badges {
		switch badge.SetId {
		case "broadcaster", "moderator", "lead_moderator":
			return true
		}
	}
	return false
}

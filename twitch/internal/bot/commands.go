package bot

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/twitch/internal/store"
)

const (
	statsCommandMinParts    = 5
	blindBoxCommandMinParts = 2
)

type Command struct {
	ModeratorOnly bool
	Execute       func(b *Bot, event twitch.EventChannelChatMessage)
}

// Commands returns the full map of chat commands keyed by trigger word.
//
//nolint:cyclop // Commands is a registry mapping every command.
func Commands(seriesConfigs []SeriesConfig) map[string]Command {
	cmds := map[string]Command{
		"collections": {
			Execute: func(b *Bot, _ twitch.EventChannelChatMessage) {
				collections, err := b.store.GetCompletedCollections(b.ctx)
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

				if _, err = GetOrCreateStats(
					b.ctx,
					b.store,
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

				if err = b.store.ModifyStatValue(b.ctx, store.ModifyStatValueParams{
					Value:    -1003,
					UserID:   mentionedUser.UserID,
					StatName: "penis",
				}); err != nil {
					slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to update stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				stats, err := b.store.GetUserStats(b.ctx, mentionedUser.UserID)
				if err != nil {
					slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to get updated stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}
				b.SendMessage(SendMessageParams{
					Message:              FormatStats(mentionedUser.UserLogin, stats),
					ReplyParentMessageID: event.MessageId,
				})
			},
		},

		"leaderboard": {
			Execute: func(b *Bot, _ twitch.EventChannelChatMessage) {
				rows, err := b.store.GetStatLeaderboard(b.ctx)
				if err != nil {
					slog.Error("failed to get leaderboard", "err", err)
					return
				}

				type leaderboardEntry struct {
					DisplayName string `json:"displayName"`
					Username    string `json:"username"`
					Value       int64  `json:"value"`
				}
				entries := make([]leaderboardEntry, len(rows))
				for i, r := range rows {
					entries[i] = leaderboardEntry{
						DisplayName: r.ShortName,
						Username:    r.Username,
						Value:       r.Value,
					}
				}

				b.BroadcastOverlayEvent(OverlayEvent{
					Type: EventTypeLeaderboard,
					Data: entries,
				})
			},
		},

		"stats": {
			Execute: func(b *Bot, event twitch.EventChannelChatMessage) {
				parts := strings.Fields(event.Message.Text)

				if len(parts) == 1 || !IsModerator(event) {
					stats, err := GetOrCreateStats(b.ctx, b.store, event.ChatterUserId, event.ChatterUserName)
					if err != nil {
						slog.Error("failed to get stats", "err", err, "user", event.ChatterUserName)
						return
					}
					b.SendMessage(SendMessageParams{
						Message:              FormatStats(event.ChatterUserName, stats),
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				if len(parts) < statsCommandMinParts {
					b.SendMessage(SendMessageParams{
						Message:              "Usage: !stats <add|rm> <@user> <stat> <amount>",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				subcommand := strings.ToLower(parts[1])
				if subcommand != "add" && subcommand != "rm" {
					b.SendMessage(SendMessageParams{
						Message:              "Usage: !stats <add|rm> <@user> <stat> <amount>",
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

				if subcommand == "rm" {
					amount = -amount
				}

				if _, err = GetOrCreateStats(
					b.ctx,
					b.store,
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

				if err = b.store.ModifyStatValue(b.ctx, store.ModifyStatValueParams{
					Value:    amount,
					UserID:   mentionedUser.UserID,
					StatName: statColumn,
				}); err != nil {
					slog.Error("failed to modify stat", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to update stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}

				stats, err := b.store.GetUserStats(b.ctx, mentionedUser.UserID)
				if err != nil {
					slog.Error("failed to get stats", "err", err, "user", mentionedUser.UserLogin)
					b.SendMessage(SendMessageParams{
						Message:              "Failed to get stats",
						ReplyParentMessageID: event.MessageId,
					})
					return
				}
				b.SendMessage(SendMessageParams{Message: FormatStats(mentionedUser.UserLogin, stats)})
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
					RedeemBlindBox(b, event.ChatterUserId, event.ChatterUserName, cfg)
				case "reset":
					if !IsModerator(event) {
						b.SendMessage(SendMessageParams{
							Message:              "You must be a moderator to use this subcommand",
							ReplyParentMessageID: event.MessageId,
						})
						return
					}
					if err := b.store.ResetUserPlushies(b.ctx, store.ResetUserPlushiesParams{
						UserID: event.ChatterUserId,
						Series: cfg.Series,
					}); err != nil {
						slog.Error("failed to reset collection", "err", err, "user", event.ChatterUserName)
						return
					}
					slog.Info("collection reset", "user", event.ChatterUserName, "series", cfg.Series)
				default:
					userID := event.ChatterUserId
					username := event.ChatterUserName

					slots, err := b.store.GetCollectedPlushies(b.ctx, store.GetCollectedPlushiesParams{
						UserID: userID,
						Series: cfg.Series,
					})
					if err != nil {
						slog.Error("failed to get collection", "err", err, "user", username)
						b.SendMessage(SendMessageParams{
							Message: fmt.Sprintf("Failed to get %s's collection", username),
						})
						return
					}

					b.BroadcastOverlayEvent(OverlayEvent{
						Type: EventTypeCollectionDisplay,
						Data: map[string]any{
							"userId":         userID,
							"username":       username,
							"series":         cfg.Series,
							"collection":     slots,
							"collectionSize": len(slots),
						},
					})

					slog.Info("displaying collection", "user", username, "series", cfg.Series, "size", len(slots))
				}
			},
		}
	}

	return cmds
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

func IsModerator(event twitch.EventChannelChatMessage) bool {
	for _, badge := range event.Badges {
		switch badge.SetId {
		case "broadcaster", "moderator", "lead_moderator":
			return true
		}
	}
	return false
}

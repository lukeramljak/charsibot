package charsibot

import (
	"context"
	"fmt"
	"strings"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/server"
	"github.com/lukeramljak/charsibot/stats"
)

type Command struct {
	Execute func(ctx context.Context, b *Bot, event twitch.EventChannelChatMessage)
}

// Commands returns the full map of chat commands keyed by trigger word.
func Commands(seriesConfigs []blindbox.SeriesConfig) map[string]Command {
	cmds := map[string]Command{
		"collections": {
			Execute: func(ctx context.Context, b *Bot, _ twitch.EventChannelChatMessage) {
				collections, err := b.blindboxService.GetCompletedCollections(ctx)
				if err != nil {
					b.logger.Error("failed to get completed collections", "err", err)
					return
				}
				b.SendMessage(SendMessageParams{Message: "The following chatters have completed the below blind box collections:"})
				for _, row := range collections {
					b.SendMessage(SendMessageParams{Message: fmt.Sprintf("%s: %s", row.SeriesName, row.Usernames)})
				}
			},
		},
		"leaderboard": {
			Execute: func(ctx context.Context, b *Bot, _ twitch.EventChannelChatMessage) {
				rows, err := b.statsService.GetStatLeaderboard(ctx)
				if err != nil {
					b.logger.Error("failed to get leaderboard", "err", err)
					b.SendMessage(SendMessageParams{Message: "Failed to get leaderboard"})
					return
				}
				parts := make([]string, len(rows))
				for i, row := range rows {
					parts[i] = fmt.Sprintf("%s %s (%d)", row.Emoji, row.Username, row.Value)
				}
				b.SendMessage(SendMessageParams{Message: strings.Join(parts, " | ")})
			},
		},
		"stats": {
			Execute: func(ctx context.Context, b *Bot, event twitch.EventChannelChatMessage) {
				if len(strings.Fields(event.Message.Text)) != 1 {
					return
				}
				userStats, err := b.statsService.GetOrCreateStats(ctx, event.ChatterUserId, event.ChatterUserName)
				if err != nil {
					b.logger.Error("failed to get stats", "err", err, "user", event.ChatterUserName)
					return
				}
				b.SendMessage(SendMessageParams{Message: stats.FormatStats(event.ChatterUserName, userStats), ReplyParentMessageID: event.MessageId})
			},
		},
	}

	for _, cfg := range seriesConfigs {
		cmds[cfg.Series] = Command{
			Execute: func(ctx context.Context, b *Bot, event twitch.EventChannelChatMessage) {
				if len(strings.Fields(event.Message.Text)) != 1 {
					return
				}
				slots, err := b.blindboxService.GetCollection(ctx, event.ChatterUserId, cfg.Series)
				if err != nil {
					b.logger.Error("failed to get collection", "err", err, "user", event.ChatterUserName)
					b.SendMessage(SendMessageParams{Message: fmt.Sprintf("Failed to get %s's collection", event.ChatterUserName)})
					return
				}
				b.broadcast(server.OverlayEvent{Type: server.EventTypeCollectionDisplay, Data: blindbox.BlindBoxDisplayData{Username: event.ChatterUserName, Collection: slots, Config: cfg}})
				b.logger.Info("displaying collection", "user", event.ChatterUserName, "series", cfg.Series, "size", len(slots))
			},
		}
	}

	return cmds
}

package twitchapp

import (
	"context"

	"github.com/lukeramljak/charsibot/twitch/internal/constants"
	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

func (c *Client) registerBuiltInCommands() {
	c.RegisterCommand(constants.CmdStats, c.handleStatsCommand)
}

func (c *Client) handleStatsCommand(ctx context.Context, msg *eventsub.ChannelChatMessage) (string, error) {
	return c.statsStore.GetMessage(ctx, msg.ChatterUserID, msg.ChatterUserLogin)
}

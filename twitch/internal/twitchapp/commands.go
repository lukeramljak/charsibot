package twitchapp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/lukeramljak/charsibot/twitch/internal/constants"
	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

func (c *Client) registerBuiltInCommands() {
	c.RegisterCommand(constants.CmdGetStats, c.handleGetStatsCommand)
	c.RegisterCommand(constants.CmdAddStat, c.handleModifyStatCommand)
	c.RegisterCommand(constants.CmdRmStat, c.handleModifyStatCommand)
}

func (c *Client) handleGetStatsCommand(ctx context.Context, msg *eventsub.ChannelChatMessage) (string, error) {
	return c.statsStore.GetMessage(ctx, msg.ChatterUserID, msg.ChatterUserLogin)
}

func (c *Client) handleModifyStatCommand(ctx context.Context, msg *eventsub.ChannelChatMessage) (string, error) {
	if !isModerator(msg) {
		return "You must be a moderator to use this command", nil
	}

	stats, err := parseStatCommand(msg)
	if err != nil {
		return "", err
	}

	if err := c.statsStore.ModifyStat(ctx, stats.userID, stats.username, stats.stat, stats.amount); err != nil {
		return "", err
	}

	return c.statsStore.GetMessage(ctx, stats.userID, stats.username)
}

func isModerator(msg *eventsub.ChannelChatMessage) bool {
	for _, badge := range msg.Badges {
		if badge.SetID == "moderator" || badge.SetID == "broadcaster" {
			return true
		}
	}
	return false
}

type statCommand struct {
	userID   string
	username string
	stat     string
	amount   int
}

func parseStatCommand(msg *eventsub.ChannelChatMessage) (*statCommand, error) {
	var stats statCommand

	foundMention := false
	remainder := ""

	fields := strings.Fields(msg.Message.Text)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty message")
	}

	cmd := strings.TrimSpace(fields[0])

	for _, frag := range msg.Message.Fragments {
		if frag.Type == "mention" {
			stats.userID = frag.Mention.UserID
			stats.username = frag.Mention.UserLogin
			foundMention = true
			continue
		}
		if foundMention && frag.Type == "text" {
			remainder += frag.Text
		}
	}

	if !foundMention {
		return nil, fmt.Errorf("no user mention found")
	}

	remainder = strings.TrimSpace(remainder)
	if remainder == "" {
		return nil, fmt.Errorf("no stat or amount provided")
	}

	parts := strings.Fields(remainder)
	if len(parts) < 2 {
		return nil, fmt.Errorf("expected 'stat amount'")
	}

	stats.stat = parts[0]

	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid number: %v", err)
	}

	if cmd == constants.CmdRmStat {
		stats.amount = -n
	} else {
		stats.amount = n
	}

	return &stats, nil
}

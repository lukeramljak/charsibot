package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

func (b *Bot) handleStatsCommand(ctx context.Context, msg *eventsub.ChannelChatMessage) (string, error) {
	return b.getStatsMessage(ctx, msg.ChatterUserID, msg.ChatterUserLogin)
}

func (b *Bot) handleModifyStatCommand(ctx context.Context, msg *eventsub.ChannelChatMessage) (string, error) {
	isMod := false
	for _, badge := range msg.Badges {
		if badge.SetID == "moderator" || badge.SetID == "broadcaster" {
			isMod = true
			break
		}
	}
	if !isMod {
		return "You must be a moderator to use this command", nil
	}

	var userID, username, statColumn string
	var amount int

	fields := strings.Fields(msg.Message.Text)
	if len(fields) == 0 {
		return "", fmt.Errorf("empty message")
	}
	cmd := fields[0]

	foundMention := false
	remainder := ""
	for _, frag := range msg.Message.Fragments {
		if frag.Type == "mention" {
			userID = frag.Mention.UserID
			username = frag.Mention.UserLogin
			foundMention = true
			continue
		}
		if foundMention && frag.Type == "text" {
			remainder += frag.Text
		}
	}

	if !foundMention {
		return "", fmt.Errorf("no user mention found")
	}

	remainder = strings.TrimSpace(remainder)
	parts := strings.Fields(remainder)
	if len(parts) < 2 {
		return "", fmt.Errorf("expected 'stat amount'")
	}

	statColumn = parts[0]
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid number: %v", err)
	}

	if cmd == "!rmstat" {
		amount = -n
	} else {
		amount = n
	}

	if err := b.store.ModifyStat(ctx, userID, username, statColumn, amount); err != nil {
		return "", err
	}

	return b.getStatsMessage(ctx, userID, username)
}

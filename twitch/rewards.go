package main

import (
	"context"
	"fmt"

	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

func (b *Bot) onDrinkPotion(ctx context.Context, event *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error {
	stat := b.randomStat()
	delta := b.randomDelta()

	if err := b.store.ModifyStat(ctx, event.UserID, event.UserLogin, stat.Column, delta); err != nil {
		return err
	}

	outcome := "gained"
	if delta < 0 {
		outcome = "lost"
	}

	msg := fmt.Sprintf("A shifty looking merchant hands %s a glittering potion. Without hesitation, they sink the whole drink. %s %s %s",
		event.UserLogin, event.UserLogin, outcome, stat.Display)

	if err := b.send(msg, ""); err != nil {
		return err
	}

	statsMsg, err := b.getStatsMessage(ctx, event.UserID, event.UserLogin)
	if err != nil {
		return err
	}
	return b.send(statsMsg, "")
}

func (b *Bot) onTemptDice(ctx context.Context, event *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error {
	if err := b.send(fmt.Sprintf("%s has rolled with initiative.", event.UserLogin), ""); err != nil {
		return err
	}

	statsMsg, err := b.getStatsMessage(ctx, event.UserID, event.UserLogin)
	if err != nil {
		return err
	}
	return b.send(statsMsg, "")
}

package twitchapp

import (
	"context"
	"fmt"

	"github.com/lukeramljak/charsibot/twitch/internal/constants"
	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

func (c *Client) registerBuiltInRewards() {
	c.RegisterReward(constants.RewardDrinkPotion, c.rewardDrinkPotion)
	c.RegisterReward(constants.RewardTemptDice, c.rewardRollDice)
}

func (c *Client) rewardDrinkPotion(ctx context.Context, event *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error {
	stat := c.statsStore.RandomStat(c.rng)
	delta := c.statsStore.RandomDelta(c.rng)

	if err := c.statsStore.IncrementStat(ctx, event.UserID, event.UserLogin, stat.Column, delta); err != nil {
		return err
	}

	outcome := "gained"
	if delta < 0 {
		outcome = "lost"
	}

	msg := fmt.Sprintf("A shifty looking merchant hands %s a glittering potion. Without hesitation, they sink the whole drink. %s %s %s", event.UserLogin, event.UserLogin, outcome, stat.Display)
	if err := c.SendChatMessage(ctx, msg); err != nil {
		return err
	}

	return c.sendUserStats(ctx, event.UserID, event.UserLogin)
}

func (c *Client) rewardRollDice(ctx context.Context, event *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error {
	if err := c.SendChatMessage(ctx, fmt.Sprintf("%s has rolled with initiative.", event.UserLogin)); err != nil {
		return err
	}
	return c.sendUserStats(ctx, event.UserID, event.UserLogin)
}

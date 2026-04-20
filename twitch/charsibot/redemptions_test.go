package charsibot

import (
	"context"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"
)

func TestOnChannelPointRedemption(t *testing.T) {
	tests := []struct {
		name           string
		rewardTitle    string
		redemptions    map[string]RedemptionFunc
		expectExecuted string // reward title key expected to fire, empty if none
	}{
		{
			name:           "matching redemption executes",
			rewardTitle:    "Special Reward",
			expectExecuted: "Special Reward",
			redemptions: map[string]RedemptionFunc{
				"Special Reward": nil, // replaced per-test below
			},
		},
		{
			name:        "non-matching reward does nothing",
			rewardTitle: "Other Reward",
			redemptions: map[string]RedemptionFunc{
				"Special Reward": nil,
			},
		},
		{
			name:        "empty redemptions map",
			rewardTitle: "Any Reward",
			redemptions: map[string]RedemptionFunc{},
		},
		{
			name:        "nil redemptions map",
			rewardTitle: "Any Reward",
			redemptions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed := map[string]bool{}

			redemptions := make(map[string]RedemptionFunc, len(tt.redemptions))
			for key := range tt.redemptions {
				k := key
				redemptions[k] = func(_ *Bot, _ twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
					executed[k] = true
				}
			}

			b := createTestBotForRedemption(t)
			b.redemptions = redemptions

			event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
				User: twitch.User{UserName: "testuser"},
				Reward: twitch.CustomChannelPointReward{
					Title: tt.rewardTitle,
				},
			}

			b.onChannelPointRedemption(event)

			if tt.expectExecuted != "" {
				if !executed[tt.expectExecuted] {
					t.Errorf("expected redemption %q to execute, but it didn't", tt.expectExecuted)
				}
			}

			for key := range executed {
				if key != tt.expectExecuted {
					t.Errorf("redemption %q should not have executed", key)
				}
			}
		})
	}
}

func TestRedemptionTitleMatching(t *testing.T) {
	b := createTestBotForRedemption(t)

	executed := false
	b.redemptions = map[string]RedemptionFunc{
		"Special Reward": func(_ *Bot, _ twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			executed = true
		},
	}

	// Reward that should trigger
	b.onChannelPointRedemption(twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User:   twitch.User{UserName: "testuser"},
		Reward: twitch.CustomChannelPointReward{Title: "Special Reward"},
	})

	if !executed {
		t.Error("redemption should have executed for matching reward")
	}

	executed = false

	// Reward that should not trigger
	b.onChannelPointRedemption(twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User:   twitch.User{UserName: "testuser"},
		Reward: twitch.CustomChannelPointReward{Title: "Different Reward"},
	})

	if executed {
		t.Error("redemption should not have executed for non-matching reward")
	}
}

func createTestBotForRedemption(t *testing.T) *Bot {
	t.Helper()

	cfg := Config{
		BotUserID:     "bot123",
		ChannelUserID: "channel456",
	}

	return &Bot{
		config: cfg,
		ctx:    context.Background(),
	}
}

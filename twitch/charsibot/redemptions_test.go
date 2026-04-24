package charsibot

import (
	"context"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/twitch/db"
)

func TestDrinkAPotionCreatesStatsForNewUser(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	b := &Bot{
		config: Config{BotUserID: "bot1", ChannelUserID: "ch1"},
		ctx:    ctx,
		store:  queries,
	}
	b.redemptions = Redemptions(nil)

	event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User:   twitch.User{UserID: "newuser1", UserName: "newuser"},
		Reward: twitch.CustomChannelPointReward{Title: "Drink a Potion"},
	}

	statsBefore, _ := queries.GetUserStats(ctx, "newuser1")
	if len(statsBefore) != 0 {
		t.Fatalf("expected no stats before redemption, got %d", len(statsBefore))
	}

	b.onChannelPointRedemption(event)

	statsAfter, err := queries.GetUserStats(ctx, "newuser1")
	if err != nil {
		t.Fatalf("GetUserStats failed: %v", err)
	}
	if len(statsAfter) == 0 {
		t.Fatal("expected stats to be created for new user after Drink a Potion, got none")
	}

	changed := 0
	for _, s := range statsAfter {
		if s.Value != 3 {
			changed++
		}
	}
	if changed == 0 {
		t.Error("expected at least one stat to be modified by the potion, but all remain at default")
	}
}

func TestDrinkAPotionModifiesStatForExistingUser(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	if _, err := GetOrCreateStats(ctx, queries, "existinguser1", "existing"); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	b := &Bot{
		config: Config{BotUserID: "bot1", ChannelUserID: "ch1"},
		ctx:    ctx,
		store:  queries,
	}
	b.redemptions = Redemptions(nil)

	event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User:   twitch.User{UserID: "existinguser1", UserName: "existing"},
		Reward: twitch.CustomChannelPointReward{Title: "Drink a Potion"},
	}

	b.onChannelPointRedemption(event)

	stats, err := queries.GetUserStats(ctx, "existinguser1")
	if err != nil {
		t.Fatalf("GetUserStats failed: %v", err)
	}
	if len(stats) == 0 {
		t.Fatal("expected stats to exist after redemption")
	}

	changed := 0
	for _, s := range stats {
		if s.Value != 3 {
			changed++
		}
	}
	if changed == 0 {
		t.Error("expected at least one stat to be modified by the potion, but all remain at default")
	}
}

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

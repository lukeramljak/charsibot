package bot

import (
	"context"
	"log/slog"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/config"
)

// MockRedemption is a test redemption implementation
type MockRedemption struct {
	shouldTrigger   bool
	executeCalled   bool
	executeEvent    *twitch.EventChannelChannelPointsCustomRewardRedemptionAdd
	shouldTriggerFn func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool
}

func (m *MockRedemption) ShouldTrigger(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool {
	if m.shouldTriggerFn != nil {
		return m.shouldTriggerFn(event)
	}
	return m.shouldTrigger
}

func (m *MockRedemption) Execute(bot *Bot, event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	m.executeCalled = true
	m.executeEvent = &event
}

func TestRedemptionHandler_Process(t *testing.T) {
	tests := []struct {
		name           string
		rewardTitle    string
		redemptions    []Redemption
		expectExecuted []bool // which redemptions should execute
	}{
		{
			name:        "redemption matches and executes",
			rewardTitle: "Test Reward",
			redemptions: []Redemption{
				&MockRedemption{shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:        "redemption doesn't match",
			rewardTitle: "Test Reward",
			redemptions: []Redemption{
				&MockRedemption{shouldTrigger: false},
			},
			expectExecuted: []bool{false},
		},
		{
			name:        "multiple redemptions all match",
			rewardTitle: "Test Reward",
			redemptions: []Redemption{
				&MockRedemption{shouldTrigger: true},
				&MockRedemption{shouldTrigger: true},
			},
			expectExecuted: []bool{true, true},
		},
		{
			name:        "some redemptions match",
			rewardTitle: "Test Reward",
			redemptions: []Redemption{
				&MockRedemption{shouldTrigger: true},
				&MockRedemption{shouldTrigger: false},
				&MockRedemption{shouldTrigger: true},
			},
			expectExecuted: []bool{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := createTestBotForRedemption(t)

			// Reset mock redemptions
			for _, redem := range tt.redemptions {
				if mock, ok := redem.(*MockRedemption); ok {
					mock.executeCalled = false
					mock.executeEvent = nil
				}
			}

			handler := NewRedemptionHandler(tt.redemptions)

			event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
				User: twitch.User{
					UserName: "testuser",
				},
				Reward: twitch.CustomChannelPointReward{
					Title: tt.rewardTitle,
				},
			}

			handler.Process(bot, event)

			// Check if expected redemptions were executed
			for i, expected := range tt.expectExecuted {
				mock := tt.redemptions[i].(*MockRedemption)
				if mock.executeCalled != expected {
					t.Errorf("redemption %d: executeCalled = %v, want %v", i, mock.executeCalled, expected)
				}
				if expected && mock.executeEvent == nil {
					t.Errorf("redemption %d: execute was called but event was nil", i)
				}
				if expected && mock.executeEvent != nil && mock.executeEvent.Reward.Title != tt.rewardTitle {
					t.Errorf("redemption %d: event reward title = %q, want %q", i, mock.executeEvent.Reward.Title, tt.rewardTitle)
				}
			}
		})
	}
}

func TestRedemptionHandler_ProcessEmptyHandler(t *testing.T) {
	bot := createTestBotForRedemption(t)
	handler := NewRedemptionHandler([]Redemption{})

	event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User: twitch.User{
			UserName: "testuser",
		},
		Reward: twitch.CustomChannelPointReward{
			Title: "Test Reward",
		},
	}

	// Should not panic with empty redemption list
	handler.Process(bot, event)
}

func TestRedemptionHandler_ProcessNilRedemptions(t *testing.T) {
	bot := createTestBotForRedemption(t)
	handler := NewRedemptionHandler(nil)

	event := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User: twitch.User{
			UserName: "testuser",
		},
		Reward: twitch.CustomChannelPointReward{
			Title: "Test Reward",
		},
	}

	// Should not panic with nil redemption list
	handler.Process(bot, event)
}

func TestRedemptionHandler_RewardTitleMatching(t *testing.T) {
	bot := createTestBotForRedemption(t)

	mock := &MockRedemption{
		shouldTriggerFn: func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) bool {
			return event.Reward.Title == "Special Reward"
		},
	}

	handler := NewRedemptionHandler([]Redemption{mock})

	// Reward that should trigger
	event1 := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User: twitch.User{
			UserName: "testuser",
		},
		Reward: twitch.CustomChannelPointReward{
			Title: "Special Reward",
		},
	}

	handler.Process(bot, event1)

	if !mock.executeCalled {
		t.Error("redemption should have executed for matching reward")
	}

	// Reset
	mock.executeCalled = false

	// Reward that should not trigger
	event2 := twitch.EventChannelChannelPointsCustomRewardRedemptionAdd{
		User: twitch.User{
			UserName: "testuser",
		},
		Reward: twitch.CustomChannelPointReward{
			Title: "Different Reward",
		},
	}

	handler.Process(bot, event2)

	if mock.executeCalled {
		t.Error("redemption should not have executed for non-matching reward")
	}
}

// Helper function to create a test bot
func createTestBotForRedemption(t *testing.T) *Bot {
	t.Helper()

	cfg := config.Config{
		BotUserID:     "bot123",
		ChannelUserID: "channel456",
	}

	logger := slog.Default()

	bot := &Bot{
		config: cfg,
		logger: logger,
		ctx:    context.Background(),
	}

	return bot
}

package bot

import (
	"context"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/internal/config"
)

func TestProcessTriggers(t *testing.T) {
	type triggerSpec struct {
		chance        int
		shouldTrigger bool
	}

	tests := []struct {
		name           string
		message        string
		specs          []triggerSpec
		expectExecuted []bool
	}{
		{
			name:    "trigger matches and executes",
			message: "hello",
			specs: []triggerSpec{
				{chance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "trigger doesn't match",
			message: "hello",
			specs: []triggerSpec{
				{chance: 100, shouldTrigger: false},
			},
			expectExecuted: []bool{false},
		},
		{
			name:    "multiple triggers all match",
			message: "hello",
			specs: []triggerSpec{
				{chance: 100, shouldTrigger: true},
				{chance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true, true},
		},
		{
			name:    "some triggers match",
			message: "hello",
			specs: []triggerSpec{
				{chance: 100, shouldTrigger: true},
				{chance: 100, shouldTrigger: false},
				{chance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true, false, true},
		},
		{
			name:    "0% chance always executes if matches",
			message: "hello",
			specs: []triggerSpec{
				{chance: 0, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "100% chance always executes if matches",
			message: "hello",
			specs: []triggerSpec{
				{chance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "negative chance always executes if matches",
			message: "hello",
			specs: []triggerSpec{
				{chance: -10, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "greater than 100 chance always executes if matches",
			message: "hello",
			specs: []triggerSpec{
				{chance: 150, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed := make([]bool, len(tt.specs))
			triggers := make([]Trigger, len(tt.specs))
			for i, spec := range tt.specs {
				idx := i
				trigger := spec.shouldTrigger
				triggers[i] = Trigger{
					Chance: spec.chance,
					ShouldTrigger: func(_ twitch.EventChannelChatMessage) bool {
						return trigger
					},
					Execute: func(_ *Bot, _ twitch.EventChannelChatMessage) {
						executed[idx] = true
					},
				}
			}

			b := createTestBotForTrigger(t)
			b.triggers = triggers

			event := twitch.EventChannelChatMessage{
				Chatter: twitch.Chatter{
					ChatterUserId:   "user123",
					ChatterUserName: "testuser",
				},
				Message: twitch.ChatMessage{Text: tt.message},
			}

			b.processTriggers(event)

			for i, expected := range tt.expectExecuted {
				if executed[i] != expected {
					t.Errorf("trigger %d: executed = %v, want %v", i, executed[i], expected)
				}
			}
		})
	}
}

func TestProcessTriggers_ProbabilityDistribution(t *testing.T) {
	b := createTestBotForTrigger(t)

	iterations := 1000
	executionCounts := 0

	event := twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "test"},
	}

	for range iterations {
		count := 0
		b.triggers = []Trigger{{
			Chance:        50,
			ShouldTrigger: func(_ twitch.EventChannelChatMessage) bool { return true },
			Execute:       func(_ *Bot, _ twitch.EventChannelChatMessage) { count++ },
		}}
		b.processTriggers(event)
		executionCounts += count
	}

	minExpected := int(float64(iterations) * 0.40)
	maxExpected := int(float64(iterations) * 0.60)

	if executionCounts < minExpected || executionCounts > maxExpected {
		t.Errorf("with 50%% chance and %d iterations, got %d executions, want between %d and %d",
			iterations, executionCounts, minExpected, maxExpected)
	}

	t.Logf(
		"50%% chance: %d/%d executions (%.1f%%)",
		executionCounts,
		iterations,
		float64(executionCounts)/float64(iterations)*100,
	)
}

func TestProcessTriggers_ProbabilityLowChance(t *testing.T) {
	b := createTestBotForTrigger(t)

	iterations := 10000
	executionCounts := 0

	event := twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "test"},
	}

	for range iterations {
		count := 0
		b.triggers = []Trigger{{
			Chance:        1,
			ShouldTrigger: func(_ twitch.EventChannelChatMessage) bool { return true },
			Execute:       func(_ *Bot, _ twitch.EventChannelChatMessage) { count++ },
		}}
		b.processTriggers(event)
		executionCounts += count
	}

	if executionCounts < 50 || executionCounts > 200 {
		t.Errorf("with 1%% chance and %d iterations, got %d executions, want between 50 and 200",
			iterations, executionCounts)
	}

	t.Logf(
		"1%% chance: %d/%d executions (%.2f%%)",
		executionCounts,
		iterations,
		float64(executionCounts)/float64(iterations)*100,
	)
}

func TestProcessTriggers_ProbabilityHighChance(t *testing.T) {
	b := createTestBotForTrigger(t)

	iterations := 1000
	executionCounts := 0

	event := twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "test"},
	}

	for range iterations {
		count := 0
		b.triggers = []Trigger{{
			Chance:        99,
			ShouldTrigger: func(_ twitch.EventChannelChatMessage) bool { return true },
			Execute:       func(_ *Bot, _ twitch.EventChannelChatMessage) { count++ },
		}}
		b.processTriggers(event)
		executionCounts += count
	}

	if executionCounts < 970 {
		t.Errorf("with 99%% chance and %d iterations, got %d executions, want at least 970",
			iterations, executionCounts)
	}

	t.Logf(
		"99%% chance: %d/%d executions (%.1f%%)",
		executionCounts,
		iterations,
		float64(executionCounts)/float64(iterations)*100,
	)
}

func TestProcessTriggers_Empty(t *testing.T) {
	b := createTestBotForTrigger(t)
	b.triggers = []Trigger{}

	b.processTriggers(twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "test"},
	})
}

func TestProcessTriggers_Nil(t *testing.T) {
	b := createTestBotForTrigger(t)
	b.triggers = nil

	b.processTriggers(twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "test"},
	})
}

func TestProcessTriggers_MessageContent(t *testing.T) {
	b := createTestBotForTrigger(t)

	executed := false
	b.triggers = []Trigger{{
		Chance: 100,
		ShouldTrigger: func(event twitch.EventChannelChatMessage) bool {
			return event.Message.Text == "trigger me"
		},
		Execute: func(_ *Bot, _ twitch.EventChannelChatMessage) {
			executed = true
		},
	}}

	b.processTriggers(twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "trigger me"},
	})

	if !executed {
		t.Error("trigger should have executed for matching message")
	}

	executed = false

	b.processTriggers(twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "don't trigger"},
	})

	if executed {
		t.Error("trigger should not have executed for non-matching message")
	}
}

func createTestBotForTrigger(t *testing.T) *Bot {
	t.Helper()

	cfg := config.Config{
		BotUserID:     "bot123",
		ChannelUserID: "channel456",
	}

	return &Bot{
		config: cfg,
		ctx:    context.Background(),
	}
}

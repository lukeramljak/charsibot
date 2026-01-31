package bot

import (
	"context"
	"log/slog"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/config"
)

type MockTrigger struct {
	triggerChance   int
	shouldTrigger   bool
	executeCalled   bool
	executeEvent    *twitch.EventChannelChatMessage
	shouldTriggerFn func(event twitch.EventChannelChatMessage) bool
}

func (m *MockTrigger) TriggerChance() int {
	return m.triggerChance
}

func (m *MockTrigger) ShouldTrigger(event twitch.EventChannelChatMessage) bool {
	if m.shouldTriggerFn != nil {
		return m.shouldTriggerFn(event)
	}
	return m.shouldTrigger
}

func (m *MockTrigger) Execute(bot *Bot, event twitch.EventChannelChatMessage) {
	m.executeCalled = true
	m.executeEvent = &event
}

func TestTriggerHandler_Process(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		triggers       []Trigger
		expectExecuted []bool // which triggers should execute
	}{
		{
			name:    "trigger matches and executes",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "trigger doesn't match",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 100, shouldTrigger: false},
			},
			expectExecuted: []bool{false},
		},
		{
			name:    "multiple triggers all match",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true, true},
		},
		{
			name:    "some triggers match",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
				&MockTrigger{triggerChance: 100, shouldTrigger: false},
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true, false, true},
		},
		{
			name:    "0% chance always executes if matches",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 0, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "100% chance always executes if matches",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 100, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "negative chance always executes if matches",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: -10, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
		{
			name:    "greater than 100 chance always executes if matches",
			message: "hello",
			triggers: []Trigger{
				&MockTrigger{triggerChance: 150, shouldTrigger: true},
			},
			expectExecuted: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot := createTestBotForTrigger(t)

			// Reset mock triggers
			for _, trig := range tt.triggers {
				if mock, ok := trig.(*MockTrigger); ok {
					mock.executeCalled = false
					mock.executeEvent = nil
				}
			}

			handler := NewTriggerHandler(tt.triggers)

			event := twitch.EventChannelChatMessage{
				Chatter: twitch.Chatter{
					ChatterUserId:   "user123",
					ChatterUserName: "testuser",
				},
				Message: twitch.ChatMessage{
					Text: tt.message,
				},
			}

			handler.Process(bot, event)

			// Check if expected triggers were executed
			for i, expected := range tt.expectExecuted {
				mock := tt.triggers[i].(*MockTrigger)
				if mock.executeCalled != expected {
					t.Errorf("trigger %d: executeCalled = %v, want %v", i, mock.executeCalled, expected)
				}
				if expected && mock.executeEvent == nil {
					t.Errorf("trigger %d: execute was called but event was nil", i)
				}
			}
		})
	}
}

func TestTriggerHandler_ProbabilityDistribution(t *testing.T) {
	// Test that probability works statistically
	// With 50% chance, over 1000 runs we should see roughly 400-600 executions
	bot := createTestBotForTrigger(t)

	iterations := 1000
	executionCounts := 0

	for range iterations {
		mock := &MockTrigger{triggerChance: 50, shouldTrigger: true}
		handler := NewTriggerHandler([]Trigger{mock})

		event := twitch.EventChannelChatMessage{
			Chatter: twitch.Chatter{
				ChatterUserId:   "user123",
				ChatterUserName: "testuser",
			},
			Message: twitch.ChatMessage{
				Text: "test",
			},
		}

		handler.Process(bot, event)

		if mock.executeCalled {
			executionCounts++
		}
	}

	// With 50% chance, we expect roughly 500 executions
	// Allow 10% variance (450-550)
	minExpected := int(float64(iterations) * 0.40)
	maxExpected := int(float64(iterations) * 0.60)

	if executionCounts < minExpected || executionCounts > maxExpected {
		t.Errorf("with 50%% chance and %d iterations, got %d executions, want between %d and %d",
			iterations, executionCounts, minExpected, maxExpected)
	}

	t.Logf("50%% chance: %d/%d executions (%.1f%%)", executionCounts, iterations, float64(executionCounts)/float64(iterations)*100)
}

func TestTriggerHandler_ProbabilityLowChance(t *testing.T) {
	// Test that low probability (1%) works correctly
	bot := createTestBotForTrigger(t)

	iterations := 10000
	executionCounts := 0

	for range iterations {
		mock := &MockTrigger{triggerChance: 1, shouldTrigger: true}
		handler := NewTriggerHandler([]Trigger{mock})

		event := twitch.EventChannelChatMessage{
			Chatter: twitch.Chatter{
				ChatterUserId:   "user123",
				ChatterUserName: "testuser",
			},
			Message: twitch.ChatMessage{
				Text: "test",
			},
		}

		handler.Process(bot, event)

		if mock.executeCalled {
			executionCounts++
		}
	}

	// With 1% chance and 10000 iterations, we expect roughly 100 executions
	// Allow larger variance for low probability (50-200)
	minExpected := 50
	maxExpected := 200

	if executionCounts < minExpected || executionCounts > maxExpected {
		t.Errorf("with 1%% chance and %d iterations, got %d executions, want between %d and %d",
			iterations, executionCounts, minExpected, maxExpected)
	}

	t.Logf("1%% chance: %d/%d executions (%.2f%%)", executionCounts, iterations, float64(executionCounts)/float64(iterations)*100)
}

func TestTriggerHandler_ProbabilityHighChance(t *testing.T) {
	// Test that high probability (99%) works correctly
	bot := createTestBotForTrigger(t)

	iterations := 1000
	executionCounts := 0

	for range iterations {
		mock := &MockTrigger{triggerChance: 99, shouldTrigger: true}
		handler := NewTriggerHandler([]Trigger{mock})

		event := twitch.EventChannelChatMessage{
			Chatter: twitch.Chatter{
				ChatterUserId:   "user123",
				ChatterUserName: "testuser",
			},
			Message: twitch.ChatMessage{
				Text: "test",
			},
		}

		handler.Process(bot, event)

		if mock.executeCalled {
			executionCounts++
		}
	}

	// With 99% chance, we expect at least 970 executions (allow 3% variance)
	minExpected := 970

	if executionCounts < minExpected {
		t.Errorf("with 99%% chance and %d iterations, got %d executions, want at least %d",
			iterations, executionCounts, minExpected)
	}

	t.Logf("99%% chance: %d/%d executions (%.1f%%)", executionCounts, iterations, float64(executionCounts)/float64(iterations)*100)
}

func TestTriggerHandler_ProcessEmptyHandler(t *testing.T) {
	bot := createTestBotForTrigger(t)
	handler := NewTriggerHandler([]Trigger{})

	event := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "test",
		},
	}

	// Should not panic with empty trigger list
	handler.Process(bot, event)
}

func TestTriggerHandler_ProcessNilTriggers(t *testing.T) {
	bot := createTestBotForTrigger(t)
	handler := NewTriggerHandler(nil)

	event := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "test",
		},
	}

	// Should not panic with nil trigger list
	handler.Process(bot, event)
}

func TestTriggerHandler_MessageContent(t *testing.T) {
	// Test that triggers can access message content
	bot := createTestBotForTrigger(t)

	mock := &MockTrigger{
		triggerChance: 100,
		shouldTriggerFn: func(event twitch.EventChannelChatMessage) bool {
			return event.Message.Text == "trigger me"
		},
	}

	handler := NewTriggerHandler([]Trigger{mock})

	// Message that should trigger
	event1 := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "trigger me",
		},
	}

	handler.Process(bot, event1)

	if !mock.executeCalled {
		t.Error("trigger should have executed for matching message")
	}

	// Reset
	mock.executeCalled = false

	// Message that should not trigger
	event2 := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "don't trigger",
		},
	}

	handler.Process(bot, event2)

	if mock.executeCalled {
		t.Error("trigger should not have executed for non-matching message")
	}
}

// Helper function to create a test bot for trigger tests
func createTestBotForTrigger(t *testing.T) *Bot {
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

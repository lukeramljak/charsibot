package bot

import (
	"context"
	"log/slog"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/config"
)

type MockCommand struct {
	modOnly       bool
	triggerCmd    string
	executeCalled bool
	executeEvent  *twitch.EventChannelChatMessage
}

func (m *MockCommand) ModeratorOnly() bool {
	return m.modOnly
}

func (m *MockCommand) ShouldTrigger(command string) bool {
	return command == m.triggerCmd
}

func (m *MockCommand) Execute(bot *Bot, event twitch.EventChannelChatMessage) {
	m.executeCalled = true
	m.executeEvent = &event
}

func TestIsModerator(t *testing.T) {
	tests := []struct {
		name   string
		badges []twitch.ChatMessageUserBadge
		want   bool
	}{
		{
			name: "broadcaster badge",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "broadcaster", Id: "1", Info: ""},
			},
			want: true,
		},
		{
			name: "moderator badge",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "moderator", Id: "1", Info: ""},
			},
			want: true,
		},
		{
			name: "lead_moderator badge",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "lead_moderator", Id: "1", Info: ""},
			},
			want: true,
		},
		{
			name: "subscriber badge only",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "subscriber", Id: "12", Info: ""},
			},
			want: false,
		},
		{
			name: "vip badge only",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "vip", Id: "1", Info: ""},
			},
			want: false,
		},
		{
			name:   "no badges",
			badges: []twitch.ChatMessageUserBadge{},
			want:   false,
		},
		{
			name: "multiple badges with moderator",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "subscriber", Id: "12", Info: ""},
				{SetId: "moderator", Id: "1", Info: ""},
				{SetId: "vip", Id: "1", Info: ""},
			},
			want: true,
		},
		{
			name: "multiple non-mod badges",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "subscriber", Id: "12", Info: ""},
				{SetId: "vip", Id: "1", Info: ""},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := twitch.EventChannelChatMessage{
				Badges: tt.badges,
			}

			got := IsModerator(event)
			if got != tt.want {
				t.Errorf("IsModerator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommandHandler_Process(t *testing.T) {
	tests := []struct {
		name              string
		message           string
		badges            []twitch.ChatMessageUserBadge
		commands          []Command
		expectExecuted    bool
		expectModWarning  bool
		expectedExecIndex int // which command should execute
	}{
		{
			name:    "non-command message ignored",
			message: "hello world",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted: false,
		},
		{
			name:    "command with only exclamation mark",
			message: "!",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted: false,
		},
		{
			name:    "executes matching command",
			message: "!test",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "command with arguments",
			message: "!test arg1 arg2",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "case insensitive command",
			message: "!TEST",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "no matching command",
			message: "!unknown",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted: false,
		},
		{
			name:    "moderator can execute mod-only command",
			message: "!modcmd",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "moderator", Id: "1", Info: ""},
			},
			commands: []Command{
				&MockCommand{modOnly: true, triggerCmd: "modcmd"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "broadcaster can execute mod-only command",
			message: "!modcmd",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "broadcaster", Id: "1", Info: ""},
			},
			commands: []Command{
				&MockCommand{modOnly: true, triggerCmd: "modcmd"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "non-moderator blocked from mod-only command",
			message: "!modcmd",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: true, triggerCmd: "modcmd"},
			},
			expectExecuted:   false,
			expectModWarning: true,
		},
		{
			name:    "first matching command executes",
			message: "!test",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
		{
			name:    "skips non-matching commands",
			message: "!second",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "first"},
				&MockCommand{modOnly: false, triggerCmd: "second"},
				&MockCommand{modOnly: false, triggerCmd: "third"},
			},
			expectExecuted:    true,
			expectedExecIndex: 1,
		},
		{
			name:    "extra whitespace in message",
			message: "!test   arg1   arg2",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: []Command{
				&MockCommand{modOnly: false, triggerCmd: "test"},
			},
			expectExecuted:    true,
			expectedExecIndex: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test bot
			bot := createTestBot(t)

			// Reset mock commands
			for _, cmd := range tt.commands {
				if mock, ok := cmd.(*MockCommand); ok {
					mock.executeCalled = false
					mock.executeEvent = nil
				}
			}

			handler := NewCommandHandler(tt.commands)

			event := twitch.EventChannelChatMessage{
				Chatter: twitch.Chatter{
					ChatterUserId:   "user123",
					ChatterUserName: "testuser",
				},
				MessageId: "msg123",
				Badges:    tt.badges,
				Message: twitch.ChatMessage{
					Text: tt.message,
				},
			}

			handler.Process(bot, event)

			// Check if expected command was executed
			if tt.expectExecuted {
				mock := tt.commands[tt.expectedExecIndex].(*MockCommand)
				if !mock.executeCalled {
					t.Errorf("expected command at index %d to be executed, but it wasn't", tt.expectedExecIndex)
				}
				if mock.executeEvent == nil {
					t.Error("execute was called but event was nil")
				} else if mock.executeEvent.ChatterUserName != event.ChatterUserName {
					t.Errorf("event user = %q, want %q", mock.executeEvent.ChatterUserName, event.ChatterUserName)
				}

				// Verify other commands were not executed
				for i, cmd := range tt.commands {
					if i != tt.expectedExecIndex {
						if mock, ok := cmd.(*MockCommand); ok && mock.executeCalled {
							t.Errorf("command at index %d should not have been executed", i)
						}
					}
				}
			} else {
				// Verify no commands were executed
				for i, cmd := range tt.commands {
					if mock, ok := cmd.(*MockCommand); ok && mock.executeCalled {
						t.Errorf("command at index %d should not have been executed", i)
					}
				}
			}
		})
	}
}

func TestCommandHandler_ProcessEmptyHandler(t *testing.T) {
	bot := createTestBot(t)
	handler := NewCommandHandler([]Command{})

	event := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "!test",
		},
	}

	// Should not panic with empty command list
	handler.Process(bot, event)
}

func TestCommandHandler_ProcessNilCommands(t *testing.T) {
	bot := createTestBot(t)
	handler := NewCommandHandler(nil)

	event := twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{
			ChatterUserId:   "user123",
			ChatterUserName: "testuser",
		},
		Message: twitch.ChatMessage{
			Text: "!test",
		},
	}

	// Should not panic with nil command list
	handler.Process(bot, event)
}

// Helper function to create a test bot
func createTestBot(t *testing.T) *Bot {
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

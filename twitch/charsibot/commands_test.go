package charsibot

import (
	"context"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/twitch/db"
)

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

func TestProcessCommand(t *testing.T) {
	tests := []struct {
		name             string
		message          string
		badges           []twitch.ChatMessageUserBadge
		commands         map[string]Command
		expectExecuted   string // which command key should execute, empty if none
		expectModWarning bool
	}{
		{
			name:    "non-command message ignored",
			message: "hello world",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: map[string]Command{
				"test": {Execute: func(_ *Bot, _ twitch.EventChannelChatMessage) {}},
			},
		},
		{
			name:    "command with only exclamation mark",
			message: "!",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: map[string]Command{
				"test": {Execute: func(_ *Bot, _ twitch.EventChannelChatMessage) {}},
			},
		},
		{
			name:           "executes matching command",
			message:        "!test",
			badges:         []twitch.ChatMessageUserBadge{},
			expectExecuted: "test",
			commands: map[string]Command{
				"test": {},
			},
		},
		{
			name:           "command with arguments",
			message:        "!test arg1 arg2",
			badges:         []twitch.ChatMessageUserBadge{},
			expectExecuted: "test",
			commands: map[string]Command{
				"test": {},
			},
		},
		{
			name:           "case insensitive command",
			message:        "!TEST",
			badges:         []twitch.ChatMessageUserBadge{},
			expectExecuted: "test",
			commands: map[string]Command{
				"test": {},
			},
		},
		{
			name:    "no matching command",
			message: "!unknown",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: map[string]Command{
				"test": {},
			},
		},
		{
			name:    "moderator can execute mod-only command",
			message: "!modcmd",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "moderator", Id: "1", Info: ""},
			},
			expectExecuted: "modcmd",
			commands: map[string]Command{
				"modcmd": {ModeratorOnly: true},
			},
		},
		{
			name:    "broadcaster can execute mod-only command",
			message: "!modcmd",
			badges: []twitch.ChatMessageUserBadge{
				{SetId: "broadcaster", Id: "1", Info: ""},
			},
			expectExecuted: "modcmd",
			commands: map[string]Command{
				"modcmd": {ModeratorOnly: true},
			},
		},
		{
			name:    "non-moderator blocked from mod-only command",
			message: "!modcmd",
			badges:  []twitch.ChatMessageUserBadge{},
			commands: map[string]Command{
				"modcmd": {ModeratorOnly: true},
			},
			expectModWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed := map[string]bool{}

			commands := make(map[string]Command, len(tt.commands))
			for key, cmd := range tt.commands {
				k := key
				commands[k] = Command{
					ModeratorOnly: cmd.ModeratorOnly,
					Execute: func(_ *Bot, _ twitch.EventChannelChatMessage) {
						executed[k] = true
					},
				}
			}

			b := createTestBot(t)
			b.commands = commands

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

			b.processCommand(event)

			if tt.expectExecuted != "" {
				if !executed[tt.expectExecuted] {
					t.Errorf("expected command %q to execute, but it didn't", tt.expectExecuted)
				}
			}

			for key := range executed {
				if key != tt.expectExecuted {
					t.Errorf("command %q should not have executed", key)
				}
			}
		})
	}
}

func TestProcessCommand_Empty(t *testing.T) {
	b := createTestBot(t)
	b.commands = map[string]Command{}

	event := twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "!test"},
	}
	b.processCommand(event)
}

func TestProcessCommand_Nil(t *testing.T) {
	b := createTestBot(t)
	b.commands = nil

	event := twitch.EventChannelChatMessage{
		Message: twitch.ChatMessage{Text: "!test"},
	}
	b.processCommand(event)
}

// Helper function to create a test bot.
func createTestBot(t *testing.T) *Bot {
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

func TestStatsCommandAddSetRm(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	if _, err := GetOrCreateStats(ctx, queries, "target1", "targetuser"); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	mention := twitch.ChatMessageFragment{
		Type: "mention",
		Mention: &twitch.ChatMessageFragmentMention{
			UserID:    "target1",
			UserLogin: "targetuser",
			UserName:  "targetuser",
		},
	}

	makeEvent := func(msg string) twitch.EventChannelChatMessage {
		return twitch.EventChannelChatMessage{
			Chatter: twitch.Chatter{
				ChatterUserId:   "mod1",
				ChatterUserName: "moduser",
			},
			Badges: []twitch.ChatMessageUserBadge{{SetId: "moderator", Id: "1"}},
			Message: twitch.ChatMessage{
				Text:      msg,
				Fragments: []twitch.ChatMessageFragment{mention},
			},
		}
	}

	statValue := func(t *testing.T, userID, stat string) int64 {
		t.Helper()
		stats, err := queries.GetUserStats(ctx, userID)
		if err != nil {
			t.Fatalf("GetUserStats failed: %v", err)
		}
		for _, s := range stats {
			if s.Name == stat {
				return s.Value
			}
		}
		t.Fatalf("stat %q not found", stat)
		return 0
	}

	b := &Bot{
		config:   Config{BotUserID: "bot1", ChannelUserID: "ch1"},
		ctx:      ctx,
		store:    queries,
		commands: Commands(nil),
	}

	t.Run("add increases stat by given amount", func(t *testing.T) {
		before := statValue(t, "target1", "strength")
		b.processCommand(makeEvent("!stats add @targetuser strength 5"))
		after := statValue(t, "target1", "strength")
		if after != before+5 {
			t.Errorf("strength = %d, want %d", after, before+5)
		}
	})

	t.Run("set sets a stat to a given amount", func(t *testing.T) {
		b.processCommand(makeEvent("!stats set @targetuser strength 20"))
		after := statValue(t, "target1", "strength")
		if after != 20 {
			t.Errorf("strength = %d, want %d", after, 20)
		}
	})

	t.Run("rm decreases stat by given amount", func(t *testing.T) {
		before := statValue(t, "target1", "strength")
		b.processCommand(makeEvent("!stats rm @targetuser strength 2"))
		after := statValue(t, "target1", "strength")
		if after != before-2 {
			t.Errorf("strength = %d, want %d", after, before-2)
		}
	})

	t.Run("rm with negative amount adds instead", func(t *testing.T) {
		// !stats rm @user strength -3 → negates -3 to +3
		before := statValue(t, "target1", "strength")
		b.processCommand(makeEvent("!stats rm @targetuser strength -3"))
		after := statValue(t, "target1", "strength")
		if after != before+3 {
			t.Errorf("strength = %d, want %d", after, before+3)
		}
	})
}

func TestStatsCommandValidation(t *testing.T) {
	// Validation paths all return before touching the db.
	// Using a nil store: if the guard fails the code reaches b.store and panics,
	// so no panic = correct early return.

	modBadges := []twitch.ChatMessageUserBadge{{SetId: "moderator", Id: "1"}}
	mention := twitch.ChatMessageFragment{
		Type: "mention",
		Mention: &twitch.ChatMessageFragmentMention{
			UserID: "t1", UserLogin: "u1", UserName: "u1",
		},
	}

	makeBot := func() *Bot {
		return &Bot{
			config:   Config{BotUserID: "bot1", ChannelUserID: "ch1"},
			ctx:      context.Background(),
			commands: Commands(nil),
		}
	}

	t.Run("too few parts returns early", func(_ *testing.T) {
		makeBot().processCommand(twitch.EventChannelChatMessage{
			Badges:  modBadges,
			Message: twitch.ChatMessage{Text: "!stats add @user"},
		})
	})

	t.Run("unknown subcommand returns early", func(_ *testing.T) {
		makeBot().processCommand(twitch.EventChannelChatMessage{
			Badges: modBadges,
			Message: twitch.ChatMessage{
				Text:      "!stats unknown @user strength 5",
				Fragments: []twitch.ChatMessageFragment{mention},
			},
		})
	})

	t.Run("missing mention returns early", func(_ *testing.T) {
		makeBot().processCommand(twitch.EventChannelChatMessage{
			Badges:  modBadges,
			Message: twitch.ChatMessage{Text: "!stats add @user strength 5"},
		})
	})

	t.Run("invalid amount returns early", func(_ *testing.T) {
		makeBot().processCommand(twitch.EventChannelChatMessage{
			Badges: modBadges,
			Message: twitch.ChatMessage{
				Text:      "!stats add @user strength notanumber",
				Fragments: []twitch.ChatMessageFragment{mention},
			},
		})
	})
}

func newTestServer() (*Server, chan OverlayEvent) {
	s := &Server{clients: make(map[chan OverlayEvent]struct{})}
	ch := make(chan OverlayEvent, 10)
	s.clients[ch] = struct{}{}
	return s, ch
}

func drainEvents(ch chan OverlayEvent) []OverlayEvent {
	var events []OverlayEvent
	for {
		select {
		case e := <-ch:
			events = append(events, e)
		default:
			return events
		}
	}
}

func TestSeriesCommandRegistered(t *testing.T) {
	cfg := SeriesConfig{
		BlindBoxSeries: db.BlindBoxSeries{Series: "coobubu", RedemptionTitle: "Cooper Series Blind Box"},
	}

	cmds := Commands([]SeriesConfig{cfg})

	if _, ok := cmds["coobubu"]; !ok {
		t.Error("expected command \"coobubu\" to be registered")
	}
}

func TestSeriesCommandShowCollection(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
		UserID: "user1", Username: "alice", Series: "coobubu", Key: "cutey",
	})

	srv, ch := newTestServer()
	cfg := SeriesConfig{
		BlindBoxSeries: db.BlindBoxSeries{Series: "coobubu", RedemptionTitle: "Cooper Series Blind Box"},
	}

	b := &Bot{
		config:   Config{BotUserID: "bot1", ChannelUserID: "ch1"},
		ctx:      ctx,
		store:    queries,
		commands: Commands([]SeriesConfig{cfg}),
		server:   srv,
	}

	tests := []struct {
		name    string
		message string
	}{
		{"no subcommand shows collection", "!coobubu"},
		{"unknown subcommand falls through to show", "!coobubu foobar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b.processCommand(twitch.EventChannelChatMessage{
				Chatter: twitch.Chatter{ChatterUserId: "user1", ChatterUserName: "alice"},
				Message: twitch.ChatMessage{Text: tt.message},
			})
			events := drainEvents(ch)
			if len(events) != 1 {
				t.Fatalf("expected 1 overlay event, got %d", len(events))
			}
			if events[0].Type != EventTypeCollectionDisplay {
				t.Errorf("event type = %q, want %q", events[0].Type, EventTypeCollectionDisplay)
			}
		})
	}
}

func TestSeriesCommandReset(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()

	for _, key := range []string{"cutey", "blueberry", "secret"} {
		queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
			UserID: "user1", Username: "alice", Series: "coobubu", Key: key,
		})
	}

	cfg := SeriesConfig{
		BlindBoxSeries: db.BlindBoxSeries{Series: "coobubu", RedemptionTitle: "Cooper Series Blind Box"},
	}

	b := &Bot{
		config:   Config{BotUserID: "bot1", ChannelUserID: "ch1"},
		ctx:      ctx,
		store:    queries,
		commands: Commands([]SeriesConfig{cfg}),
	}

	b.processCommand(twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{ChatterUserId: "mod1", ChatterUserName: "moduser"},
		Badges:  []twitch.ChatMessageUserBadge{{SetId: "moderator", Id: "1"}},
		Message: twitch.ChatMessage{Text: "!coobubu reset"},
	})

	keys, err := queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: "user1", Series: "coobubu",
	})
	if err != nil {
		t.Fatalf("GetCollectedPlushies failed: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected collection untouched (reset targets mod, not user1), got %d plushies", len(keys))
	}

	b.processCommand(twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{ChatterUserId: "user1", ChatterUserName: "alice"},
		Badges:  []twitch.ChatMessageUserBadge{{SetId: "moderator", Id: "1"}},
		Message: twitch.ChatMessage{Text: "!coobubu reset"},
	})

	keys, err = queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: "user1", Series: "coobubu",
	})
	if err != nil {
		t.Fatalf("GetCollectedPlushies failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty collection after reset, got %d plushies", len(keys))
	}
}

func TestBlindboxModGuard(t *testing.T) {
	// The mod guard inside !blindbox redeem and !blindbox reset returns before any
	// store access. Using a nil store: if the guard is missing the code calls
	// b.store and panics, so no panic = guard is in place.

	cfg := SeriesConfig{
		BlindBoxSeries: db.BlindBoxSeries{Series: "test", RedemptionTitle: "Test"},
	}

	makeBot := func() *Bot {
		return &Bot{
			config:   Config{BotUserID: "bot1", ChannelUserID: "ch1"},
			ctx:      context.Background(),
			commands: Commands([]SeriesConfig{cfg}),
		}
	}

	nonModEvent := func(msg string) twitch.EventChannelChatMessage {
		return twitch.EventChannelChatMessage{
			Chatter: twitch.Chatter{ChatterUserId: "u1", ChatterUserName: "user1"},
			Message: twitch.ChatMessage{Text: msg},
		}
	}

	t.Run("non-mod cannot use redeem subcommand", func(_ *testing.T) {
		makeBot().processCommand(nonModEvent("!test redeem"))
	})

	t.Run("non-mod cannot use reset subcommand", func(_ *testing.T) {
		makeBot().processCommand(nonModEvent("!test reset"))
	})
}

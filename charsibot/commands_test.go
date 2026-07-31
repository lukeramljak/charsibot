package charsibot

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/joeyak/go-twitch-eventsub/v3"

	"github.com/lukeramljak/charsibot/blindbox"
	"github.com/lukeramljak/charsibot/catalog"
	"github.com/lukeramljak/charsibot/db"
	"github.com/lukeramljak/charsibot/server"
)

func testCatalog(t *testing.T) catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}
	return cat
}

func TestProcessCommand(t *testing.T) {
	executed := false
	b := createTestBot(t)
	b.commands = map[string]Command{
		"test": {Execute: func(_ context.Context, _ *Bot, _ twitch.EventChannelChatMessage) { executed = true }},
	}

	b.processCommand(twitch.EventChannelChatMessage{Message: twitch.ChatMessage{Text: "!TEST arg"}})
	if !executed {
		t.Error("expected matching command to execute")
	}
}

func TestProcessCommandIgnoresUnknownAndEmptyCommands(t *testing.T) {
	b := createTestBot(t)
	b.commands = map[string]Command{}
	for _, text := range []string{"hello", "!", "!unknown"} {
		b.processCommand(twitch.EventChannelChatMessage{Message: twitch.ChatMessage{Text: text}})
	}
}

func TestResolveConnectResult(t *testing.T) {
	t.Run("returns connect error when no reconnect requested", func(t *testing.T) {
		b := createTestBot(t)
		reconnectCh := make(chan error, 1)
		connectErr := errors.New("connect failed")
		if got := b.resolveConnectResult(connectErr, reconnectCh); !errors.Is(got, connectErr) {
			t.Fatalf("resolveConnectResult() = %v, want %v", got, connectErr)
		}
	})

	t.Run("prefers reconnect request over connect result", func(t *testing.T) {
		b := createTestBot(t)
		reconnectCh := make(chan error, 1)
		reconnectErr := errors.New("reconnect")
		reconnectCh <- reconnectErr
		if got := b.resolveConnectResult(errors.New("connection closed"), reconnectCh); !errors.Is(got, reconnectErr) {
			t.Fatalf("resolveConnectResult() = %v, want %v", got, reconnectErr)
		}
	})
}

func createTestBot(t *testing.T) *Bot {
	t.Helper()
	return &Bot{
		config: Config{BotUserID: "bot123", ChannelUserID: "channel456"},
		logger: slog.New(slog.DiscardHandler),
	}
}

func newBroadcast() (func(server.OverlayEvent), chan server.OverlayEvent) {
	ch := make(chan server.OverlayEvent, 10)
	return func(event server.OverlayEvent) { ch <- event }, ch
}

func TestSeriesCommandRegistered(t *testing.T) {
	cmds := Commands([]blindbox.SeriesConfig{{Series: "coobubu"}})
	if _, ok := cmds["coobubu"]; !ok {
		t.Error("expected series command to be registered")
	}
}

func TestSeriesCommandShowsCollection(t *testing.T) {
	queries, sqlDB := db.NewTestDB(t)
	defer sqlDB.Close()
	ctx := context.Background()
	appCatalog := testCatalog(t)
	service, err := blindbox.NewService(queries, appCatalog.Series)
	if err != nil {
		t.Fatal(err)
	}
	if err := queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
		UserID: "user1", Username: "alice", Series: "coobubu", Key: "cutey",
	}); err != nil {
		t.Fatal(err)
	}

	broadcast, events := newBroadcast()
	b := &Bot{
		logger:          slog.New(slog.DiscardHandler),
		blindboxService: service,
		commands:        Commands([]blindbox.SeriesConfig{{Series: "coobubu"}}),
		broadcast:       broadcast,
	}
	b.processCommand(twitch.EventChannelChatMessage{
		Chatter: twitch.Chatter{ChatterUserId: "user1", ChatterUserName: "alice"},
		Message: twitch.ChatMessage{Text: "!coobubu"},
	})

	select {
	case event := <-events:
		if event.Type != server.EventTypeCollectionDisplay {
			t.Errorf("event type = %q, want %q", event.Type, server.EventTypeCollectionDisplay)
		}
	default:
		t.Error("expected collection display event")
	}
}

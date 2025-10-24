package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/nicklaw5/helix/v2"
	twitchws "github.com/vpetrigo/go-twitch-ws"
	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

type Config struct {
	ClientID          string
	ClientSecret      string
	OAuthToken        string // Streamer's access token (for EventSub subscriptions)
	RefreshToken      string // Streamer's refresh token
	BotOAuthToken     string // Bot's access token (for sending chat messages)
	BotRefreshToken   string // Bot's refresh token
	BotUserID         string
	ChatChannelUserID string
	DbURL             string
	DbAuthToken       string
}

type CommandHandler func(context.Context, *eventsub.ChannelChatMessage) (string, error)
type RewardHandler func(context.Context, *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error

type Bot struct {
	store          *Store
	helixClient    *helix.Client // Authenticated as streamer (for EventSub)
	botHelixClient *helix.Client // Authenticated as bot (for sending messages)
	cfg            *Config
	rng            *rand.Rand

	commands map[string]CommandHandler
	rewards  map[string]RewardHandler
}

func NewBot(store *Store, cfg *Config) *Bot {
	b := &Bot{
		store:    store,
		cfg:      cfg,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		commands: make(map[string]CommandHandler),
		rewards:  make(map[string]RewardHandler),
	}

	b.commands = map[string]CommandHandler{
		"!stats":   b.onStatsCommand,
		"!addstat": b.onModifyStatCommand,
		"!rmstat":  b.onModifyStatCommand,
	}

	b.rewards = map[string]RewardHandler{
		"Drink a Potion": b.onDrinkPotion,
		"Tempt the Dice": b.onTemptDice,
	}

	return b
}

func (b *Bot) Connect(ctx context.Context, websocketUrl string) error {
	client := twitchws.NewClient(
		websocketUrl,
		twitchws.WithOnWelcome(b.onWelcomeEvent),
		twitchws.WithOnNotification(b.onNotificationEvent),
		twitchws.WithOnConnect(func() { slog.Info("connected to Twitch websocket") }),
		twitchws.WithOnDisconnect(func() { slog.Warn("disconnected from Twitch websocket") }),
		twitchws.WithOnRevocation(func(_ *twitchws.Metadata, payload *twitchws.Payload) {
			slog.Error("EventSub subscription revoked", "payload", payload)
		}),
		twitchws.WithOnReconnect(func(metadata *twitchws.Metadata, payload *twitchws.Payload) {
			slog.Warn("websocket reconnecting", "metadata", metadata, "payload", payload)
		}))

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- client.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = client.Close()
		return ctx.Err()
	case err := <-waitErrCh:
		if err != nil {
			slog.Error("wait error", "err", err)
		}
		_ = client.Close()
		return err
	}
}

func (b *Bot) onWelcomeEvent(_ *twitchws.Metadata, payload *twitchws.Payload) {
	session, _ := payload.Payload.(twitchws.Session)
	slog.Info("received welcome event", "session_id", session.ID)

	ctx := context.Background()

	if err := b.store.InitTokenSchema(ctx); err != nil {
		slog.Error("init token schema failed", "error", err)
		return
	}

	streamerAccessToken := b.cfg.OAuthToken
	streamerRefreshToken := b.cfg.RefreshToken
	if tokens, err := b.store.GetTokens(ctx, "streamer"); err == nil && tokens != nil {
		slog.Info("loaded streamer tokens from database")
		streamerAccessToken = tokens.AccessToken
		streamerRefreshToken = tokens.RefreshToken
	} else {
		slog.Info("using streamer tokens from environment")
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        b.cfg.ClientID,
		ClientSecret:    b.cfg.ClientSecret,
		UserAccessToken: streamerAccessToken,
		RefreshToken:    streamerRefreshToken,
	})
	if err != nil {
		slog.Error("create streamer helix client failed", "error", err)
		return
	}

	helixClient.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
		slog.Info("streamer tokens auto-refreshed")
		if err := b.store.SaveTokens(context.Background(), "streamer", newAccessToken, newRefreshToken); err != nil {
			slog.Error("failed to save streamer tokens during auto-refresh", "error", err)
		}
	})

	b.helixClient = helixClient
	slog.Info("streamer helix client initialised")

	botAccessToken := b.cfg.BotOAuthToken
	botRefreshToken := b.cfg.BotRefreshToken
	if tokens, err := b.store.GetTokens(ctx, "bot"); err == nil && tokens != nil {
		slog.Info("loaded bot tokens from database")
		botAccessToken = tokens.AccessToken
		botRefreshToken = tokens.RefreshToken
	} else {
		slog.Info("using bot tokens from environment")
	}

	botHelixClient, err := helix.NewClient(&helix.Options{
		ClientID:        b.cfg.ClientID,
		ClientSecret:    b.cfg.ClientSecret,
		UserAccessToken: botAccessToken,
		RefreshToken:    botRefreshToken,
	})
	if err != nil {
		slog.Error("create bot helix client failed", "error", err)
		return
	}

	botHelixClient.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
		slog.Info("bot tokens auto-refreshed")
		if err := b.store.SaveTokens(context.Background(), "bot", newAccessToken, newRefreshToken); err != nil {
			slog.Error("failed to save bot tokens during auto-refresh", "error", err)
		}
	})

	b.botHelixClient = botHelixClient
	slog.Info("bot helix client initialised")

	transport := helix.EventSubTransport{Method: "websocket", SessionID: session.ID}
	subs := []helix.EventSubSubscription{
		{
			Type:      helix.EventSubTypeChannelChatMessage,
			Version:   "1",
			Condition: helix.EventSubCondition{BroadcasterUserID: b.cfg.ChatChannelUserID, UserID: b.cfg.ChatChannelUserID},
			Transport: transport,
		},
		{
			Type:      helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd,
			Version:   "1",
			Condition: helix.EventSubCondition{BroadcasterUserID: b.cfg.ChatChannelUserID},
			Transport: transport,
		},
	}

	for _, sub := range subs {
		resp, err := helixClient.CreateEventSubSubscription(&sub)
		if err != nil {
			slog.Error("eventsub subscription failed", "type", sub.Type, "error", err)
			continue
		}
		slog.Debug("eventsub subscribed", "type", sub.Type, "status", resp.StatusCode)
	}
}

func (b *Bot) onNotificationEvent(_ *twitchws.Metadata, payload *twitchws.Payload) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	notification := payload.Payload.(twitchws.Notification)

	switch event := notification.Event.(type) {
	case *eventsub.ChannelChatMessage:
		b.onChatMessage(ctx, event)
	case *eventsub.ChannelPointsCustomRewardRedemptionAddEvent:
		b.onRewardRedemption(ctx, event)
	}
}

func (b *Bot) onChatMessage(ctx context.Context, event *eventsub.ChannelChatMessage) {
	if event.Reply != nil {
		return
	}

	text := strings.TrimSpace(event.Message.Text)
	if len(text) == 0 || text[0] != '!' {
		return
	}

	cmd, _, _ := strings.Cut(strings.ToLower(text), " ")
	handler, ok := b.commands[cmd]
	if !ok {
		return
	}

	response, err := handler(ctx, event)
	if err != nil {
		slog.Error("command error", "command", cmd, "error", err)
		return
	}

	if err := b.send(response, event.MessageID); err != nil {
		slog.Error("command reply failed", "command", cmd, "error", err)
	} else {
		slog.Info("command handled", "command", cmd, "user", event.ChatterUserLogin, "reply", response)
	}
}

func (b *Bot) onRewardRedemption(ctx context.Context, event *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) {
	handler, ok := b.rewards[event.Reward.Title]
	if !ok {
		return
	}

	if err := handler(ctx, event); err != nil {
		slog.Error("reward handler error", "reward", event.Reward.Title, "error", err, "user", event.UserLogin)
	} else {
		slog.Info("reward handled", "reward", event.Reward.Title, "user", event.UserLogin, "cost", event.Reward.Cost)
	}
}

func (b *Bot) send(message, replyToMessageID string) error {
	if b.botHelixClient == nil {
		return fmt.Errorf("bot helix client not initialised")
	}

	params := &helix.SendChatMessageParams{
		BroadcasterID: b.cfg.ChatChannelUserID,
		Message:       message,
		SenderID:      b.cfg.BotUserID,
	}

	if replyToMessageID != "" {
		params.ReplyParentMessageID = replyToMessageID
	}

	_, err := b.botHelixClient.SendChatMessage(params)

	if err != nil && strings.Contains(err.Error(), "Failed to decode API response") {
		time.Sleep(50 * time.Millisecond)
		_, err = b.botHelixClient.SendChatMessage(params)
	}

	return err
}

func (b *Bot) getStatsMessage(ctx context.Context, userID, username string) (string, error) {
	st, err := b.store.GetStats(ctx, userID, username)
	if err != nil {
		return "", err
	}
	return FormatStats(username, st), nil
}

func (b *Bot) randomStat() Stat {
	return statList[b.rng.Intn(len(statList))]
}

func (b *Bot) randomDelta() int {
	if b.rng.Intn(20) == 0 {
		return -1
	}
	return 1
}

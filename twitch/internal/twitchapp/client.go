package twitchapp

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/lukeramljak/charsibot/twitch/internal/router"
	"github.com/lukeramljak/charsibot/twitch/internal/stats"
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

type Client struct {
	db             *sql.DB
	helixClient    *helix.Client // Authenticated as streamer (for EventSub)
	botHelixClient *helix.Client // Authenticated as bot (for sending messages)
	cfg            *Config
	statsStore     *stats.Store
	commands       *router.CommandRouter
	rewards        *router.RewardRouter
	rng            *rand.Rand
	wsClient       *twitchws.Client
}

func New(db *sql.DB, cfg *Config) *Client {
	c := &Client{
		db:         db,
		cfg:        cfg,
		statsStore: stats.NewStore(db),
		commands:   router.NewCommandRouter(),
		rewards:    router.NewRewardRouter(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	c.registerBuiltInCommands()
	c.registerBuiltInRewards()
	return c
}

func (c *Client) RegisterCommand(name string, h router.CommandHandler) {
	c.commands.Register(name, h)
}
func (c *Client) RegisterReward(title string, h router.RewardHandler) {
	c.rewards.Register(title, h)
}

func (c *Client) SendChatMessage(ctx context.Context, message string) error {
	if c.botHelixClient == nil {
		return fmt.Errorf("bot helix client not initialised")
	}
	_, err := c.botHelixClient.SendChatMessage(&helix.SendChatMessageParams{
		BroadcasterID: c.cfg.ChatChannelUserID,
		Message:       message,
		SenderID:      c.cfg.BotUserID,
	})
	if err != nil {
		slog.Error("chat send failed", "mode", "plain", "error", err)
	}
	return err
}

func (c *Client) SendReply(ctx context.Context, parentID string, message string) error {
	if c.botHelixClient == nil {
		return fmt.Errorf("bot helix client not initialised")
	}
	_, err := c.botHelixClient.SendChatMessage(&helix.SendChatMessageParams{
		BroadcasterID:        c.cfg.ChatChannelUserID,
		Message:              message,
		SenderID:             c.cfg.BotUserID,
		ReplyParentMessageID: parentID,
	})
	if err != nil {
		slog.Error("chat reply failed", "parent_id", parentID, "error", err)
	}
	return err
}

func (c *Client) sendUserStats(ctx context.Context, userID, userLogin string) error {
	msg, err := c.statsStore.GetMessage(ctx, userID, userLogin)
	if err != nil {
		return err
	}
	return c.SendChatMessage(ctx, msg)
}

func (c *Client) Connect(ctx context.Context, websocketUrl string) error {
	client := twitchws.NewClient(
		websocketUrl,
		twitchws.WithOnWelcome(c.onWelcomeEvent),
		twitchws.WithOnNotification(c.onNotificationEvent),
		twitchws.WithOnConnect(c.onConnect),
		twitchws.WithOnDisconnect(c.onDisconnect),
		twitchws.WithOnRevocation(c.onRevocationEvent),
		twitchws.WithOnReconnect(c.onReconnect))
	c.wsClient = client

	if err := client.Connect(); err != nil {
		slog.Error("connect error", "err", err)
		return err
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
			_ = client.Close()
			return err
		}
	}

	_ = client.Close()

	return nil
}

func (c *Client) onWelcomeEvent(_ *twitchws.Metadata, payload *twitchws.Payload) {
	session, _ := payload.Payload.(twitchws.Session)
	if err := c.initHelix(session.ID); err != nil {
		slog.Error("helix init failed", "error", err)
	}
}

func (c *Client) onNotificationEvent(_ *twitchws.Metadata, payload *twitchws.Payload) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	notification := payload.Payload.(twitchws.Notification)

	switch event := notification.Event.(type) {
	case *eventsub.ChannelChatMessage:
		slog.Debug("received chat message", "user", event.ChatterUserLogin, "message", event.Message.Text)
		if event.Reply != nil {
			return
		}
		if responseMessage, handled := c.commands.Handle(ctx, event); handled {
			if err := c.SendReply(ctx, event.MessageID, responseMessage); err != nil {
				slog.Error("command reply failed", "command", event.Message.Text, "error", err)
			} else {
				slog.Info("command handled", "command", event.Message.Text, "user", event.ChatterUserLogin, "user_id", event.ChatterUserID, "reply", responseMessage)
			}
		}
	case *eventsub.ChannelPointsCustomRewardRedemptionAddEvent:
		c.rewards.Handle(ctx, event)
	}
}

func (c *Client) onReconnect(metadata *twitchws.Metadata, payload *twitchws.Payload) {
	slog.Info("Reconnect", "metadata", metadata, "payload", payload)
}
func (c *Client) onRevocationEvent(_ *twitchws.Metadata, payload *twitchws.Payload) {
	slog.Info("Revocation", "payload", payload)
}
func (c *Client) onConnect() {
	slog.Info("Connected to twitch")
}
func (c *Client) onDisconnect() {
	slog.Info("Disconnected from twitch")
}

func (c *Client) initHelix(sessionID string) error {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        c.cfg.ClientID,
		ClientSecret:    c.cfg.ClientSecret,
		UserAccessToken: c.cfg.OAuthToken,
		RefreshToken:    c.cfg.RefreshToken,
		APIBaseURL:      helix.DefaultAPIBaseURL,
	})
	if err != nil {
		return fmt.Errorf("create streamer helix client: %w", err)
	}

	refresh, err := helixClient.RefreshUserAccessToken(helixClient.GetRefreshToken())
	if err != nil {
		return fmt.Errorf("refresh streamer tokens: %w", err)
	}
	helixClient.SetUserAccessToken(refresh.Data.AccessToken)
	helixClient.SetRefreshToken(refresh.Data.RefreshToken)

	c.helixClient = helixClient
	slog.Info("Streamer authenticated successfully", "expires_in", refresh.Data.ExpiresIn)

	botHelixClient, err := helix.NewClient(&helix.Options{
		ClientID:        c.cfg.ClientID,
		ClientSecret:    c.cfg.ClientSecret,
		UserAccessToken: c.cfg.BotOAuthToken,
		RefreshToken:    c.cfg.BotRefreshToken,
		APIBaseURL:      helix.DefaultAPIBaseURL,
	})
	if err != nil {
		return fmt.Errorf("create bot helix client: %w", err)
	}

	botRefresh, err := botHelixClient.RefreshUserAccessToken(botHelixClient.GetRefreshToken())
	if err != nil {
		return fmt.Errorf("refresh bot tokens: %w", err)
	}
	botHelixClient.SetUserAccessToken(botRefresh.Data.AccessToken)
	botHelixClient.SetRefreshToken(botRefresh.Data.RefreshToken)

	c.botHelixClient = botHelixClient
	slog.Info("Bot authenticated successfully", "expires_in", botRefresh.Data.ExpiresIn)

	// Create EventSub subscriptions using the streamer's authenticated client
	transport := helix.EventSubTransport{Method: "websocket", SessionID: sessionID}
	subs := []helix.EventSubSubscription{
		// Listen to all messages in the streamer's chat (UserID = the authenticated user, which is the streamer)
		{Type: helix.EventSubTypeChannelChatMessage, Version: "1", Condition: helix.EventSubCondition{BroadcasterUserID: c.cfg.ChatChannelUserID, UserID: c.cfg.ChatChannelUserID}, Transport: transport},
		{Type: helix.EventSubTypeChannelPointsCustomRewardRedemptionAdd, Version: "1", Condition: helix.EventSubCondition{BroadcasterUserID: c.cfg.ChatChannelUserID}, Transport: transport},
	}

	for _, sub := range subs {
		resp, err := helixClient.CreateEventSubSubscription(&sub)
		if err != nil {
			slog.Error("eventsub subscription failed", "type", sub.Type, "error", err, "response", resp)
			continue
		}
		slog.Debug("eventsub subscribed", "type", sub.Type, "response", resp)
	}

	return nil
}

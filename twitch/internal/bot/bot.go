package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/lukeramljak/charsibot/internal/config"
	"github.com/lukeramljak/charsibot/internal/server"
	"github.com/lukeramljak/charsibot/internal/store"
	"github.com/nicklaw5/helix/v2"
)

type Bot struct {
	config            config.Config
	store             *store.Queries
	logger            *slog.Logger
	commandHandler    *CommandHandler
	triggerHandler    *TriggerHandler
	redemptionHandler *RedemptionHandler
	twitchClient      *twitch.Client
	helixClient       *helix.Client
	botHelixClient    *helix.Client
	overlayServer     *server.Server
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	isRunning         bool
	mu                sync.RWMutex
}

type SendMessageParams struct {
	Message              string
	ReplyParentMessageID string
}

func New(cfg config.Config, queries *store.Queries, logger *slog.Logger, commandHandler *CommandHandler, triggerHandler *TriggerHandler, redemptionHandler *RedemptionHandler) (*Bot, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}
	if queries == nil {
		return nil, errors.New("store queries cannot be nil")
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Bot{
		config:            cfg,
		store:             queries,
		logger:            logger,
		commandHandler:    commandHandler,
		triggerHandler:    triggerHandler,
		redemptionHandler: redemptionHandler,
		overlayServer:     server.NewServer(cfg.ServerPort, logger),
		ctx:               ctx,
		cancel:            cancel,
	}, nil
}

func (b *Bot) Start() error {
	b.mu.Lock()
	if b.isRunning {
		b.mu.Unlock()
		return errors.New("bot is already running")
	}
	b.mu.Unlock()

	url := "wss://eventsub.wss.twitch.tv/ws"

	if b.config.UseMockServer {
		url = "ws://localhost:8080/ws"
	}

	client := twitch.NewClientWithUrl(url)
	b.twitchClient = client

	client.OnError(func(err error) {
		b.logger.Error("twitch client error", "err", err)
	})

	client.OnWelcome(func(message twitch.WelcomeMessage) {
		b.logger.Info("connected to twitch eventsub", "session_id", message.Payload.Session.ID)
		if err := b.init(message); err != nil {
			b.logger.Error("failed to initialize bot", "err", err)
			b.Shutdown()
		}
	})

	client.OnNotification(func(message twitch.NotificationMessage) {
		b.logger.Debug("eventsub notification", "type", message.Payload.Subscription.Type)
	})

	client.OnKeepAlive(func(message twitch.KeepAliveMessage) {
		b.logger.Debug("keepalive received")
	})

	client.OnRevoke(func(message twitch.RevokeMessage) {
		b.logger.Warn("subscription revoked", "subscription", message.Payload.Subscription)
	})

	client.OnRawEvent(func(event string, metadata twitch.MessageMetadata, subscription twitch.PayloadSubscription) {
		b.logger.Debug("raw event", "type", subscription.Type)
	})

	client.OnEventChannelChatMessage(func(event twitch.EventChannelChatMessage) {
		b.wg.Go(func() {
			b.onMessage(event)
		})
	})

	client.OnEventChannelChannelPointsCustomRewardRedemptionAdd(func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
		b.wg.Go(func() {
			b.onChannelPointRedemption(event)
		})
	})

	client.OnEventChannelRaid(func(event twitch.EventChannelRaid) {
		userName := event.FromBroadcasterUserName
		b.SendMessage(SendMessageParams{
			Message: fmt.Sprintf("!so @%s", userName),
		})
	})

	b.mu.Lock()
	b.isRunning = true
	b.mu.Unlock()

	if err := client.Connect(); err != nil {
		b.mu.Lock()
		b.isRunning = false
		b.mu.Unlock()
		return fmt.Errorf("failed to connect to twitch: %w", err)
	}

	b.logger.Info("bot started successfully")
	return nil
}

func (b *Bot) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.isRunning {
		return
	}

	b.logger.Info("shutting down bot")
	b.cancel()

	if b.overlayServer != nil {
		b.overlayServer.Stop()
	}

	if b.twitchClient != nil {
		if err := b.twitchClient.Close(); err != nil {
			b.logger.Error("error closing twitch client", "err", err)
		}
	}

	b.wg.Wait()
	b.isRunning = false
	b.logger.Info("bot stopped")
}

func (b *Bot) init(message twitch.WelcomeMessage) error {
	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:        b.config.ClientID,
		ClientSecret:    b.config.ClientSecret,
		UserAccessToken: b.config.StreamerAccessToken,
		APIBaseURL:      helix.DefaultAPIBaseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to create helix client: %w", err)
	}

	refresh, err := helixClient.RefreshUserAccessToken(b.config.StreamerRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh streamer token: %w", err)
	}

	helixClient.SetUserAccessToken(refresh.Data.AccessToken)
	helixClient.SetRefreshToken(refresh.Data.RefreshToken)

	if err := b.store.SaveTokens(b.ctx, store.SaveTokensParams{
		TokenType:    "streamer",
		AccessToken:  refresh.Data.AccessToken,
		RefreshToken: refresh.Data.RefreshToken,
	}); err != nil {
		b.logger.Error("failed to save streamer tokens", "err", err)
	} else {
		b.logger.Info("saved streamer tokens")
	}

	helixClient.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
		if err := b.store.SaveTokens(b.ctx, store.SaveTokensParams{
			TokenType:    "streamer",
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		}); err != nil {
			b.logger.Error("failed to save auto-refreshed streamer tokens", "err", err)
		} else {
			b.logger.Info("streamer tokens auto-refreshed")
		}
	})

	b.helixClient = helixClient

	botHelixClient, err := helix.NewClient(&helix.Options{
		ClientID:        b.config.ClientID,
		ClientSecret:    b.config.ClientSecret,
		UserAccessToken: b.config.BotAccessToken,
		APIBaseURL:      helix.DefaultAPIBaseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to create bot helix client: %w", err)
	}

	botRefresh, err := botHelixClient.RefreshUserAccessToken(b.config.BotRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh bot token: %w", err)
	}

	botHelixClient.SetUserAccessToken(botRefresh.Data.AccessToken)
	botHelixClient.SetRefreshToken(botRefresh.Data.RefreshToken)

	if err := b.store.SaveTokens(b.ctx, store.SaveTokensParams{
		TokenType:    "bot",
		AccessToken:  botRefresh.Data.AccessToken,
		RefreshToken: botRefresh.Data.RefreshToken,
	}); err != nil {
		b.logger.Error("failed to save bot tokens", "err", err)
	} else {
		b.logger.Info("saved bot tokens")
	}

	botHelixClient.OnUserAccessTokenRefreshed(func(newAccessToken, newRefreshToken string) {
		if err := b.store.SaveTokens(b.ctx, store.SaveTokensParams{
			TokenType:    "bot",
			AccessToken:  newAccessToken,
			RefreshToken: newRefreshToken,
		}); err != nil {
			b.logger.Error("failed to save auto-refreshed bot tokens", "err", err)
		} else {
			b.logger.Info("bot tokens auto-refreshed")
		}
	})

	b.botHelixClient = botHelixClient

	if !b.config.UseMockServer {
		events := []twitch.EventSubscription{
			twitch.SubChannelChatMessage,
			twitch.SubChannelChannelPointsCustomRewardRedemptionAdd,
		}

		for _, event := range events {
			b.logger.Info("subscribing to event", "event", event)
			if _, err := twitch.SubscribeEvent(twitch.SubscribeRequest{
				SessionID:   message.Payload.Session.ID,
				ClientID:    b.config.ClientID,
				AccessToken: helixClient.GetUserAccessToken(),
				Event:       event,
				Condition: map[string]string{
					"broadcaster_user_id": b.config.ChannelUserID,
					"user_id":             b.config.ChannelUserID,
				},
			}); err != nil {
				return fmt.Errorf("failed to subscribe to event %s: %w", event, err)
			}
		}
	}

	if err := b.overlayServer.Start(); err != nil {
		return fmt.Errorf("failed to start SSE server: %w", err)
	}

	b.logger.Info("bot initialized successfully")
	return nil
}

func (b *Bot) onMessage(event twitch.EventChannelChatMessage) {
	// Skip bot's own messages
	if event.ChatterUserId == b.config.BotUserID {
		return
	}

	b.logger.Debug("processing message",
		"user", event.ChatterUserName,
		"message", event.Message.Text,
	)

	if b.commandHandler != nil {
		b.commandHandler.Process(b, event)
	}
	if b.triggerHandler != nil {
		b.triggerHandler.Process(b, event)
	}
}

func (b *Bot) onChannelPointRedemption(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	b.logger.Info("channel point redemption",
		"user", event.UserName,
		"reward", event.Reward.Title,
	)

	if b.redemptionHandler != nil {
		b.redemptionHandler.Process(b, event)
	}
}

func (b *Bot) ID() string {
	return b.config.BotUserID
}

func (b *Bot) Store() *store.Queries {
	return b.store
}

func (b *Bot) Context() context.Context {
	return b.ctx
}

func (b *Bot) HelixClient() *helix.Client {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.helixClient
}

func (b *Bot) SendMessage(params SendMessageParams) {
	b.mu.RLock()
	if !b.isRunning {
		b.mu.RUnlock()
		b.logger.Error("failed to send message: bot is not running")
		return
	}
	b.mu.RUnlock()

	msgParams := &helix.SendChatMessageParams{
		SenderID:      b.config.BotUserID,
		BroadcasterID: b.config.ChannelUserID,
		Message:       params.Message,
	}

	if params.ReplyParentMessageID != "" {
		msgParams.ReplyParentMessageID = params.ReplyParentMessageID
	}

	resp, err := b.botHelixClient.SendChatMessage(msgParams)
	if err != nil {
		if resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
			b.logger.Info("token may have expired, refreshing and retrying")

			refresh, err := b.botHelixClient.RefreshUserAccessToken(b.botHelixClient.GetRefreshToken())
			if err != nil {
				b.logger.Error("failed to refresh token", "err", err)
				b.logger.Error("failed to send message", "err", err, "message", params.Message)
				return
			}

			b.botHelixClient.SetUserAccessToken(refresh.Data.AccessToken)
			b.botHelixClient.SetRefreshToken(refresh.Data.RefreshToken)

			if err = b.store.SaveTokens(b.ctx, store.SaveTokensParams{
				TokenType:    "bot",
				AccessToken:  refresh.Data.AccessToken,
				RefreshToken: refresh.Data.RefreshToken,
			}); err != nil {
				b.logger.Error("failed to save refreshed bot tokens", "err", err)
			}

			resp, err = b.botHelixClient.SendChatMessage(msgParams)
			if err != nil {
				b.logger.Error("failed to send message after token refresh", "err", err, "message", params.Message)
				return
			}
		} else {
			b.logger.Error("failed to send message", "err", err, "message", params.Message)
			return
		}
	}

	if resp.Error != "" {
		b.logger.Warn("message send warning", "error", resp.Error, "message", params.Message)
	}

	b.logger.Debug("message sent", "message", params.Message)
}

func (b *Bot) BroadcastOverlayEvent(event server.OverlayEvent) {
	if b.overlayServer != nil {
		b.overlayServer.Broadcast(event)
	}
}

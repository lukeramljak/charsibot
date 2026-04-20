package charsibot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeyak/go-twitch-eventsub/v3"
	"github.com/nicklaw5/helix/v2"

	"github.com/lukeramljak/charsibot/twitch/db"
)

type Bot struct {
	config       Config
	store        *db.Queries
	commands     map[string]Command
	redemptions  map[string]RedemptionFunc
	triggers     []Trigger
	twitchClient *twitch.Client
	helixClient  *helix.Client
	conduitID    string
	server       *Server
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	shuttingDown atomic.Bool
}

const reconnectDelay = 10 * time.Second

type SendMessageParams struct {
	Message              string
	ReplyParentMessageID string
}

func New(cfg Config, queries *db.Queries, server *Server) (*Bot, error) {
	if queries == nil {
		return nil, errors.New("store queries cannot be nil")
	}

	seriesConfigs, err := LoadAllSeries(context.Background(), queries)
	if err != nil {
		return nil, fmt.Errorf("load blind box series: %w", err)
	}

	return &Bot{
		config:      cfg,
		store:       queries,
		commands:    Commands(seriesConfigs),
		redemptions: Redemptions(seriesConfigs),
		triggers:    Triggers(),
		server:      server,
	}, nil
}

func (b *Bot) Start() error {
	b.ctx, b.cancel = context.WithCancel(context.Background())

	if err := b.initHelixClient(); err != nil {
		return fmt.Errorf("init helix client: %w", err)
	}

	conduitID, err := getOrCreateConduit(b.config.ClientID, b.helixClient.GetAppAccessToken())
	if err != nil {
		return fmt.Errorf("get/create conduit: %w", err)
	}
	b.conduitID = conduitID
	slog.Info("conduit ready", "conduit_id", conduitID)

	url := "wss://eventsub.wss.twitch.tv/ws"

	if b.config.UseMockServer {
		url = "ws://localhost:8080/ws"
	}

	for {
		if err := b.connectOnce(url); err != nil {
			if b.shuttingDown.Load() {
				return nil
			}
			slog.Error("eventsub disconnected, reconnecting", "err", err, "delay", reconnectDelay)
			select {
			case <-time.After(reconnectDelay):
			case <-b.ctx.Done():
				return nil
			}
			continue
		}
		return nil
	}
}

func (b *Bot) connectOnce(url string) error {
	client := twitch.NewClientWithUrl(url)
	b.twitchClient = client

	client.OnError(func(err error) {
		slog.Error("twitch client error", "err", err)
	})

	client.OnWelcome(func(message twitch.WelcomeMessage) {
		slog.Info("connected to twitch eventsub", "session_id", message.Payload.Session.ID)
		if err := b.subscribeEvents(message.Payload.Session.ID); err != nil {
			slog.Error("failed to subscribe to events", "err", err)
			b.Shutdown()
		}
	})

	client.OnNotification(func(message twitch.NotificationMessage) {
		slog.Debug("eventsub notification", "type", message.Payload.Subscription.Type)
	})

	client.OnRevoke(func(message twitch.RevokeMessage) {
		slog.Warn("subscription revoked", "subscription", message.Payload.Subscription)
	})

	client.OnReconnect(func(message twitch.ReconnectMessage) {
		slog.Debug("client reconnected", "msg", message)
	})

	client.OnEventConduitShardDisabled(func(event twitch.EventConduitShardDisabled) {
		slog.Warn("conduit shard disabled, reconnecting",
			"conduit_id", event.ConduitId,
			"shard_id", event.ShardId,
			"status", event.Status,
		)
		if err := client.Close(); err != nil {
			slog.Error("error closing client after shard disabled", "err", err)
		}
	})

	client.OnEventChannelChatMessage(func(event twitch.EventChannelChatMessage) {
		b.wg.Go(func() {
			b.onMessage(event)
		})
	})

	client.OnEventChannelChannelPointsCustomRewardRedemptionAdd(
		func(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
			b.wg.Go(func() {
				b.onChannelPointRedemption(event)
			})
		},
	)

	client.OnEventChannelRaid(func(event twitch.EventChannelRaid) {
		b.onChannelRaid(event)
	})

	return client.Connect()
}

func (b *Bot) Shutdown() {
	slog.Info("shutting down bot")

	b.shuttingDown.Store(true)
	b.cancel()
	if b.twitchClient != nil {
		if err := b.twitchClient.Close(); err != nil {
			slog.Debug("error closing twitch client", "err", err)
		}
	}

	b.wg.Wait()
	slog.Info("bot stopped")
}

func (b *Bot) onMessage(event twitch.EventChannelChatMessage) {
	if event.ChatterUserId == b.config.BotUserID {
		return
	}

	slog.Debug("processing message",
		"user", event.ChatterUserName,
		"message", event.Message.Text,
	)

	b.processCommand(event)
	b.processTriggers(event)
}

func (b *Bot) processCommand(event twitch.EventChannelChatMessage) {
	if !strings.HasPrefix(event.Message.Text, "!") {
		return
	}

	fields := strings.Fields(strings.ToLower(event.Message.Text))
	if len(fields) == 0 {
		return
	}

	cmd := strings.TrimPrefix(fields[0], "!")
	if cmd == "" {
		return
	}

	command, ok := b.commands[cmd]
	if !ok {
		return
	}

	slog.Info("chat command received",
		"command", cmd,
		"user", event.ChatterUserName,
		"message", event.Message.Text,
	)

	if command.ModeratorOnly && !IsModerator(event) {
		slog.Warn("non-moderator attempted mod command",
			"user", event.ChatterUserName,
			"command", cmd,
		)
		b.SendMessage(SendMessageParams{
			Message:              "You must be a moderator to use this command",
			ReplyParentMessageID: event.MessageId,
		})
		return
	}

	slog.Info("executing command", "command", cmd, "user", event.ChatterUserName)
	command.Execute(b, event)
}

func (b *Bot) processTriggers(event twitch.EventChannelChatMessage) {
	for _, t := range b.triggers {
		if !t.ShouldTrigger(event) {
			continue
		}

		const percentMax = 100

		if chance := t.Chance; chance > 0 && chance < 100 {
			roll := rand.IntN(percentMax) + 1
			if roll > chance {
				slog.Debug("trigger failed chance roll", "roll", roll, "chance", chance)
				continue
			}
		}

		slog.Info("executing trigger",
			"user", event.ChatterUserName,
			"message", event.Message.Text,
		)
		t.Execute(b, event)
	}
}

func (b *Bot) onChannelPointRedemption(event twitch.EventChannelChannelPointsCustomRewardRedemptionAdd) {
	slog.Info("channel point redemption",
		"user", event.UserName,
		"reward", event.Reward.Title,
	)

	fn, ok := b.redemptions[event.Reward.Title]
	if !ok {
		return
	}

	fn(b, event)
}

func (b *Bot) onChannelRaid(event twitch.EventChannelRaid) {
	userName := event.FromBroadcasterUserName
	b.SendMessage(SendMessageParams{
		Message: fmt.Sprintf("!so @%s", userName),
	})
}

func (b *Bot) initHelixClient() error {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     b.config.ClientID,
		ClientSecret: b.config.ClientSecret,
	})
	if err != nil {
		return fmt.Errorf("create app helix client: %w", err)
	}

	resp, err := client.RequestAppAccessToken(nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get app access token: %w", err)
	}
	client.SetAppAccessToken(resp.Data.AccessToken)

	b.helixClient = client
	return nil
}

func (b *Bot) subscribeEvents(sessionID string) error {
	if b.config.UseMockServer {
		return nil
	}

	appToken := b.helixClient.GetAppAccessToken()

	if err := updateConduitShard(b.config.ClientID, appToken, b.conduitID, sessionID); err != nil {
		return err
	}

	type sub struct {
		subType   twitch.EventSubscription
		version   string
		condition map[string]string
	}

	subs := []sub{
		{
			subType: twitch.SubChannelChatMessage,
			version: "1",
			condition: map[string]string{
				"broadcaster_user_id": b.config.ChannelUserID,
				"user_id":             b.config.BotUserID,
			},
		},
		{
			subType: twitch.SubChannelChannelPointsCustomRewardRedemptionAdd,
			version: "1",
			condition: map[string]string{
				"broadcaster_user_id": b.config.ChannelUserID,
			},
		},
		{
			subType: twitch.SubChannelRaid,
			version: "1",
			condition: map[string]string{
				"to_broadcaster_user_id": b.config.ChannelUserID,
			},
		},
		{
			subType: twitch.SubConduitShardDisabled,
			version: "1",
			condition: map[string]string{
				"client_id": b.config.ClientID,
			},
		},
	}

	for _, s := range subs {
		slog.Info("subscribing to event via conduit", "type", s.subType)
		if err := createConduitSubscription(
			b.config.ClientID,
			appToken,
			b.conduitID,
			string(s.subType),
			s.version,
			s.condition,
		); err != nil {
			return fmt.Errorf("subscribe to %s: %w", s.subType, err)
		}
	}

	return nil
}

func (b *Bot) SendMessage(params SendMessageParams) {
	if b.helixClient == nil {
		return
	}

	msgParams := &helix.SendChatMessageParams{
		SenderID:      b.config.BotUserID,
		BroadcasterID: b.config.ChannelUserID,
		Message:       params.Message,
	}

	if params.ReplyParentMessageID != "" {
		msgParams.ReplyParentMessageID = params.ReplyParentMessageID
	}

	resp, err := b.helixClient.SendChatMessage(msgParams)
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		slog.Error("failed to send message", "err", err, "status_code", statusCode, "message", params.Message)
		return
	}

	if resp.Error != "" {
		slog.Warn("message send warning", "error", resp.Error, "message", params.Message)
	}

	slog.Debug("message sent", "message", params.Message)
}

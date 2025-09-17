package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
)

type WebSocketClient struct {
	twitchClient     *TwitchClient
	sessionID        string
	keepaliveTimeout time.Duration
	lastKeepalive    time.Time
	isLocalTesting   bool
}

func NewWebSocketClient(tc *TwitchClient) *WebSocketClient {
	return &WebSocketClient{
		twitchClient:   tc,
		isLocalTesting: false,
	}
}

func (wsc *WebSocketClient) EnableLocalTesting() {
	wsc.isLocalTesting = true
	slog.Info("🧪 Enabled local testing mode - will connect to CLI WebSocket server")
}

func (wsc *WebSocketClient) Start(ctx context.Context) error {
	for {
		slog.Info("🔌 Connecting to WebSocket...")

		if err := wsc.connect(ctx); err != nil {
			slog.Error("❌ WebSocket error", "error", err)
			slog.Info("🔄 Reconnecting in 5 seconds...")

			select {
			case <-time.After(5 * time.Second):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	}
}

func (wsc *WebSocketClient) connect(ctx context.Context) error {
	wsURL := "wss://eventsub.wss.twitch.tv/ws"
	if wsc.isLocalTesting {
		wsURL = "ws://localhost:8080/ws"
	}

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusInternalError, "closing")

	slog.Info("✅ WebSocket connected")

	wsCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go wsc.monitorKeepalive(wsCtx, cancel)

	for {
		select {
		case <-wsCtx.Done():
			return fmt.Errorf("keepalive timeout")
		default:
		}

		_, data, err := conn.Read(wsCtx)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Warn("Invalid WS message", "error", err)
			continue
		}

		wsc.handleMessage(&msg)
	}
}

func (wsc *WebSocketClient) monitorKeepalive(ctx context.Context, cancel context.CancelFunc) {
	if wsc.keepaliveTimeout == 0 {
		return
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(wsc.lastKeepalive) > wsc.keepaliveTimeout {
				slog.Warn("⚠️ Keepalive timeout exceeded")
				cancel()
				return
			}
		}
	}
}

func (wsc *WebSocketClient) handleMessage(msg *WebSocketMessage) {
	wsc.lastKeepalive = time.Now()

	switch msg.Metadata.MessageType {
	case "session_welcome":
		if msg.Payload.Session != nil {
			wsc.sessionID = msg.Payload.Session.ID
			if msg.Payload.Session.KeepaliveTimeoutSeconds != nil {
				wsc.keepaliveTimeout = time.Duration(*msg.Payload.Session.KeepaliveTimeoutSeconds) * time.Second
			}
			wsc.lastKeepalive = time.Now()

			slog.Info("📡 Received session_welcome", "session_id", wsc.sessionID, "local_testing", wsc.isLocalTesting)

			if !wsc.isLocalTesting {
				if err := wsc.twitchClient.registerEventSubListeners(wsc.sessionID); err != nil {
					slog.Error("Failed to register EventSub", "error", err)
					return
				}
			} else {
				slog.Info("🧪 Local testing mode - skipping EventSub registration")
			}
		}

	case "session_keepalive":
		slog.Debug("💓 Keepalive received")

	case "notification":
		switch msg.Metadata.SubscriptionType {
		case "channel.chat.message":
			wsc.handleChatMessage(msg)
		case "channel.channel_points_custom_reward_redemption.add":
			wsc.handleChannelPointsRedemption(msg)
		default:
			slog.Warn("Unknown subscription type", "type", msg.Metadata.SubscriptionType)
		}

	case "revocation":
		if msg.Payload.Subscription != nil {
			slog.Warn("❌ Subscription revoked",
				"subscription_id", msg.Payload.Subscription.ID,
				"status", msg.Payload.Subscription.Status)
		}

	default:
		slog.Warn("Unknown message type", "type", msg.Metadata.MessageType)
	}
}

func (wsc *WebSocketClient) handleChatMessage(msg *WebSocketMessage) {
	if msg.Payload.Event == nil {
		slog.Warn("Received chat message notification with no event data")
		return
	}

	event := msg.Payload.Event
	text := event.Message.Text
	username := event.ChatterUserLogin

	slog.Info("💬 Chat message received",
		"channel", event.BroadcasterUserLogin,
		"username", username,
		"user_id", event.ChatterUserID,
		"message_id", event.MessageID,
		"text", text)

	wsc.twitchClient.handleChatCommand(event)
}

func (wsc *WebSocketClient) handleChannelPointsRedemption(msg *WebSocketMessage) {
	if msg.Payload.Event == nil {
		slog.Warn("Received channel points redemption notification with no event data")
		return
	}

	event := msg.Payload.Event

	slog.Info("🎯 Channel points redemption received",
		"channel", event.BroadcasterUserLogin,
		"user", event.UserLogin,
		"user_id", event.UserID,
		"reward_title", func() string {
			if event.Reward != nil {
				return event.Reward.Title
			}
			return "unknown"
		}(),
		"user_input", event.UserInput,
		"status", event.Status)

	if err := wsc.twitchClient.handleChannelPointsRedemption(event); err != nil {
		slog.Error("Failed to handle channel points redemption", "error", err)
	}
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type TwitchClient struct {
	config       *Config
	accessToken  string
	refreshToken string
	httpClient   *http.Client
	handlers     map[string]func(username, message string) error
	db           *sql.DB
}

func NewTwitchClient(config *Config, db *sql.DB) *TwitchClient {
	tc := &TwitchClient{
		config:       config,
		accessToken:  config.OAuthToken,
		refreshToken: config.RefreshToken,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		handlers:     make(map[string]func(username, message string) error),
		db:           db,
	}

	tc.RegisterChatCommand("!stats", tc.handleStats)

	return tc
}

func (tc *TwitchClient) ValidateAuth() error {
	req, _ := http.NewRequest(http.MethodGet, "https://id.twitch.tv/oauth2/validate", nil)
	req.Header.Set("Authorization", "Bearer "+tc.accessToken)

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Invalid token", "status", resp.StatusCode, "body", string(body))
		if err := tc.refreshAccessToken(); err != nil {
			return err
		}
	}

	slog.Info("✅ Validated token")

	return nil
}

func (tc *TwitchClient) refreshAccessToken() error {
	slog.Info("🔄 Refreshing token...")

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tc.refreshToken)
	form.Set("client_id", tc.config.ClientID)
	form.Set("client_secret", tc.config.ClientSecret)

	res, err := http.Post(
		"https://id.twitch.tv/oauth2/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		data, _ := io.ReadAll(res.Body)
		return fmt.Errorf("token refresh failed (%d): %s", res.StatusCode, data)
	}

	var tr TokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return err
	}

	tc.accessToken = tr.AccessToken
	tc.refreshToken = tr.RefreshToken

	slog.Info("🔑 Refreshed token")

	return nil
}

func (tc *TwitchClient) SendChatMessage(message string) error {
	body := map[string]any{
		"broadcaster_id": tc.config.ChatChannelUserID,
		"sender_id":      tc.config.BotUserID,
		"message":        message,
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "https://api.twitch.tv/helix/chat/messages", bytes.NewReader(b))
	res, err := tc.doRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		data, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to send message (%d): %s", res.StatusCode, data)
	}

	slog.Info("✅ Sent chat message", "message", message)
	return nil
}

func (tc *TwitchClient) registerEventSubListeners(sessionID string) error {
	if err := tc.subscribeToEvent("channel.chat.message", "1", map[string]string{
		"broadcaster_user_id": tc.config.ChatChannelUserID,
		"user_id":             tc.config.BotUserID,
	}, sessionID); err != nil {
		return fmt.Errorf("failed to subscribe to chat messages: %w", err)
	}

	if err := tc.subscribeToEvent("channel.channel_points_custom_reward_redemption.add", "1", map[string]string{
		"broadcaster_user_id": tc.config.ChatChannelUserID,
	}, sessionID); err != nil {
		return fmt.Errorf("failed to subscribe to channel points: %w", err)
	}

	return nil
}

func (tc *TwitchClient) subscribeToEvent(eventType, version string, condition map[string]string, sessionID string) error {
	body := map[string]any{
		"type":      eventType,
		"version":   version,
		"condition": condition,
		"transport": map[string]string{
			"method":     "websocket",
			"session_id": sessionID,
		},
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewReader(b))
	res, err := tc.doRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 202 {
		data, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to subscribe to %s (%d): %s", eventType, res.StatusCode, data)
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Data) > 0 {
		slog.Info("✅ Subscribed to event", "event_type", eventType, "subscription_id", data.Data[0].ID)
	}
	return nil
}

func (tc *TwitchClient) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+tc.accessToken)
	req.Header.Set("Client-Id", tc.config.ClientID)
	req.Header.Set("Content-Type", "application/json")

	res, err := tc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == 401 {
		res.Body.Close()
		slog.Warn("Access token expired, refreshing...")
		if err := tc.refreshAccessToken(); err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+tc.accessToken)
		return tc.httpClient.Do(req)
	}

	return res, nil
}

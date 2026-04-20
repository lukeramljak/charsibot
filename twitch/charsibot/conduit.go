package charsibot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const twitchAPIBase = "https://api.twitch.tv/helix"

type conduitData struct {
	ID         string `json:"id"`
	ShardCount int    `json:"shard_count"`
}

type conduitListResponse struct {
	Data []conduitData `json:"data"`
}

type createConduitRequest struct {
	ShardCount int `json:"shard_count"`
}

type shardTransport struct {
	Method    string `json:"method"`
	SessionID string `json:"session_id"`
}

type shardData struct {
	ID        string         `json:"id"`
	Transport shardTransport `json:"transport"`
}

type updateShardsRequest struct {
	ConduitID string      `json:"conduit_id"`
	Shards    []shardData `json:"shards"`
}

type shardError struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type updateShardsResponse struct {
	Data   []shardData  `json:"data"`
	Errors []shardError `json:"errors"`
}

type conduitTransport struct {
	Method    string `json:"method"`
	ConduitID string `json:"conduit_id"`
}

type conduitSubscriptionRequest struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	Transport conduitTransport  `json:"transport"`
}

func twitchRequest(method, clientID, token, endpoint string, payload, result any) (int, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return 0, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, twitchAPIBase+endpoint, bodyReader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Client-Id", clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return resp.StatusCode, fmt.Errorf("%s %s failed (%d): %s", method, endpoint, resp.StatusCode, respBody)
	}

	return resp.StatusCode, nil
}

// getOrCreateConduit returns the ID of the first existing conduit, or creates
// a new single-shard conduit if none exist.
func getOrCreateConduit(clientID, appToken string) (string, error) {
	var list conduitListResponse
	if _, err := twitchRequest(http.MethodGet, clientID, appToken, "/eventsub/conduits", nil, &list); err != nil {
		return "", fmt.Errorf("list conduits: %w", err)
	}

	if len(list.Data) > 0 {
		return list.Data[0].ID, nil
	}

	var created conduitListResponse
	if _, err := twitchRequest(
		http.MethodPost,
		clientID,
		appToken,
		"/eventsub/conduits",
		createConduitRequest{ShardCount: 1},
		&created,
	); err != nil {
		return "", fmt.Errorf("create conduit: %w", err)
	}
	if len(created.Data) == 0 {
		return "", errors.New("conduit creation returned empty response")
	}

	return created.Data[0].ID, nil
}

// updateConduitShard points shard 0 of the conduit at the given WebSocket session.
func updateConduitShard(clientID, appToken, conduitID, sessionID string) error {
	payload := updateShardsRequest{
		ConduitID: conduitID,
		Shards: []shardData{
			{
				ID: "0",
				Transport: shardTransport{
					Method:    "websocket",
					SessionID: sessionID,
				},
			},
		},
	}

	var result updateShardsResponse
	if _, err := twitchRequest(
		http.MethodPatch,
		clientID,
		appToken,
		"/eventsub/conduits/shards",
		payload,
		&result,
	); err != nil {
		return fmt.Errorf("update conduit shard: %w", err)
	}

	if len(result.Errors) > 0 {
		return fmt.Errorf("conduit shard update error: %s (%s)", result.Errors[0].Message, result.Errors[0].Code)
	}

	return nil
}

// createConduitSubscription creates an EventSub subscription using conduit
// transport. A 409 Conflict (already exists) is treated as success.
func createConduitSubscription(
	clientID, appToken, conduitID, subType, version string,
	condition map[string]string,
) error {
	payload := conduitSubscriptionRequest{
		Type:      subType,
		Version:   version,
		Condition: condition,
		Transport: conduitTransport{
			Method:    "conduit",
			ConduitID: conduitID,
		},
	}

	status, err := twitchRequest(http.MethodPost, clientID, appToken, "/eventsub/subscriptions", payload, nil)
	if err != nil && status != http.StatusConflict {
		return err
	}

	return nil
}

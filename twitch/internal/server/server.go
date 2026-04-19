package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	helix "github.com/nicklaw5/helix/v2"

	"github.com/lukeramljak/charsibot/twitch/internal/bot"
	"github.com/lukeramljak/charsibot/twitch/internal/store"
)

type client struct {
	ch chan bot.OverlayEvent
}

type Server struct {
	port         int
	clientID     string
	clientSecret string
	redirectURI  string
	server       *http.Server
	clients      map[*client]bool
	mu           sync.RWMutex
	queries      *store.Queries
}

func NewServer(port int, clientID, clientSecret, redirectURI string, queries *store.Queries) *Server {
	return &Server{
		port:         port,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		clients:      make(map[*client]bool),
		queries:      queries,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.handleSSE)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			slog.Error("failed to write health response", "err", err)
		}
	})
	mux.HandleFunc("GET /oauth/start", s.handleOAuthStart)
	mux.HandleFunc("GET /oauth/callback", s.handleOAuthCallback)
	mux.HandleFunc("GET /api/blindbox", s.handleBlindBox)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: 0, // No timeout for SSE connections
	}

	slog.Info("SSE server started", "port", s.port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("SSE server error", "err", err)
		}
	}()

	return nil
}

const (
	serverReadTimeout  = 10 * time.Second
	eventChannelBuffer = 100
	shutdownTimeout    = 5 * time.Second
)

func oauthScopes(account string) ([]string, bool) {
	scopes := map[string][]string{
		"streamer": {"channel:read:redemptions", "channel:bot"},
		"bot":      {"user:read:chat", "user:write:chat", "user:bot"},
	}
	s, ok := scopes[account]
	return s, ok
}

func (s *Server) handleOAuthStart(w http.ResponseWriter, r *http.Request) {
	account := r.URL.Query().Get("account")
	scopes, ok := oauthScopes(account)
	if !ok {
		http.Error(w, `account must be "streamer" or "bot"`, http.StatusBadRequest)
		return
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:    s.clientID,
		RedirectURI: s.redirectURI,
	})
	if err != nil {
		slog.Error("failed to create helix client", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	authURL := helixClient.GetAuthorizationURL(&helix.AuthorizationURLParams{
		ResponseType: "code",
		Scopes:       scopes,
		State:        account,
		ForceVerify:  true,
	})

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Server) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	account := r.URL.Query().Get("state")
	if _, ok := oauthScopes(account); !ok {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error_description")
		if errMsg == "" {
			errMsg = r.URL.Query().Get("error")
		}
		http.Error(w, "auth denied: "+errMsg, http.StatusBadRequest)
		return
	}

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		RedirectURI:  s.redirectURI,
	})
	if err != nil {
		slog.Error("failed to create helix client", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tokenResp, err := helixClient.RequestUserAccessToken(code)
	if err != nil {
		slog.Error("token exchange request failed", "account", account, "err", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}
	if tokenResp.ErrorMessage != "" {
		slog.Error(
			"twitch returned error during token exchange",
			"account",
			account,
			"error",
			tokenResp.Error,
			"message",
			tokenResp.ErrorMessage,
		)
		http.Error(w, "token exchange failed: "+tokenResp.ErrorMessage, http.StatusInternalServerError)
		return
	}

	slog.Info("OAuth authorization complete", "account", account)
	accountLabel := strings.ToUpper(account[:1]) + account[1:]
	fmt.Fprintf(w, "%s authorization complete.", accountLabel)
}

func (s *Server) handleBlindBox(w http.ResponseWriter, r *http.Request) {
	series, err := bot.LoadAllSeries(r.Context(), s.queries)
	if err != nil {
		http.Error(w, "failed to load series", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(series); err != nil {
		slog.Error("failed to encode series response", "err", err)
	}
}

func (s *Server) addClient(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c] = true
}

func (s *Server) removeClient(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c)
	close(c.ch)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	client := &client{
		ch: make(chan bot.OverlayEvent, eventChannelBuffer),
	}
	s.addClient(client)
	defer s.removeClient(client)

	slog.Info("client connected")

	connectedEvent := map[string]any{
		"type":      "connected",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	connData, err := json.Marshal(connectedEvent)
	if err == nil {
		fmt.Fprintf(w, "data: %s\n\n", connData)
		flusher.Flush()
	}

	clientGone := r.Context().Done()

	for {
		select {
		case <-clientGone:
			slog.Debug("client disconnected")
			return
		case event := <-client.ch:
			eventData, err := json.Marshal(event)
			if err != nil {
				slog.Error("failed to marshal event", "err", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", eventData)
			flusher.Flush()
		}
	}
}

func (s *Server) Broadcast(event bot.OverlayEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.clients) == 0 {
		slog.Warn("no clients connected, event dropped", "type", event.Type)
		return
	}

	for client := range s.clients {
		select {
		case client.ch <- event:
			// Event sent successfully
		default:
			slog.Warn("client channel full, dropping event", "type", event.Type)
		}
	}

	slog.Info("event sent", "type", event.Type, "clients", len(s.clients))
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			slog.Error("error shutting down server", "err", err)
		}
	}
}

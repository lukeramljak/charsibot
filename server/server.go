package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	helix "github.com/nicklaw5/helix/v2"
)

//go:embed all:web
var webFS embed.FS

const (
	serverReadTimeout  = 10 * time.Second
	shutdownTimeout    = 5 * time.Second
	eventChannelBuffer = 10
	pingInterval       = 30 * time.Second
)

type ServerConfig struct {
	Port             int
	ClientID         string
	ClientSecret     string
	OAuthRedirectURI string
}

// Server handles SSE streaming and OAuth.
type Server struct {
	cfg     ServerConfig
	logger  *slog.Logger
	server  *http.Server
	clients map[chan OverlayEvent]struct{}
	mu      sync.RWMutex
}

func NewServer(cfg ServerConfig, logger *slog.Logger) *Server {
	return &Server{
		cfg:     cfg,
		logger:  logger,
		clients: make(map[chan OverlayEvent]struct{}),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", s.handleSSE)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /oauth/start", s.handleOAuthStart)
	mux.HandleFunc("GET /oauth/callback", s.handleOAuthCallback)

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		return fmt.Errorf("create web sub-filesystem: %w", err)
	}
	fileServer := http.FileServerFS(webContent)
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve real files directly; fall back to index.html for SPA routes.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(webContent, path); err != nil {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	}))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Port),
		Handler:      mux,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: 0, // No timeout for SSE connections
	}

	s.logger.Info("server started", "port", s.cfg.Port)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server error", "err", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("error shutting down server", "err", err)
		}
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan OverlayEvent, eventChannelBuffer)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		close(ch)
		s.mu.Unlock()
	}()

	s.logger.Info("SSE client connected", "remote_addr", r.RemoteAddr)

	fmt.Fprintf(w, ": ping\n\n")
	flusher.Flush()

	ping := time.NewTicker(pingInterval)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			s.logger.Debug("SSE client disconnected")
			return
		case <-ping.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case event := <-ch:
			data, err := json.Marshal(event.Data)
			if err != nil {
				s.logger.Error("failed to marshal event", "err", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}

func (s *Server) Broadcast(event OverlayEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.clients {
		select {
		case ch <- event:
		default:
			s.logger.Warn("SSE client buffer full, dropping event", "type", event.Type)
		}
	}
}

func oauthScopes(account string) ([]string, bool) {
	scopes := map[string][]string{
		"streamer": {"channel:manage:redemptions", "channel:read:redemptions", "channel:bot"},
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
		ClientID:    s.cfg.ClientID,
		RedirectURI: s.cfg.OAuthRedirectURI,
	})
	if err != nil {
		s.logger.Error("failed to create helix client", "err", err)
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
		ClientID:     s.cfg.ClientID,
		ClientSecret: s.cfg.ClientSecret,
		RedirectURI:  s.cfg.OAuthRedirectURI,
	})
	if err != nil {
		s.logger.Error("failed to create helix client", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tokenResp, err := helixClient.RequestUserAccessToken(code)
	if err != nil {
		s.logger.Error("token exchange request failed", "account", account, "err", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}
	if tokenResp.ErrorMessage != "" {
		s.logger.Error(
			"twitch returned error during token exchange",
			"account", account,
			"error", tokenResp.Error,
			"message", tokenResp.ErrorMessage,
		)
		http.Error(w, "token exchange failed: "+tokenResp.ErrorMessage, http.StatusInternalServerError)
		return
	}

	s.logger.Info("OAuth authorization complete", "account", account)
	accountLabel := strings.ToUpper(account[:1]) + account[1:]
	fmt.Fprintf(w, "%s authorization complete.", accountLabel)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		s.logger.Error("failed to write health response", "err", err)
	}
}

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	port     int
	logger   *slog.Logger
	server   *http.Server
	clientCh chan OverlayEvent
}

func NewServer(port int, logger *slog.Logger) *Server {
	return &Server{
		port:     port,
		logger:   logger,
		clientCh: make(chan OverlayEvent, 10),
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.handleSSE)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // No timeout for SSE connections
	}

	s.logger.Info("SSE server started", "port", s.port)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("SSE server error", "err", err)
	}

	return nil
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

	s.logger.Info("client connected")

	connectedEvent := map[string]any{
		"type":      "connected",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(connectedEvent)
	if err == nil {
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	clientGone := r.Context().Done()

	for {
		select {
		case <-clientGone:
			s.logger.Debug("client disconnected")
			return
		case event := <-s.clientCh:
			data, err := json.Marshal(event)
			if err != nil {
				s.logger.Error("failed to marshal event", "err", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (s *Server) Broadcast(event OverlayEvent) {
	select {
	case s.clientCh <- event:
		s.logger.Info("event sent", "type", event.Type)
	default:
		s.logger.Warn("no client connected, event dropped", "type", event.Type)
	}
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}
}

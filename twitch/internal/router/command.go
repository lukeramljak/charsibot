package router

import (
	"context"
	"log/slog"
	"strings"

	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

type CommandHandler func(context.Context, *eventsub.ChannelChatMessage) (string, error)

type CommandRouter struct {
	handlers map[string]CommandHandler
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{handlers: map[string]CommandHandler{}}
}

func (r *CommandRouter) Register(name string, h CommandHandler) {
	r.handlers[strings.ToLower(name)] = h
}

func (r *CommandRouter) Handle(ctx context.Context, ev *eventsub.ChannelChatMessage) (string, bool) {
	if ev == nil || ev.Message.Text == "" {
		return "", false
	}

	text := strings.TrimSpace(ev.Message.Text)
	if len(text) == 0 || text[0] != '!' {
		return "", false
	}

	cmd, _, _ := strings.Cut(strings.ToLower(text), " ")
	h, ok := r.handlers[cmd]
	if !ok {
		return "", false
	}

	res, err := h(ctx, ev)
	if err != nil {
		slog.Error("command handler error", "command", cmd, "error", err)
		return "", false
	}

	return res, true
}

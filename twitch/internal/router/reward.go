package router

import (
	"context"
	"log/slog"

	"github.com/vpetrigo/go-twitch-ws/pkg/eventsub"
)

type RewardHandler func(context.Context, *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) error

type RewardRouter struct {
	handlers map[string]RewardHandler
}

func NewRewardRouter() *RewardRouter {
	return &RewardRouter{handlers: map[string]RewardHandler{}}
}

func (r *RewardRouter) Register(title string, h RewardHandler) {
	r.handlers[title] = h
}

func (r *RewardRouter) Handle(ctx context.Context, ev *eventsub.ChannelPointsCustomRewardRedemptionAddEvent) {
	if ev == nil {
		return
	}
	if h, ok := r.handlers[ev.Reward.Title]; ok {
		if err := h(ctx, ev); err != nil {
			slog.Error("reward handler error", "reward", ev.Reward.Title, "error", err, "user", ev.UserLogin, "user_id", ev.UserID, "redemption_id", ev.ID)
		} else {
			slog.Info("reward handled", "reward", ev.Reward.Title, "user", ev.UserLogin, "user_id", ev.UserID, "redemption_id", ev.ID, "cost", ev.Reward.Cost)
		}
	}
}

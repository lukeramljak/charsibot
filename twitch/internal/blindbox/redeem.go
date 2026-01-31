package blindbox

import (
	"fmt"
	"log/slog"

	"github.com/lukeramljak/charsibot/internal/bot"
	"github.com/lukeramljak/charsibot/internal/server"
)

// RedeemBlindBox handles the common logic for blind box redemptions
// Used by both channel point redemptions and moderator commands
func RedeemBlindBox(b *bot.Bot, userId, username string, config BlindBoxConfig) {
	plushieKey := GetWeightedRandomPlushie(config.Plushies)

	isNew, collection, err := b.Store().AddPlushieToCollection(
		b.Context(),
		userId,
		username,
		config.CollectionType,
		plushieKey,
	)
	if err != nil {
		slog.Error("failed to add plushie to collection", "err", err, "user", username)
		return
	}

	b.BroadcastOverlayEvent(server.OverlayEvent{
		Type: "blindbox_redemption",
		Data: map[string]any{
			"userId":         userId,
			"username":       username,
			"collectionType": config.CollectionType,
			"seriesName":     config.RewardTitle,
			"plushie":        fmt.Sprintf("reward%d", plushieKey),
			"isNew":          isNew,
			"collectionSize": len(collection),
			"collection":     intSliceToRewardKeys(collection),
		},
	})

	slog.Info("blind box redeemed",
		"user", username,
		"collection", config.CollectionType,
		"plushie", plushieKey,
		"isNew", isNew,
	)
}

func intSliceToRewardKeys(ints []int) []string {
	keys := make([]string, len(ints))
	for i, num := range ints {
		keys[i] = fmt.Sprintf("reward%d", num)
	}
	return keys
}

package bot

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"

	"github.com/lukeramljak/charsibot/internal/store"
)

// SeriesConfig holds the runtime config for a blind box series.
// BlindBoxSeries is embedded so its fields are promoted to the top level.
type SeriesConfig struct {
	store.BlindBoxSeries
	Plushies []store.BlindBoxPlushie `json:"plushies"`
}

// LoadAllSeries queries all series and their plushies from the DB and returns
// one SeriesConfig per series. The caller decides what to register.
func LoadAllSeries(ctx context.Context, q *store.Queries) ([]SeriesConfig, error) {
	rows, err := q.GetAllSeriesWithPlushies(ctx)
	if err != nil {
		return nil, fmt.Errorf("load blind box series: %w", err)
	}
	return groupSeriesRows(rows), nil
}

func groupSeriesRows(rows []store.GetAllSeriesWithPlushiesRow) []SeriesConfig {
	var configs []SeriesConfig
	index := map[string]int{}

	for _, row := range rows {
		if _, seen := index[row.Series]; !seen {
			index[row.Series] = len(configs)
			configs = append(configs, SeriesConfig{
				BlindBoxSeries: store.BlindBoxSeries{
					Series:          row.Series,
					RedemptionTitle: row.RedemptionTitle,
					Name:            row.Name,
					RevealSound:     row.RevealSound,
					BoxFrontFace:    row.BoxFrontFace,
					BoxSideFace:     row.BoxSideFace,
					DisplayColor:    row.DisplayColor,
					TextColor:       row.TextColor,
				},
			})
		}
		if row.PlushieKey.Valid {
			i := index[row.Series]
			configs[i].Plushies = append(configs[i].Plushies, store.BlindBoxPlushie{
				ID:         row.PlushieID.Int64,
				Series:     row.Series,
				Key:        row.PlushieKey.String,
				SortOrder:  row.PlushieSortOrder.Int64,
				Weight:     row.PlushieWeight.Int64,
				Name:       row.PlushieName.String,
				Image:      row.PlushieImage.String,
				EmptyImage: row.PlushieEmptyImage.String,
			})
		}
	}

	return configs
}

func weightedRandomKey(plushies []store.BlindBoxPlushie) string {
	weighted := []string{}
	for _, p := range plushies {
		for i := 0; i < int(p.Weight); i++ {
			weighted = append(weighted, p.Key)
		}
	}
	if len(weighted) == 0 {
		return "secret"
	}
	return weighted[rand.IntN(len(weighted))]
}

func addPlushieToCollection(ctx context.Context, q *store.Queries, userID, username, series, key string) (bool, []string, error) {
	// INSERT OR IGNORE: inserts only if the row doesn't exist.
	// changes() returns 1 for a new insert, 0 if the row already existed.
	if err := q.InsertUserPlushieIfNew(ctx, store.InsertUserPlushieIfNewParams{
		UserID:   userID,
		Username: username,
		Series:   series,
		Key:      key,
	}); err != nil {
		return false, nil, fmt.Errorf("insert user plushie: %w", err)
	}

	n, err := q.LastChangeCount(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("get change count: %w", err)
	}
	isNew := n == 1

	// Always sync the username in case it changed (e.g. display name update).
	if err := q.UpsertUserPlushie(ctx, store.UpsertUserPlushieParams{
		UserID:   userID,
		Username: username,
		Series:   series,
		Key:      key,
	}); err != nil {
		return false, nil, fmt.Errorf("sync username: %w", err)
	}

	keys, err := q.GetCollectedPlushies(ctx, store.GetCollectedPlushiesParams{
		UserID: userID,
		Series: series,
	})
	if err != nil {
		return false, nil, fmt.Errorf("get collected plushies: %w", err)
	}

	return isNew, keys, nil
}

// RedeemBlindBox handles blind box redemptions.
// Used by both channel point redemptions and moderator commands.
func RedeemBlindBox(b *Bot, userID, username string, cfg SeriesConfig) {
	key := weightedRandomKey(cfg.Plushies)

	isNew, collection, err := addPlushieToCollection(
		b.ctx,
		b.store,
		userID,
		username,
		cfg.Series,
		key,
	)
	if err != nil {
		slog.Error("failed to add plushie to collection", "err", err, "user", username)
		return
	}

	b.BroadcastOverlayEvent(OverlayEvent{
		Type: EventTypeBlindBoxRedemption,
		Data: map[string]any{
			"userId":         userID,
			"username":       username,
			"series":         cfg.Series,
			"seriesName":     cfg.RedemptionTitle,
			"plushie":        key,
			"isNew":          isNew,
			"collectionSize": len(collection),
			"collection":     collection,
		},
	})

	slog.Info("blind box redeemed",
		"user", username,
		"series", cfg.Series,
		"plushie", key,
		"isNew", isNew,
	)
}

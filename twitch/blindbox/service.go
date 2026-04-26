package blindbox

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/lukeramljak/charsibot/twitch/db"
)

// SeriesConfig holds the runtime config for a blind box series.
// BlindBoxSeries is embedded so its fields are promoted to the top level.
type SeriesConfig struct {
	db.BlindBoxSeries

	Plushies []db.BlindBoxPlushie `json:"plushies"`
}

// RedemptionResult holds the outcome of a blind box redemption.
type RedemptionResult struct {
	UserID     string
	Username   string
	Series     string
	Plushie    string
	IsNew      bool
	Collection []string
}

type Service struct {
	queries *db.Queries
}

// NewService creates a new blind box Service backed by the given queries.
func NewService(queries *db.Queries) (*Service, error) {
	if queries == nil {
		return nil, errors.New("queries must not be nil")
	}

	return &Service{queries}, nil
}

// LoadAllSeries queries all series and their plushies from the DB and returns
// one SeriesConfig per series. The caller decides what to register.
func (s *Service) LoadAllSeries(ctx context.Context) ([]SeriesConfig, error) {
	rows, err := s.queries.GetAllSeriesWithPlushies(ctx)
	if err != nil {
		return nil, fmt.Errorf("load blind box series: %w", err)
	}
	return groupSeriesRows(rows), nil
}

// AddPlushieToCollection inserts a plushie into the user's collection if not
// already present, syncs the username, and returns whether the plushie was new
// and the user's full collection for the series.
func (s *Service) AddPlushieToCollection(
	ctx context.Context,
	userID,
	username,
	series,
	key string,
) (bool, []string, error) {
	// INSERT OR IGNORE: inserts only if the row doesn't exist.
	// changes() returns 1 for a new insert, 0 if the row already existed.
	if err := s.queries.InsertUserPlushieIfNew(ctx, db.InsertUserPlushieIfNewParams{
		UserID:   userID,
		Username: username,
		Series:   series,
		Key:      key,
	}); err != nil {
		return false, nil, fmt.Errorf("insert user plushie: %w", err)
	}

	n, err := s.queries.LastChangeCount(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("get change count: %w", err)
	}
	isNew := n == 1

	// Always sync the username in case it changed (e.g. display name update).
	if err = s.queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
		UserID:   userID,
		Username: username,
		Series:   series,
		Key:      key,
	}); err != nil {
		return false, nil, fmt.Errorf("sync username: %w", err)
	}

	keys, err := s.queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: userID,
		Series: series,
	})
	if err != nil {
		return false, nil, fmt.Errorf("get collected plushies: %w", err)
	}

	return isNew, keys, nil
}

// Redeem records a blind box redemption for a user and returns the result.
// The caller is responsible for selecting the plushie key (e.g. via PickPlushie)
// and for broadcasting the resulting event.
func (s *Service) Redeem(ctx context.Context, userID, username, series, key string) (*RedemptionResult, error) {
	isNew, collection, err := s.AddPlushieToCollection(ctx, userID, username, series, key)
	if err != nil {
		return nil, fmt.Errorf("add plushie to collection: %w", err)
	}

	return &RedemptionResult{
		UserID:     userID,
		Username:   username,
		Series:     series,
		Plushie:    key,
		IsNew:      isNew,
		Collection: collection,
	}, nil
}

// GetCompletedCollections returns all users who have completed a collection.
func (s *Service) GetCompletedCollections(ctx context.Context) ([]db.GetCompletedCollectionsRow, error) {
	return s.queries.GetCompletedCollections(ctx)
}

// GetCollection returns the plushie keys collected by a user for a series.
func (s *Service) GetCollection(ctx context.Context, userID, series string) ([]string, error) {
	return s.queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: userID,
		Series: series,
	})
}

// ResetCollection removes all plushies for a user in a series.
func (s *Service) ResetCollection(ctx context.Context, userID, series string) error {
	return s.queries.ResetUserPlushies(ctx, db.ResetUserPlushiesParams{
		UserID: userID,
		Series: series,
	})
}

// PickPlushie selects a random plushie key from the given plushies using
// weighted random selection. Returns "secret" if the list is empty.
func PickPlushie(plushies []db.BlindBoxPlushie) string {
	weighted := []string{}
	for _, p := range plushies {
		for range p.Weight {
			weighted = append(weighted, p.Key)
		}
	}
	if len(weighted) == 0 {
		return "secret"
	}
	return weighted[rand.IntN(len(weighted))]
}

func groupSeriesRows(rows []db.GetAllSeriesWithPlushiesRow) []SeriesConfig {
	var configs []SeriesConfig
	index := map[string]int{}

	for _, row := range rows {
		if _, seen := index[row.Series]; !seen {
			index[row.Series] = len(configs)
			configs = append(configs, SeriesConfig{
				BlindBoxSeries: db.BlindBoxSeries{
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
			configs[i].Plushies = append(configs[i].Plushies, db.BlindBoxPlushie{
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

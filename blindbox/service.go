package blindbox

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/lukeramljak/charsibot/db"
)

// SeriesConfig holds the runtime config for a blind box series.
type SeriesConfig struct {
	Series          string    `json:"series"`
	RedemptionTitle string    `json:"redemptionTitle"`
	Name            string    `json:"name"`
	RevealSound     string    `json:"revealSound"`
	BoxFrontFace    string    `json:"boxFrontFace"`
	BoxSideFace     string    `json:"boxSideFace"`
	DisplayColor    string    `json:"displayColor"`
	TextColor       string    `json:"textColor"`
	Plushies        []Plushie `json:"plushies"`
}

// Plushie is a catalog entry that can be awarded by a blind box.
type Plushie struct {
	Series     string `json:"series"`
	Key        string `json:"key"`
	SortOrder  int64  `json:"sortOrder"`
	Weight     int64  `json:"weight"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	EmptyImage string `json:"emptyImage"`
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

type CompletedCollection struct {
	SeriesName string
	Usernames  string
}

type Service struct {
	queries *db.Queries
	series  []SeriesConfig
}

// NewService creates a new blind box Service backed by the given queries and JSON catalog series.
func NewService(queries *db.Queries, series []SeriesConfig) (*Service, error) {
	if queries == nil {
		return nil, errors.New("queries must not be nil")
	}
	if len(series) == 0 {
		return nil, errors.New("blind-box series must not be empty")
	}
	return &Service{queries: queries, series: append([]SeriesConfig(nil), series...)}, nil
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
	if err := s.queries.InsertUserPlushieIfNew(ctx, db.InsertUserPlushieIfNewParams{
		UserID:   userID,
		Username: username,
		Series:   series,
		Key:      key,
	}); err != nil {
		return false, nil, fmt.Errorf("insert plushie: %w", err)
	}
	changed, err := s.queries.LastChangeCount(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("read change count: %w", err)
	}
	isNew := changed > 0
	if !isNew {
		err = s.queries.UpsertUserPlushie(ctx, db.UpsertUserPlushieParams{
			UserID:   userID,
			Username: username,
			Series:   series,
			Key:      key,
		})
		if err != nil {
			return false, nil, fmt.Errorf("sync username: %w", err)
		}
	}
	collection, err := s.queries.GetCollectedPlushies(ctx, db.GetCollectedPlushiesParams{
		UserID: userID,
		Series: series,
	})
	if err != nil {
		return false, nil, fmt.Errorf("get collection: %w", err)
	}
	return isNew, collection, nil
}

// Redeem records a blind box redemption for a user and returns the result.
// The caller is responsible for selecting the plushie key (e.g. via PickPlushie)
// and for broadcasting the resulting event.
func (s *Service) Redeem(ctx context.Context, userID, username, series, key string) (*RedemptionResult, error) {
	isNew, collection, err := s.AddPlushieToCollection(ctx, userID, username, series, key)
	if err != nil {
		return nil, err
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
func (s *Service) GetCompletedCollections(ctx context.Context) ([]CompletedCollection, error) {
	counts, err := s.queries.GetUserPlushieCounts(ctx)
	if err != nil {
		return nil, err
	}
	seriesByKey := make(map[string]SeriesConfig, len(s.series))
	for _, cfg := range s.series {
		seriesByKey[cfg.Series] = cfg
	}
	completed := make(map[string][]string)
	for _, count := range counts {
		cfg, ok := seriesByKey[count.Series]
		if !ok || int(count.Count) != len(cfg.Plushies) {
			continue
		}
		completed[cfg.Name] = append(completed[cfg.Name], count.Username)
	}
	names := make([]string, 0, len(completed))
	for name := range completed {
		names = append(names, name)
	}
	sort.Strings(names)
	rows := make([]CompletedCollection, 0, len(names))
	for _, name := range names {
		usernames := completed[name]
		sort.Strings(usernames)
		rows = append(rows, CompletedCollection{
			SeriesName: name,
			Usernames:  strings.Join(usernames, ", "),
		})
	}
	return rows, nil
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

// PickPlushie selects a random plushie using weighted random selection.
// Returns an error if no plushies have a positive weight.
func PickPlushie(plushies []Plushie) (Plushie, error) {
	weighted := []Plushie{}
	for _, p := range plushies {
		for range p.Weight {
			weighted = append(weighted, p)
		}
	}
	if len(weighted) == 0 {
		return Plushie{}, errors.New("no plushies with positive weight")
	}
	return weighted[rand.IntN(len(weighted))], nil
}

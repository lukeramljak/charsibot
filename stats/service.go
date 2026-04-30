package stats

import (
	"context"
	"errors"
	"fmt"

	"github.com/lukeramljak/charsibot/db"
)

type Service struct {
	queries *db.Queries
}

// NewService creates a new stats Service backed by the given queries.
func NewService(queries *db.Queries) (*Service, error) {
	if queries == nil {
		return nil, errors.New("queries must not be nil")
	}

	return &Service{queries}, nil
}

// GetOrCreateStats ensures stat rows exist for a user then returns their stats.
func (s *Service) GetOrCreateStats(ctx context.Context, userID, username string) ([]db.GetUserStatsRow, error) {
	if err := s.queries.EnsureUserStats(ctx, db.EnsureUserStatsParams{
		UserID:   userID,
		Username: username,
		UserID_2: userID,
	}); err != nil {
		return nil, fmt.Errorf("ensure stats: %w", err)
	}
	if err := s.queries.UpdateUsername(ctx, db.UpdateUsernameParams{
		Username: username,
		UserID:   userID,
	}); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}
	return s.queries.GetUserStats(ctx, userID)
}

func (s *Service) GetUserStats(ctx context.Context, userID string) ([]db.GetUserStatsRow, error) {
	return s.queries.GetUserStats(ctx, userID)
}

func (s *Service) GetStatLeaderboard(ctx context.Context) ([]db.GetStatLeaderboardRow, error) {
	return s.queries.GetStatLeaderboard(ctx)
}

func (s *Service) GetRandomStatDefinition(ctx context.Context) (db.StatDefinition, error) {
	return s.queries.GetRandomStatDefinition(ctx)
}

func (s *Service) ModifyStatValue(ctx context.Context, userID, statName string, value int64) error {
	return s.queries.ModifyStatValue(ctx, db.ModifyStatValueParams{
		UserID:   userID,
		StatName: statName,
		Value:    value,
	})
}

func (s *Service) SetStatValue(ctx context.Context, userID, statName string, value int64) error {
	return s.queries.SetStatValue(ctx, db.SetStatValueParams{
		UserID:   userID,
		StatName: statName,
		Value:    value,
	})
}

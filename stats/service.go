package stats

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"

	"github.com/lukeramljak/charsibot/db"
)

type UserStat struct {
	Name      string
	ShortName string
	LongName  string
	Value     int64
}

// Definition describes a stat sourced from the runtime catalog.
type Definition struct {
	Name         string
	ShortName    string
	LongName     string
	DefaultValue int64
	SortOrder    int64
	Emoji        string
}

type LeaderboardRow struct {
	Emoji    string
	Username string
	Value    int64
}

// User is a viewer known to the bot through stats or blind-box collection data.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Service struct {
	queries     *db.Queries
	definitions []Definition
}

// NewService creates a new stats Service backed by the given queries and JSON catalog definitions.
func NewService(queries *db.Queries, definitions []Definition) (*Service, error) {
	if queries == nil {
		return nil, errors.New("queries must not be nil")
	}
	if len(definitions) == 0 {
		return nil, errors.New("stat definitions must not be empty")
	}
	defs := append([]Definition(nil), definitions...)
	sort.Slice(defs, func(i, j int) bool { return defs[i].SortOrder < defs[j].SortOrder })
	return &Service{queries: queries, definitions: defs}, nil
}

// GetOrCreateStats ensures stat rows exist for a user then returns their stats.
func (s *Service) GetOrCreateStats(ctx context.Context, userID, username string) ([]UserStat, error) {
	for _, definition := range s.definitions {
		if err := s.queries.EnsureUserStat(ctx, db.EnsureUserStatParams{
			UserID:   userID,
			Username: username,
			StatName: definition.Name,
			Value:    definition.DefaultValue,
		}); err != nil {
			return nil, fmt.Errorf("ensure stat %s: %w", definition.Name, err)
		}
	}
	if err := s.queries.UpdateUsername(ctx, db.UpdateUsernameParams{
		Username: username,
		UserID:   userID,
	}); err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}
	return s.GetUserStats(ctx, userID)
}

func (s *Service) GetUserStats(ctx context.Context, userID string) ([]UserStat, error) {
	values, err := s.queries.GetUserStatValues(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	byName := make(map[string]int64, len(values))
	for _, value := range values {
		byName[value.StatName] = value.Value
	}
	stats := make([]UserStat, 0, len(s.definitions))
	for _, definition := range s.definitions {
		value, ok := byName[definition.Name]
		if !ok {
			continue
		}
		stats = append(stats, UserStat{
			Name:      definition.Name,
			ShortName: definition.ShortName,
			LongName:  definition.LongName,
			Value:     value,
		})
	}
	return stats, nil
}

func (s *Service) GetStatLeaderboard(ctx context.Context) ([]LeaderboardRow, error) {
	values, err := s.queries.GetAllUserStatValues(ctx)
	if err != nil {
		return nil, err
	}
	type bestStat struct {
		username string
		value    int64
		seen     bool
	}
	bestByName := make(map[string]bestStat, len(s.definitions))
	for _, value := range values {
		best := bestByName[value.StatName]
		if !best.seen || value.Value > best.value {
			bestByName[value.StatName] = bestStat{username: value.Username, value: value.Value, seen: true}
		}
	}
	rows := make([]LeaderboardRow, 0, len(s.definitions))
	for _, definition := range s.definitions {
		best := bestByName[definition.Name]
		if !best.seen {
			continue
		}
		rows = append(rows, LeaderboardRow{
			Emoji:    definition.Emoji,
			Username: best.username,
			Value:    best.value,
		})
	}
	return rows, nil
}

// ListUsers returns all viewers known through stats or blind-box collection data.
func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, len(rows))
	for i, row := range rows {
		users[i] = User{ID: row.UserID, Username: row.Username}
	}
	return users, nil
}

// GetUser returns a known viewer by Twitch user ID.
func (s *Service) GetUser(ctx context.Context, userID string) (User, error) {
	row, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, err
	}
	return User{ID: row.UserID, Username: row.Username}, nil
}

// Definitions returns the configured stat definitions in display order.
func (s *Service) Definitions() []Definition {
	return append([]Definition(nil), s.definitions...)
}

func (s *Service) GetRandomStatDefinition(context.Context) (Definition, error) {
	return s.definitions[rand.IntN(len(s.definitions))], nil
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

// ResetStats restores every configured stat for a user to its catalog default.
func (s *Service) ResetStats(ctx context.Context, userID string) error {
	for _, definition := range s.definitions {
		if err := s.SetStatValue(ctx, userID, definition.Name, definition.DefaultValue); err != nil {
			return fmt.Errorf("reset stat %s: %w", definition.Name, err)
		}
	}
	return nil
}

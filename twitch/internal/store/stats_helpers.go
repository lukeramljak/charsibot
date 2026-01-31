package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// ModifyStat modifies a specific stat for a user
func (q *Queries) ModifyStat(ctx context.Context, id, username, statColumn string, amount int64) error {
	// Validate stat column to prevent SQL injection
	validColumns := map[string]bool{
		"strength":     true,
		"intelligence": true,
		"charisma":     true,
		"luck":         true,
		"dexterity":    true,
		"penis":        true,
	}

	normalizedColumn := strings.ToLower(statColumn)
	if !validColumns[normalizedColumn] {
		return fmt.Errorf("invalid stat column: %s", statColumn)
	}

	// First ensure user exists
	if err := q.UpsertStatsUser(ctx, UpsertStatsUserParams{
		ID:       id,
		Username: username,
	}); err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	// Then update the specific stat
	query := fmt.Sprintf(`UPDATE stats SET %s = %s + ? WHERE id = ?`, normalizedColumn, normalizedColumn)

	_, err := q.db.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to modify stat %s: %w", normalizedColumn, err)
	}

	return nil
}

// FormatStats formats a Stat struct into a human-readable string
func FormatStats(username string, s Stat) string {
	return fmt.Sprintf("%s's stats: STR: %d | INT: %d | CHA: %d | LUCK: %d | DEX: %d | PENIS: %d",
		username, s.Strength, s.Intelligence, s.Charisma, s.Luck, s.Dexterity, s.Penis)
}

// GetUserCollection converts a UserCollection row into a collection array
func GetUserCollection(uc UserCollection) []int {
	collection := []int{}
	rewards := []sql.NullInt64{uc.Reward1, uc.Reward2, uc.Reward3, uc.Reward4, uc.Reward5, uc.Reward6, uc.Reward7, uc.Reward8}

	for i, reward := range rewards {
		if reward.Valid && reward.Int64 == 1 {
			collection = append(collection, i+1)
		}
	}

	return collection
}

// AddPlushieToCollection adds a plushie to a user's collection and returns whether it was new
func (q *Queries) AddPlushieToCollection(ctx context.Context, userID, username, collectionType string, rewardNum int) (bool, []int, error) {

	existing, err := q.GetUserCollectionRow(ctx, GetUserCollectionRowParams{
		UserID:         sql.NullString{String: userID, Valid: true},
		CollectionType: sql.NullString{String: collectionType, Valid: true},
	})

	var hadBefore bool
	if err == nil {
		// Check if they already had this reward
		rewards := []sql.NullInt64{existing.Reward1, existing.Reward2, existing.Reward3, existing.Reward4, existing.Reward5, existing.Reward6, existing.Reward7, existing.Reward8}
		if rewardNum > 0 && rewardNum <= len(rewards) {
			hadBefore = rewards[rewardNum-1].Valid && rewards[rewardNum-1].Int64 == 1
		}
	}

	// Add the reward
	switch rewardNum {
	case 1:
		err = q.AddReward1(ctx, AddReward1Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 2:
		err = q.AddReward2(ctx, AddReward2Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 3:
		err = q.AddReward3(ctx, AddReward3Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 4:
		err = q.AddReward4(ctx, AddReward4Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 5:
		err = q.AddReward5(ctx, AddReward5Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 6:
		err = q.AddReward6(ctx, AddReward6Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 7:
		err = q.AddReward7(ctx, AddReward7Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	case 8:
		err = q.AddReward8(ctx, AddReward8Params{
			UserID:         sql.NullString{String: userID, Valid: true},
			Username:       username,
			CollectionType: sql.NullString{String: collectionType, Valid: true},
		})
	default:
		return false, nil, fmt.Errorf("invalid reward number: %d", rewardNum)
	}

	if err != nil {
		return false, nil, fmt.Errorf("failed to add reward: %w", err)
	}

	// Get the updated collection
	updated, err := q.GetUserCollectionRow(ctx, GetUserCollectionRowParams{
		UserID:         sql.NullString{String: userID, Valid: true},
		CollectionType: sql.NullString{String: collectionType, Valid: true},
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to get updated collection: %w", err)
	}

	collection := GetUserCollection(updated)
	return !hadBefore, collection, nil
}

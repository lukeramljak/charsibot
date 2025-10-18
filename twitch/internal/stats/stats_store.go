package stats

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
)

var ErrInvalidStat = errors.New("invalid stat")

type Store struct {
	db       *sql.DB
	statCols map[string]struct{}
}

func NewStore(db *sql.DB) *Store {
	cols := make(map[string]struct{}, len(statList))
	for _, s := range statList {
		cols[s.Column] = struct{}{}
	}
	return &Store{db: db, statCols: cols}
}

func (s *Store) UpsertAndGet(ctx context.Context, userID, username string) (*Stats, error) {
	var st Stats
	err := s.db.QueryRowContext(ctx, `INSERT INTO stats (id, username) VALUES (?, ?)
        ON CONFLICT(id) DO UPDATE SET username=excluded.username
        RETURNING id, username, strength, intelligence, charisma, luck, dexterity, penis`,
		userID, username,
	).Scan(&st.ID, &st.Username, &st.Strength, &st.Intelligence, &st.Charisma, &st.Luck, &st.Dexterity, &st.Penis)
	if err != nil {
		return nil, fmt.Errorf("upsert stats: %w", err)
	}
	return &st, nil
}

func (s *Store) ModifyStat(ctx context.Context, userID, username, column string, delta int) error {
	if _, ok := s.statCols[column]; !ok {
		return fmt.Errorf("invalid stat column: %s", column)
	}

	res, err := s.db.ExecContext(ctx, "UPDATE stats SET "+column+" = "+column+" + ? WHERE id = ?", delta, userID)
	if err != nil {
		return fmt.Errorf("increment stat: %w", err)
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		if _, errIns := s.db.ExecContext(ctx, "INSERT INTO stats (id, username) VALUES (?, ?)", userID, username); errIns != nil {
			return fmt.Errorf("create stats row: %w", errIns)
		}
		if _, err = s.db.ExecContext(ctx, "UPDATE stats SET "+column+" = "+column+" + ? WHERE id = ?", delta, userID); err != nil {
			return fmt.Errorf("increment stat after insert: %w", err)
		}
	}

	return err
}

func (s *Store) RandomStat(r *rand.Rand) Stat {
	return statList[r.Intn(len(statList))]
}

func (s *Store) RandomDelta(r *rand.Rand) int {
	if r.Intn(20) == 0 {
		return -1
	}
	return 1
}

func (s *Store) Format(username string, st *Stats) string {
	return fmt.Sprintf("%s's stats: STR: %d | INT: %d | CHA: %d | LUCK: %d | DEX: %d | PENIS: %d", username, st.Strength, st.Intelligence, st.Charisma, st.Luck, st.Dexterity, st.Penis)
}

func (s *Store) GetMessage(ctx context.Context, userID, username string) (string, error) {
	st, err := s.UpsertAndGet(ctx, userID, username)
	if err != nil {
		return "", err
	}
	return s.Format(username, st), nil
}

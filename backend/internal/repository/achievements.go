package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

type AchievementsRepository interface {
	FindAchievementsByID(ctx context.Context, id uuid.UUID) (*model.Achievements, error)
	Update (ctx context.Context, userID uuid.UUID, a model.Achievements) error
}

type achievementsRepository struct {
	db *sql.DB
}

func NewAchievementsRepository(db *sql.DB) AchievementsRepository {
	return &achievementsRepository{db: db}
}

func (r *achievementsRepository) FindAchievementsByID(ctx context.Context, id uuid.UUID) (*model.Achievements, error) {
	const q = `
		SELECT achievements
		FROM users WHERE id = $1
		`
	return r.scanAchievements(r.db.QueryRowContext(ctx, q, id.String()))
}

// call when sth new was unlocked&after boolsettotrue
func (r *achievementsRepository) Update(ctx context.Context, userID uuid.UUID, a model.Achievements) error {
    data, err := json.Marshal(a)
    if err != nil {
        return err
    }
    _, err = r.db.ExecContext(ctx,
        `UPDATE users SET achievements = $2, updated_at = now() WHERE id = $1`,
        userID.String(), data,
    )
    return err
}

func (r *achievementsRepository) scanAchievements(row *sql.Row) (*model.Achievements, error) {
	var ajson []byte

	err := row.Scan(
		&ajson,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var a model.Achievements
	err = json.Unmarshal(ajson, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
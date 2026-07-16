package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"backend/internal/model"
)

type TwoFactorRepository interface {
	Create(ctx context.Context, c *model.TwoFactorCode) error
	GetActive(ctx context.Context, userID uuid.UUID, purpose string) (*model.TwoFactorCode, error)
	MarkConsumed(ctx context.Context, id uuid.UUID) error
	DeleteForUser(ctx context.Context, userID uuid.UUID, purpose string) error
}

type twoFactorRepository struct {
	db *sql.DB
}

func NewTwoFactorRepository(db *sql.DB) TwoFactorRepository {
	return &twoFactorRepository{db: db}
}

func (r *twoFactorRepository) Create(ctx context.Context, c *model.TwoFactorCode) error {
	const q = `
		INSERT INTO two_factor_codes (id, user_id, code_hash, purpose, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, q, c.ID.String(), c.UserID.String(), c.CodeHash, c.Purpose, c.ExpiresAt, c.CreatedAt)
	return err
}

func (r *twoFactorRepository) GetActive(ctx context.Context, userID uuid.UUID, purpose string) (*model.TwoFactorCode, error) {
	const q = `
		SELECT id, user_id, code_hash, purpose, expires_at, consumed_at, created_at
		FROM two_factor_codes
		WHERE user_id = $1 AND purpose = $2 AND consumed_at IS NULL AND expires_at > now()
		ORDER BY created_at DESC
		LIMIT 1`
	var c model.TwoFactorCode
	var idStr, uidStr string
	err := r.db.QueryRowContext(ctx, q, userID.String(), purpose).Scan( &idStr, &uidStr, &c.CodeHash, &c.Purpose, &c.ExpiresAt, &c.ConsumedAt, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	c.ID, _ = uuid.Parse(idStr)
	c.UserID, _ = uuid.Parse(uidStr)
	return &c, nil
}

func (r *twoFactorRepository) MarkConsumed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "UPDATE two_factor_codes SET consumed_at = now() WHERE id = $1", id.String())
	return err
}

func (r *twoFactorRepository) DeleteForUser(ctx context.Context, userID uuid.UUID, purpose string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM two_factor_codes WHERE user_id = $1 AND purpose = $2", userID.String(), purpose)
	return err
}

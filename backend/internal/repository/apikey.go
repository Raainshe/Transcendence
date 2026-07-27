package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

type APIKeyRepository interface {
	Create(ctx context.Context, apimod *model.APIKey) error
	FindByHash(ctx context.Context, keyHash string) (*model.APIKey, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]model.APIKeyList, error)
	Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	TouchLastUsed(ctx context.Context, id uuid.UUID) error
}

type apiKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apimod *model.APIKey) error {
	const q = `
		INSERT INTO api_keys (id, user_id, key_hash, name, created_at, last_used_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, q, apimod.ID.String(), apimod.UserID.String(), apimod.KeyHash, apimod.Name, apimod.CreatedAt, apimod.LastUsedAt, apimod.RevokedAt)
	return err
}

func (r *apiKeyRepository) FindByHash(ctx context.Context, keyHash string) (*model.APIKey, error) {
	const q = `
	SELECT id, user_id, key_hash, name, created_at, last_used_at, revoked_at
	FROM api_keys WHERE key_hash = $1`
	return r.scanAPIKey(r.db.QueryRowContext(ctx, q, keyHash))
}

func (r *apiKeyRepository) scanAPIKey(row *sql.Row) (*model.APIKey, error) {
	var a model.APIKey

	err := row.Scan(
		&a.ID, &a.UserID, &a.KeyHash, &a.Name, &a.CreatedAt, &a.LastUsedAt, &a.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *apiKeyRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]model.APIKeyList, error) {
	const q = `
	SELECT id, name, created_at, last_used_at
	FROM api_keys WHERE user_id = $1 AND revoked_at IS NULL`

	rows, err := r.db.QueryContext(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alist []model.APIKeyList

	for rows.Next() {
		var al model.APIKeyList
		var idStr string
		if err := rows.Scan(
			&idStr, &al.Name, &al.CreatedAt, &al.LastUsedAt,
		); err != nil {
			return nil, err
		}
		al.ID, err = uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		alist = append(alist, al)
	}

	return alist, rows.Err()
}

func (r *apiKeyRepository) Revoke(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const q = `
		UPDATE api_keys SET
		revoked_at = now()
		WHERE id = $1 AND user_id = $2
	`
	res, err := r.db.ExecContext(ctx, q, id.String(), userID.String())
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *apiKeyRepository) TouchLastUsed(ctx context.Context, id uuid.UUID) error {
	const q = `
		UPDATE api_keys SET
		last_used_at = now()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, q, id.String())

	return err
}

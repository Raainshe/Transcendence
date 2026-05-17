package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

var ErrNotFound = errors.New("record not found")

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	const q = `
		INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, q,
		user.ID.String(), user.Username, user.Email,
		user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const q = `
		SELECT id, username, email, password_hash, avatar_url, role,
		       is_2fa_enabled, totp_secret, created_at, updated_at, last_seen_at
		FROM users WHERE id = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, q, id.String()))
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `
		SELECT id, username, email, password_hash, avatar_url, role,
		       is_2fa_enabled, totp_secret, created_at, updated_at, last_seen_at
		FROM users WHERE email = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, q, email))
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	const q = `
		SELECT id, username, email, password_hash, avatar_url, role,
		       is_2fa_enabled, totp_secret, created_at, updated_at, last_seen_at
		FROM users WHERE username = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, q, username))
}

func (r *userRepository) scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var idStr string
	err := row.Scan(
		&idStr, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarURL,
		&u.Role, &u.Is2FAEnabled, &u.TOTPSecret,
		&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u.ID, err = uuid.Parse(idStr)
	return &u, err
}

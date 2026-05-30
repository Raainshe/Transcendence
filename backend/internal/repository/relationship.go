package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

type RelationshipRepository interface {
	Create(ctx context.Context, requesterID, receiverID uuid.UUID, status string) (*model.Relationship, error)
	Find(ctx context.Context, userA, userB uuid.UUID) (*model.Relationship, error)
	FindDirectional(ctx context.Context, requesterID, receiverID uuid.UUID) (*model.Relationship, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListFriends(ctx context.Context, userID uuid.UUID) ([]model.User, error)
	ListPendingReceived(ctx context.Context, userID uuid.UUID) ([]model.User, error)
	ListBlocked(ctx context.Context, userID uuid.UUID) ([]model.User, error)
}

type relationshipRepository struct {
	db *sql.DB
}

func NewRelationshipRepository(db *sql.DB) RelationshipRepository {
	return &relationshipRepository{db: db}
}

func (r *relationshipRepository) Create(ctx context.Context, requesterID, receiverID uuid.UUID, status string) (*model.Relationship, error) {
	const q = `
		INSERT INTO relationships (id, requester_id, receiver_id, status, created_at)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, requester_id, receiver_id, status, created_at
	`
	return r.scanRelationship(r.db.QueryRowContext(ctx, q,
		uuid.New().String(), requesterID.String(), receiverID.String(), status,
	))
}

func (r *relationshipRepository) Find(ctx context.Context, userA, userB uuid.UUID) (*model.Relationship, error) {
	const q = `
		SELECT id, requester_id, receiver_id, status, created_at
		FROM relationships
		WHERE (requester_id = $1 AND receiver_id = $2)
		   OR (requester_id = $2 AND receiver_id = $1)
		LIMIT 1
	`
	return r.scanRelationship(r.db.QueryRowContext(ctx, q, userA.String(), userB.String()))
}

func (r *relationshipRepository) FindDirectional(ctx context.Context, requesterID, receiverID uuid.UUID) (*model.Relationship, error) {
	const q = `
		SELECT id, requester_id, receiver_id, status, created_at
		FROM relationships
		WHERE requester_id = $1 AND receiver_id = $2
	`
	return r.scanRelationship(r.db.QueryRowContext(ctx, q, requesterID.String(), receiverID.String()))
}

func (r *relationshipRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE relationships SET status = $2 WHERE id = $1",
		id.String(), status,
	)
	return err
}

func (r *relationshipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM relationships WHERE id = $1", id.String())
	return err
}

func (r *relationshipRepository) ListFriends(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	const q = `
		SELECT u.id, u.username, u.email, u.avatar_url, u.role,
		       u.is_2fa_enabled, u.created_at, u.updated_at, u.last_seen_at
		FROM relationships rel JOIN users u ON u.id = rel.receiver_id
		WHERE rel.requester_id = $1 AND rel.status = 'accepted'
		UNION
		SELECT u.id, u.username, u.email, u.avatar_url, u.role,
		       u.is_2fa_enabled, u.created_at, u.updated_at, u.last_seen_at
		FROM relationships rel JOIN users u ON u.id = rel.requester_id
		WHERE rel.receiver_id = $1 AND rel.status = 'accepted'
	`
	rows, err := r.db.QueryContext(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	return r.scanUsers(rows)
}

func (r *relationshipRepository) ListPendingReceived(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	const q = `
		SELECT u.id, u.username, u.email, u.avatar_url, u.role,
		       u.is_2fa_enabled, u.created_at, u.updated_at, u.last_seen_at
		FROM relationships rel JOIN users u ON u.id = rel.requester_id
		WHERE rel.receiver_id = $1 AND rel.status = 'pending'
	`
	rows, err := r.db.QueryContext(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	return r.scanUsers(rows)
}

func (r *relationshipRepository) ListBlocked(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	const q = `
		SELECT u.id, u.username, u.email, u.avatar_url, u.role,
		       u.is_2fa_enabled, u.created_at, u.updated_at, u.last_seen_at
		FROM relationships rel JOIN users u ON u.id = rel.receiver_id
		WHERE rel.requester_id = $1 AND rel.status = 'blocked'
	`
	rows, err := r.db.QueryContext(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	return r.scanUsers(rows)
}

func (r *relationshipRepository) scanRelationship(row *sql.Row) (*model.Relationship, error) {
	var rel model.Relationship
	var idStr, requesterStr, receiverStr string
	err := row.Scan(&idStr, &requesterStr, &receiverStr, &rel.Status, &rel.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if rel.ID, err = uuid.Parse(idStr); err != nil {
		return nil, err
	}
	if rel.RequesterID, err = uuid.Parse(requesterStr); err != nil {
		return nil, err
	}
	if rel.ReceiverID, err = uuid.Parse(receiverStr); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *relationshipRepository) scanUsers(rows *sql.Rows) ([]model.User, error) {
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		var idStr string
		if err := rows.Scan(
			&idStr, &u.Username, &u.Email, &u.AvatarURL,
			&u.Role, &u.Is2FAEnabled,
			&u.CreatedAt, &u.UpdatedAt, &u.LastSeenAt,
		); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err
		}
		u.ID = id
		users = append(users, u)
	}
	return users, rows.Err()
}

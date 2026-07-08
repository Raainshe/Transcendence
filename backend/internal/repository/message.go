package repository

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"backend/internal/model"
)

type MessageRepository interface {
	Create(ctx context.Context, m *model.Message) error
	ListConversation(ctx context.Context, user1, user2 uuid.UUID, limit int) ([]model.Message, error)
	MarkRead(ctx context.Context, reader, sender uuid.UUID) error
	UnreadCounts(ctx context.Context, userID uuid.UUID) ([]model.UnreadCount, error)
}

type messageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, m *model.Message) error {
	const q = `
		INSERT INTO messages (id, sender_id, recipient_id, body, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, q, m.ID.String(), m.SenderID.String(), m.RecipientID.String(), m.Body, m.CreatedAt)
	return err
}

func (r *messageRepository) ListConversation(ctx context.Context, user1, user2 uuid.UUID, limit int) ([]model.Message, error) {
	const q = `
		SELECT id, sender_id, recipient_id, body, created_at, read_at
		FROM messages
		WHERE (sender_id = $1 AND recipient_id = $2) OR (sender_id = $2 AND recipient_id = $1)
		ORDER BY created_at DESC
		LIMIT $3`
	rows, err := r.db.QueryContext(ctx, q, user1.String(), user2.String(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []model.Message
	for rows.Next() {
		var m model.Message
		var id, s, r string
		if err := rows.Scan(&id, &s, &r, &m.Body, &m.CreatedAt, &m.ReadAt); err != nil {
			return nil, err
		}
		m.ID, _ = uuid.Parse(id)
		m.SenderID, _ = uuid.Parse(s)
		m.RecipientID, _ = uuid.Parse(r)
		ret = append(ret, m)
	}
	return ret, rows.Err()
}

func (r *messageRepository) MarkRead(ctx context.Context, reader, sender uuid.UUID) error {
	const q = `
		UPDATE messages SET read_at = now()
		WHERE recipient_id = $1 AND sender_id = $2 AND read_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, reader.String(), sender.String())
	return err
}

func (r *messageRepository) UnreadCounts(ctx context.Context, userID uuid.UUID) ([]model.UnreadCount, error) {
	const q = `
		SELECT sender_id, COUNT(*)
		FROM messages
		WHERE recipient_id = $1 AND read_at IS NULL
		GROUP BY sender_id`
	rows, err := r.db.QueryContext(ctx, q, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []model.UnreadCount
	for rows.Next() {
		var u model.UnreadCount
		var s string
		if err := rows.Scan(&s, &u.Count); err != nil {
			return nil, err
		}
		u.SenderID, _ = uuid.Parse(s)
		ret = append(ret, u)
	}
	return ret, rows.Err()
}

package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
)

type FileRepository interface {
	Create(ctx context.Context, f *model.FileRecord) error
	FindByPath(ctx context.Context, userID uuid.UUID, urlPath string) (*model.FileRecord, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type fileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, f *model.FileRecord) error {
	const q = `
		INSERT INTO files (id, user_id, filename, mime_type, size, path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, q,
		f.ID.String(), f.UserID.String(), f.Filename, f.MimeType, f.Size, f.Path, f.CreatedAt,
	)
	return err
}

// FindByPath looks up a file record by its filesystem path, scoped to a user.
// The userID scoping prevents one user from operating on another user's file
// even if the path were leaked.
func (r *fileRepository) FindByPath(ctx context.Context, userID uuid.UUID, fsPath string) (*model.FileRecord, error) {
	const q = `
		SELECT id, user_id, filename, mime_type, size, path, created_at
		FROM files WHERE user_id = $1 AND path = $2
	`
	var f model.FileRecord
	var idStr, userIDStr string
	err := r.db.QueryRowContext(ctx, q, userID.String(), fsPath).Scan(
		&idStr, &userIDStr, &f.Filename, &f.MimeType, &f.Size, &f.Path, &f.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if f.ID, err = uuid.Parse(idStr); err != nil {
		return nil, err
	}
	if f.UserID, err = uuid.Parse(userIDStr); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *fileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM files WHERE id = $1`, id.String())
	return err
}

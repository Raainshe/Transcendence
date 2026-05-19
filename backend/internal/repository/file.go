package repository

import (
	"context"
	"database/sql"

	"backend/internal/model"
)

type FileRepository interface {
	Create(ctx context.Context, f *model.FileRecord) error
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

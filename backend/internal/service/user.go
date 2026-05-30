package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

var (
	ErrInvalidFileType    = errors.New("only jpeg, png, webp, and gif images are allowed")
	ErrSelfRelationship   = errors.New("cannot create a relationship with yourself")
	ErrRelationshipExists = errors.New("relationship already exists")
)

type UserService struct {
	users         repository.UserRepository
	files         repository.FileRepository
	relationships repository.RelationshipRepository
	uploadDir     string
}

func NewUserService(users repository.UserRepository, files repository.FileRepository, rels repository.RelationshipRepository, uploadDir string) *UserService {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	return &UserService{users: users, files: files, relationships: rels, uploadDir: uploadDir}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]model.User, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	users, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.users.Count(ctx)
	return users, total, err
}

func (s *UserService) UpdateMe(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error) {
	if req.Username != nil {
		existing, err := s.users.FindByUsername(ctx, *req.Username)
		if err == nil && existing.ID != id {
			return nil, ErrUsernameTaken
		} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}
	return s.users.Update(ctx, id, req)
}

func (s *UserService) DeleteMe(ctx context.Context, id uuid.UUID) error {
	return s.users.Delete(ctx, id)
}

func (s *UserService) UploadAvatar(ctx context.Context, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*model.User, error) {
	// Detect MIME type from first 512 bytes
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	buf = buf[:n]
	mimeType := http.DetectContentType(buf)

	ext, ok := allowedImageTypes[mimeType]
	if !ok {
		return nil, ErrInvalidFileType
	}

	// Re-combine the sniffed bytes with the rest of the file
	combined := io.MultiReader(bytes.NewReader(buf), file)

	// Build destination path
	fileID := uuid.New()
	filename := fileID.String() + ext
	dirPath := filepath.Join(s.uploadDir, "avatars", userID.String())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	destPath := filepath.Join(dirPath, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, combined)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Record in files table
	avatarURL := fmt.Sprintf("/uploads/avatars/%s/%s", userID.String(), filename)
	record := &model.FileRecord{
		ID:        fileID,
		UserID:    userID,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      size,
		Path:      destPath,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.files.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to record file: %w", err)
	}

	// Update user's avatar_url
	return s.users.Update(ctx, userID, model.UpdateUserRequest{AvatarURL: &avatarURL})
}

func (s *UserService) GetFriends(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	return s.relationships.ListFriends(ctx, userID)
}

func (s *UserService) GetPendingRequests(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	return s.relationships.ListPendingReceived(ctx, userID)
}

func (s *UserService) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	return s.relationships.ListBlocked(ctx, userID)
}

func (s *UserService) SendFriendRequest(ctx context.Context, fromID, toID uuid.UUID) error {
	if fromID == toID {
		return ErrSelfRelationship
	}
	existing, err := s.relationships.Find(ctx, fromID, toID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if existing != nil {
		return ErrRelationshipExists
	}
	_, err = s.relationships.Create(ctx, fromID, toID, model.RelationshipPending)
	return err
}

func (s *UserService) AcceptFriendRequest(ctx context.Context, accepterID, requesterID uuid.UUID) error {
	rel, err := s.relationships.FindDirectional(ctx, requesterID, accepterID)
	if err != nil {
		return err
	}
	if rel.Status != model.RelationshipPending {
		return repository.ErrNotFound
	}
	return s.relationships.UpdateStatus(ctx, rel.ID, model.RelationshipAccepted)
}

func (s *UserService) RemoveFriend(ctx context.Context, userID, otherID uuid.UUID) error {
	rel, err := s.relationships.Find(ctx, userID, otherID)
	if err != nil {
		return err
	}
	if rel.Status == model.RelationshipBlocked {
		return repository.ErrNotFound
	}
	return s.relationships.Delete(ctx, rel.ID)
}

func (s *UserService) BlockUser(ctx context.Context, blockerID, targetID uuid.UUID) error {
	if blockerID == targetID {
		return ErrSelfRelationship
	}
	existing, err := s.relationships.FindDirectional(ctx, blockerID, targetID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if existing != nil {
		if existing.Status == model.RelationshipBlocked {
			return ErrRelationshipExists
		}
		return s.relationships.UpdateStatus(ctx, existing.ID, model.RelationshipBlocked)
	}
	// No row in blocker→target direction; remove any reverse row first.
	reverse, err := s.relationships.FindDirectional(ctx, targetID, blockerID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if reverse != nil {
		if err := s.relationships.Delete(ctx, reverse.ID); err != nil {
			return err
		}
	}
	_, err = s.relationships.Create(ctx, blockerID, targetID, model.RelationshipBlocked)
	return err
}

func (s *UserService) UnblockUser(ctx context.Context, blockerID, targetID uuid.UUID) error {
	rel, err := s.relationships.FindDirectional(ctx, blockerID, targetID)
	if err != nil {
		return err
	}
	if rel.Status != model.RelationshipBlocked {
		return repository.ErrNotFound
	}
	return s.relationships.Delete(ctx, rel.ID)
}

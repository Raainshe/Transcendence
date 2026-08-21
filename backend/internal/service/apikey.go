package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

const MaxNameLength = 64

var (
	ErrKeyNameRequired = errors.New("API key name is required")
	ErrKeyNameTooLong  = errors.New("API key name too long")
	ErrKeyRevoked      = errors.New("API key has been revoked")
)

type APIKeyService struct {
	r repository.APIKeyRepository
}

func NewAPIKeyService(r repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{r: r}
}

func (s *APIKeyService) GenerateAPIKey(ctx context.Context, userID uuid.UUID, name string) (*model.APIKeyCreated, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrKeyNameRequired
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return nil, ErrKeyNameTooLong
	}

	keyB64 := base64.RawURLEncoding.EncodeToString(b)
	keyText := "tra_" + keyB64
	sum := sha256.Sum256([]byte(keyText))
	var m model.APIKey

	m.ID = uuid.New()
	m.UserID = userID
	m.KeyHash = hex.EncodeToString(sum[:])
	m.Name = name
	m.CreatedAt = time.Now().UTC()
	if err := s.r.Create(ctx, &m); err != nil {
		return nil, err
	}
	return &model.APIKeyCreated{ID: m.ID, Name: m.Name, Key: keyText}, nil
}

func (s *APIKeyService) Authenticate(ctx context.Context, key string) (uuid.UUID, error) {
	sum := sha256.Sum256([]byte(key))
	hashedKey := hex.EncodeToString(sum[:])

	m, err := s.r.FindByHash(ctx, hashedKey)
	if err != nil {
		return uuid.Nil, err
	}
	if m.RevokedAt != nil {
		return uuid.Nil, ErrKeyRevoked
	}
	go s.r.TouchLastUsed(context.Background(), m.ID)
	return m.UserID, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, keyID uuid.UUID, userID uuid.UUID) error {
	return s.r.Revoke(ctx, keyID, userID)
}

func (s *APIKeyService) List(ctx context.Context, userID uuid.UUID) ([]model.APIKeyList, error) {
	var list []model.APIKeyList

	list, err := s.r.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

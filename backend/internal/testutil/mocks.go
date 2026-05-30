package testutil

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/model"
	"backend/internal/repository"
)

// ── User repo mock ────────────────────────────────────────────────────────────

type MockUserRepo struct {
	CreateFn         func(ctx context.Context, user *model.User) error
	FindByIDFn       func(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByEmailFn    func(ctx context.Context, email string) (*model.User, error)
	FindByUsernameFn func(ctx context.Context, username string) (*model.User, error)
	ListFn           func(ctx context.Context, limit, offset int) ([]model.User, error)
	CountFn          func(ctx context.Context) (int, error)
	UpdateFn         func(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error)
	DeleteFn         func(ctx context.Context, id uuid.UUID) error
}

func (m *MockUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}
func (m *MockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, email)
	}
	return nil, repository.ErrNotFound
}
func (m *MockUserRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.FindByUsernameFn != nil {
		return m.FindByUsernameFn(ctx, username)
	}
	return nil, repository.ErrNotFound
}
func (m *MockUserRepo) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, limit, offset)
	}
	return []model.User{}, nil
}
func (m *MockUserRepo) Count(ctx context.Context) (int, error) {
	if m.CountFn != nil {
		return m.CountFn(ctx)
	}
	return 0, nil
}
func (m *MockUserRepo) Update(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, id, req)
	}
	return &model.User{ID: id}, nil
}
func (m *MockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}

// ── Relationship repo mock ────────────────────────────────────────────────────

type MockRelationshipRepo struct {
	CreateFn              func(ctx context.Context, requesterID, receiverID uuid.UUID, status string) (*model.Relationship, error)
	FindFn                func(ctx context.Context, userA, userB uuid.UUID) (*model.Relationship, error)
	FindDirectionalFn     func(ctx context.Context, requesterID, receiverID uuid.UUID) (*model.Relationship, error)
	UpdateStatusFn        func(ctx context.Context, id uuid.UUID, status string) error
	DeleteFn              func(ctx context.Context, id uuid.UUID) error
	ListFriendsFn         func(ctx context.Context, userID uuid.UUID) ([]model.User, error)
	ListPendingReceivedFn func(ctx context.Context, userID uuid.UUID) ([]model.User, error)
	ListBlockedFn         func(ctx context.Context, userID uuid.UUID) ([]model.User, error)
}

func (m *MockRelationshipRepo) Create(ctx context.Context, requesterID, receiverID uuid.UUID, status string) (*model.Relationship, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, requesterID, receiverID, status)
	}
	return &model.Relationship{ID: uuid.New(), RequesterID: requesterID, ReceiverID: receiverID, Status: status}, nil
}
func (m *MockRelationshipRepo) Find(ctx context.Context, userA, userB uuid.UUID) (*model.Relationship, error) {
	if m.FindFn != nil {
		return m.FindFn(ctx, userA, userB)
	}
	return nil, repository.ErrNotFound
}
func (m *MockRelationshipRepo) FindDirectional(ctx context.Context, requesterID, receiverID uuid.UUID) (*model.Relationship, error) {
	if m.FindDirectionalFn != nil {
		return m.FindDirectionalFn(ctx, requesterID, receiverID)
	}
	return nil, repository.ErrNotFound
}
func (m *MockRelationshipRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *MockRelationshipRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return nil
}
func (m *MockRelationshipRepo) ListFriends(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	if m.ListFriendsFn != nil {
		return m.ListFriendsFn(ctx, userID)
	}
	return []model.User{}, nil
}
func (m *MockRelationshipRepo) ListPendingReceived(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	if m.ListPendingReceivedFn != nil {
		return m.ListPendingReceivedFn(ctx, userID)
	}
	return []model.User{}, nil
}
func (m *MockRelationshipRepo) ListBlocked(ctx context.Context, userID uuid.UUID) ([]model.User, error) {
	if m.ListBlockedFn != nil {
		return m.ListBlockedFn(ctx, userID)
	}
	return []model.User{}, nil
}

// ── Game repo mock ────────────────────────────────────────────────────────────

type MockGameRepo struct {
	RecordMatchFn    func(ctx context.Context, game *model.Game, player *model.GamePlayer) error
	FindByIDFn       func(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListLeaderboardFn func(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
	GetUserStatsFn   func(ctx context.Context, userID uuid.UUID) (*model.UserStats, error)
}

func (m *MockGameRepo) RecordMatch(ctx context.Context, game *model.Game, player *model.GamePlayer) error {
	if m.RecordMatchFn != nil {
		return m.RecordMatchFn(ctx, game, player)
	}
	return nil
}
func (m *MockGameRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockGameRepo) ListLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	if m.ListLeaderboardFn != nil {
		return m.ListLeaderboardFn(ctx, limit)
	}
	return []model.LeaderboardEntry{}, nil
}
func (m *MockGameRepo) GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error) {
	if m.GetUserStatsFn != nil {
		return m.GetUserStatsFn(ctx, userID)
	}
	return &model.UserStats{}, nil
}

// ── File repo mock ────────────────────────────────────────────────────────────

type MockFileRepo struct {
	CreateFn func(ctx context.Context, f *model.FileRecord) error
}

func (m *MockFileRepo) Create(ctx context.Context, f *model.FileRecord) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, f)
	}
	return nil
}

// ── JWT helpers ───────────────────────────────────────────────────────────────

type testClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// MakeTestToken issues a valid signed JWT for use in tests.
func MakeTestToken(userID uuid.UUID, secret string) string {
	c := testClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	return tok
}

// MakeExpiredToken issues a JWT that is already expired.
func MakeExpiredToken(userID uuid.UUID, secret string) string {
	c := testClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	return tok
}

// NewTestUser returns a minimal valid User for use in test setups.
func NewTestUser() *model.User {
	return &model.User{
		ID:        uuid.New(),
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// HashPassword hashes a password at minimum cost — fast for tests.
func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

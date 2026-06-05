package testutil

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/ws"
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
	ClearAvatarFn    func(ctx context.Context, id uuid.UUID) error
	UpdateLastSeenFn func(ctx context.Context, id uuid.UUID) error
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
func (m *MockUserRepo) ClearAvatar(ctx context.Context, id uuid.UUID) error {
	if m.ClearAvatarFn != nil {
		return m.ClearAvatarFn(ctx, id)
	}
	return nil
}
func (m *MockUserRepo) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	if m.UpdateLastSeenFn != nil {
		return m.UpdateLastSeenFn(ctx, id)
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
	RecordMatchFn            func(ctx context.Context, game *model.Game, player *model.GamePlayer) error
	CreateMultiplayerMatchFn func(ctx context.Context, game *model.Game, players []model.GamePlayer) error
	FindByIDFn        func(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListGamesFn       func(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Game, error)
	CountGamesFn      func(ctx context.Context, userID *uuid.UUID) (int, error)
	FindGameDetailFn  func(ctx context.Context, id uuid.UUID) (*model.GameDetail, error)
	ListLeaderboardFn func(ctx context.Context, limit int) ([]model.LeaderboardEntry, error)
	GetUserStatsFn      func(ctx context.Context, userID uuid.UUID) (*model.UserStats, error)
	IsGamePlayerFn      func(ctx context.Context, gameID, userID uuid.UUID) (bool, error)
	ListMatchPlayersFn  func(ctx context.Context, gameID uuid.UUID) ([]model.MatchPlayerView, error)
}

func (m *MockGameRepo) RecordMatch(ctx context.Context, game *model.Game, player *model.GamePlayer) error {
	if m.RecordMatchFn != nil {
		return m.RecordMatchFn(ctx, game, player)
	}
	return nil
}
func (m *MockGameRepo) CreateMultiplayerMatch(ctx context.Context, game *model.Game, players []model.GamePlayer) error {
	if m.CreateMultiplayerMatchFn != nil {
		return m.CreateMultiplayerMatchFn(ctx, game, players)
	}
	return nil
}
func (m *MockGameRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockGameRepo) ListGames(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Game, error) {
	if m.ListGamesFn != nil {
		return m.ListGamesFn(ctx, userID, limit, offset)
	}
	return []model.Game{}, nil
}
func (m *MockGameRepo) CountGames(ctx context.Context, userID *uuid.UUID) (int, error) {
	if m.CountGamesFn != nil {
		return m.CountGamesFn(ctx, userID)
	}
	return 0, nil
}
func (m *MockGameRepo) FindGameDetail(ctx context.Context, id uuid.UUID) (*model.GameDetail, error) {
	if m.FindGameDetailFn != nil {
		return m.FindGameDetailFn(ctx, id)
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
func (m *MockGameRepo) IsGamePlayer(ctx context.Context, gameID, userID uuid.UUID) (bool, error) {
	if m.IsGamePlayerFn != nil {
		return m.IsGamePlayerFn(ctx, gameID, userID)
	}
	return false, nil
}
func (m *MockGameRepo) ListMatchPlayers(ctx context.Context, gameID uuid.UUID) ([]model.MatchPlayerView, error) {
	if m.ListMatchPlayersFn != nil {
		return m.ListMatchPlayersFn(ctx, gameID)
	}
	return []model.MatchPlayerView{}, nil
}

// ── File repo mock ────────────────────────────────────────────────────────────

type MockFileRepo struct {
	CreateFn     func(ctx context.Context, f *model.FileRecord) error
	FindByPathFn func(ctx context.Context, userID uuid.UUID, fsPath string) (*model.FileRecord, error)
	DeleteFn     func(ctx context.Context, id uuid.UUID) error
}

func (m *MockFileRepo) Create(ctx context.Context, f *model.FileRecord) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, f)
	}
	return nil
}
func (m *MockFileRepo) FindByPath(ctx context.Context, userID uuid.UUID, fsPath string) (*model.FileRecord, error) {
	if m.FindByPathFn != nil {
		return m.FindByPathFn(ctx, userID, fsPath)
	}
	return nil, repository.ErrNotFound
}
func (m *MockFileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
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

// ── Lobby repo mock ───────────────────────────────────────────────────────────

type MockLobbyRepo struct {
	CreateFn                  func(ctx context.Context, lobby *model.Lobby, hostUserID uuid.UUID) error
	FindByIDFn                func(ctx context.Context, id uuid.UUID) (*model.Lobby, error)
	FindByInviteCodeFn        func(ctx context.Context, code string) (*model.Lobby, error)
	FindDetailFn              func(ctx context.Context, id uuid.UUID) (*model.LobbyDetail, error)
	FindWaitingLobbyByUserFn  func(ctx context.Context, userID uuid.UUID) (*model.Lobby, error)
	AddMemberFn               func(ctx context.Context, lobbyID, userID uuid.UUID) error
	RemoveMemberFn            func(ctx context.Context, lobbyID, userID uuid.UUID) error
	SetReadyFn                func(ctx context.Context, lobbyID, userID uuid.UUID, ready bool) error
	DeleteLobbyFn             func(ctx context.Context, id uuid.UUID) error
	LinkGameFn                func(ctx context.Context, lobbyID, gameID uuid.UUID, seed int64) error
	MemberCountFn             func(ctx context.Context, lobbyID uuid.UUID) (int, error)
	IsMemberFn                func(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error)
	FindByGameIDFn            func(ctx context.Context, gameID uuid.UUID) (*model.Lobby, error)
}

func (m *MockLobbyRepo) Create(ctx context.Context, lobby *model.Lobby, hostUserID uuid.UUID) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, lobby, hostUserID)
	}
	return nil
}
func (m *MockLobbyRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Lobby, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockLobbyRepo) FindByInviteCode(ctx context.Context, code string) (*model.Lobby, error) {
	if m.FindByInviteCodeFn != nil {
		return m.FindByInviteCodeFn(ctx, code)
	}
	return nil, repository.ErrNotFound
}
func (m *MockLobbyRepo) FindDetail(ctx context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
	if m.FindDetailFn != nil {
		return m.FindDetailFn(ctx, id)
	}
	return nil, repository.ErrNotFound
}
func (m *MockLobbyRepo) FindWaitingLobbyByUser(ctx context.Context, userID uuid.UUID) (*model.Lobby, error) {
	if m.FindWaitingLobbyByUserFn != nil {
		return m.FindWaitingLobbyByUserFn(ctx, userID)
	}
	return nil, repository.ErrNotFound
}
func (m *MockLobbyRepo) AddMember(ctx context.Context, lobbyID, userID uuid.UUID) error {
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, lobbyID, userID)
	}
	return nil
}
func (m *MockLobbyRepo) RemoveMember(ctx context.Context, lobbyID, userID uuid.UUID) error {
	if m.RemoveMemberFn != nil {
		return m.RemoveMemberFn(ctx, lobbyID, userID)
	}
	return nil
}
func (m *MockLobbyRepo) SetReady(ctx context.Context, lobbyID, userID uuid.UUID, ready bool) error {
	if m.SetReadyFn != nil {
		return m.SetReadyFn(ctx, lobbyID, userID, ready)
	}
	return nil
}
func (m *MockLobbyRepo) DeleteLobby(ctx context.Context, id uuid.UUID) error {
	if m.DeleteLobbyFn != nil {
		return m.DeleteLobbyFn(ctx, id)
	}
	return nil
}
func (m *MockLobbyRepo) LinkGame(ctx context.Context, lobbyID, gameID uuid.UUID, seed int64) error {
	if m.LinkGameFn != nil {
		return m.LinkGameFn(ctx, lobbyID, gameID, seed)
	}
	return nil
}
func (m *MockLobbyRepo) MemberCount(ctx context.Context, lobbyID uuid.UUID) (int, error) {
	if m.MemberCountFn != nil {
		return m.MemberCountFn(ctx, lobbyID)
	}
	return 0, nil
}
func (m *MockLobbyRepo) IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error) {
	if m.IsMemberFn != nil {
		return m.IsMemberFn(ctx, lobbyID, userID)
	}
	return false, nil
}
func (m *MockLobbyRepo) FindByGameID(ctx context.Context, gameID uuid.UUID) (*model.Lobby, error) {
	if m.FindByGameIDFn != nil {
		return m.FindByGameIDFn(ctx, gameID)
	}
	return nil, repository.ErrNotFound
}

// ── Lobby broadcaster mock ────────────────────────────────────────────────────

type MockBroadcaster struct {
	BroadcastLobbyFn              func(lobbyID uuid.UUID, env ws.Envelope)
	SubscribeAllInLobbyToMatchFn  func(lobbyID, gameID uuid.UUID)
	Messages                      []ws.Envelope
	AutoJoinCalls                 []struct {
		LobbyID uuid.UUID
		GameID  uuid.UUID
	}
}

func (m *MockBroadcaster) BroadcastLobby(lobbyID uuid.UUID, env ws.Envelope) {
	if m.BroadcastLobbyFn != nil {
		m.BroadcastLobbyFn(lobbyID, env)
		return
	}
	m.Messages = append(m.Messages, env)
}

func (m *MockBroadcaster) SubscribeAllInLobbyToMatch(lobbyID, gameID uuid.UUID) {
	if m.SubscribeAllInLobbyToMatchFn != nil {
		m.SubscribeAllInLobbyToMatchFn(lobbyID, gameID)
		return
	}
	m.AutoJoinCalls = append(m.AutoJoinCalls, struct {
		LobbyID uuid.UUID
		GameID  uuid.UUID
	}{LobbyID: lobbyID, GameID: gameID})
}


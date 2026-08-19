package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/ws"
)

var (
	ErrInvalidMaxPlayers = errors.New("max_players must be between 2 and 4")
	ErrNotLobbyHost      = errors.New("only the lobby host can start the match")
	ErrLobbyNotJoinable  = errors.New("lobby is not joinable")
	ErrLobbyFull         = errors.New("lobby is full")
	ErrAlreadyInLobby    = errors.New("user is already in this lobby")
	ErrInAnotherLobby    = errors.New("user is already in another waiting lobby")
	ErrNotEnoughPlayers  = errors.New("at least 2 players are required to start")
	ErrNotAllReady       = errors.New("all players must be ready before starting")
	ErrNotLobbyMember    = errors.New("user is not a lobby member")
	ErrInvalidLobbyName  = errors.New("lobby name is required")
	ErrLobbyNameTooLong  = errors.New("lobby name must be 64 characters or fewer")
	ErrLobbyNotWaiting   = errors.New("lobby is not accepting changes")
)

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type LobbyBroadcaster interface {
	BroadcastLobby(lobbyID uuid.UUID, env ws.Envelope)
	SubscribeAllInLobbyToMatch(lobbyID, gameID uuid.UUID)
}

type LobbyService struct {
	lobbies      repository.LobbyRepository
	games        repository.GameRepository
	bus          LobbyBroadcaster
	gamification *GamificationService
}

func NewLobbyService(lobbies repository.LobbyRepository, games repository.GameRepository, bus LobbyBroadcaster, gamification *GamificationService) *LobbyService {
	return &LobbyService{lobbies: lobbies, games: games, bus: bus, gamification: gamification}
}

func (s *LobbyService) CreateLobby(ctx context.Context, userID uuid.UUID, maxPlayers int) (*model.LobbyDetail, error) {
	if maxPlayers < 2 || maxPlayers > 4 {
		return nil, ErrInvalidMaxPlayers
	}
	if other, err := s.lobbies.FindWaitingLobbyByUser(ctx, userID); err == nil && other != nil {
		return nil, ErrInAnotherLobby
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	var detail *model.LobbyDetail
	for attempt := 0; attempt < 8; attempt++ {
		code, err := generateInviteCode()
		if err != nil {
			return nil, err
		}
		lobby := &model.Lobby{
			ID:         uuid.New(),
			HostUserID: userID,
			InviteCode: code,
			MaxPlayers: maxPlayers,
			Status:     model.LobbyStatusWaiting,
			CreatedAt:  time.Now().UTC(),
		}
		if err := s.lobbies.Create(ctx, lobby, userID); err != nil {
			if isUniqueViolation(err) {
				continue
			}
			return nil, err
		}
		detail, err = s.lobbies.FindDetail(ctx, lobby.ID)
		if err != nil {
			return nil, err
		}
		break
	}
	if detail == nil {
		return nil, errors.New("failed to generate unique invite code")
	}
	s.broadcastUpdated(ctx, detail.ID)
	return detail, nil
}

func (s *LobbyService) GetLobby(ctx context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
	return s.lobbies.FindDetail(ctx, id)
}

func (s *LobbyService) GetCurrentLobby(ctx context.Context, userID uuid.UUID) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindWaitingLobbyByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.lobbies.FindDetail(ctx, lobby.ID)
}

func (s *LobbyService) GetLobbyByCode(ctx context.Context, code string) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.lobbies.FindDetail(ctx, lobby.ID)
}

func (s *LobbyService) JoinLobby(ctx context.Context, userID, lobbyID uuid.UUID) (*model.LobbyDetail, error) {
	return s.join(ctx, userID, lobbyID)
}

func (s *LobbyService) JoinLobbyByCode(ctx context.Context, userID uuid.UUID, code string) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindByInviteCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.join(ctx, userID, lobby.ID)
}

func (s *LobbyService) join(ctx context.Context, userID, lobbyID uuid.UUID) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if lobby.Status != model.LobbyStatusWaiting {
		return nil, ErrLobbyNotJoinable
	}

	isMember, err := s.lobbies.IsMember(ctx, lobbyID, userID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return s.lobbies.FindDetail(ctx, lobbyID)
	}

	if other, err := s.lobbies.FindWaitingLobbyByUser(ctx, userID); err == nil && other != nil && other.ID != lobbyID {
		return nil, ErrInAnotherLobby
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	count, err := s.lobbies.MemberCount(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if count >= lobby.MaxPlayers {
		return nil, ErrLobbyFull
	}

	if err := s.lobbies.AddMember(ctx, lobbyID, userID); err != nil {
		return nil, err
	}

	detail, err := s.lobbies.FindDetail(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	s.broadcastUpdated(ctx, lobbyID)
	return detail, nil
}

func (s *LobbyService) LeaveLobby(ctx context.Context, userID, lobbyID uuid.UUID) error {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return err
	}

	if lobby.HostUserID == userID {
		if err := s.lobbies.DeleteLobby(ctx, lobbyID); err != nil {
			return err
		}
		s.broadcastClosed(lobbyID, "host_left")
		return nil
	}

	if err := s.lobbies.RemoveMember(ctx, lobbyID, userID); err != nil {
		return err
	}
	s.broadcastUpdated(ctx, lobbyID)
	return nil
}

func (s *LobbyService) SetReady(ctx context.Context, userID, lobbyID uuid.UUID, ready bool) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if lobby.Status != model.LobbyStatusWaiting {
		return nil, ErrLobbyNotJoinable
	}
	if err := s.lobbies.SetReady(ctx, lobbyID, userID, ready); err != nil {
		if errors.Is(err, repository.ErrNotLobbyMember) {
			return nil, ErrNotLobbyMember
		}
		return nil, err
	}
	detail, err := s.lobbies.FindDetail(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	s.broadcastUpdated(ctx, lobbyID)
	return detail, nil
}

func (s *LobbyService) StartLobby(ctx context.Context, hostID, lobbyID uuid.UUID) (*model.StartLobbyResult, error) {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if lobby.HostUserID != hostID {
		return nil, ErrNotLobbyHost
	}
	if lobby.Status != model.LobbyStatusWaiting {
		return nil, ErrLobbyNotJoinable
	}

	detail, err := s.lobbies.FindDetail(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if len(detail.Members) < 2 {
		return nil, ErrNotEnoughPlayers
	}
	for _, m := range detail.Members {
		if !m.IsReady {
			return nil, ErrNotAllReady
		}
	}

	seed, err := randomSeed()
	if err != nil {
		return nil, err
	}

	gameID := uuid.New()
	now := time.Now().UTC()
	game := &model.Game{
		ID:        gameID,
		Mode:      "multiplayer",
		Status:    "in_progress",
		CreatedAt: now,
	}

	players := make([]model.GamePlayer, 0, len(detail.Members))
	matchPlayers := make([]model.MatchStartPlayer, 0, len(detail.Members))
	for _, m := range detail.Members {
		players = append(players, model.GamePlayer{
			ID:           uuid.New(),
			GameID:       gameID,
			UserID:       m.UserID,
			Score:        0,
			LinesCleared: 0,
			LevelReached: 1,
			IsWinner:     false,
		})
		matchPlayers = append(matchPlayers, model.MatchStartPlayer{
			UserID:    m.UserID,
			Username:  m.Username,
			AvatarURL: m.AvatarURL,
		})
	}

	if err := s.games.CreateMultiplayerMatch(ctx, game, players); err != nil {
		return nil, err
	}
	if err := s.lobbies.LinkGame(ctx, lobbyID, gameID, seed); err != nil {
		return nil, err
	}

	result := &model.StartLobbyResult{
		GameID:     gameID,
		SharedSeed: seed,
		Players:    matchPlayers,
	}
	for _, p := range players {
		err := s.gamification.OnMPGame(ctx, p.UserID)
		if err != nil {
			log.Printf("achievement check failed for user %s: %v", p.UserID, err)
		}
	}

	s.broadcastMatchStart(lobbyID, result)
	s.broadcastClosed(lobbyID, "started")
	return result, nil
}

func (s *LobbyService) broadcastUpdated(ctx context.Context, lobbyID uuid.UUID) {
	if s.bus == nil {
		return
	}
	detail, err := s.lobbies.FindDetail(ctx, lobbyID)
	if err != nil {
		return
	}
	env, err := ws.NewEnvelope(ws.TypeLobbyUpdated, map[string]any{"lobby": detail})
	if err != nil {
		return
	}
	s.bus.BroadcastLobby(lobbyID, env)
}

func (s *LobbyService) broadcastClosed(lobbyID uuid.UUID, reason string) {
	if s.bus == nil {
		return
	}
	env, err := ws.NewEnvelope(ws.TypeLobbyClosed, map[string]string{"reason": reason})
	if err != nil {
		return
	}
	s.bus.BroadcastLobby(lobbyID, env)
}

func (s *LobbyService) broadcastMatchStart(lobbyID uuid.UUID, result *model.StartLobbyResult) {
	if s.bus == nil {
		return
	}
	env, err := ws.NewEnvelope(ws.TypeMatchStart, result)
	if err != nil {
		return
	}
	s.bus.BroadcastLobby(lobbyID, env)
	s.bus.SubscribeAllInLobbyToMatch(lobbyID, result.GameID)
}

func generateInviteCode() (string, error) {
	b := make([]byte, 6)
	max := big.NewInt(int64(len(inviteCodeChars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = inviteCodeChars[n.Int64()]
	}
	return string(b), nil
}

func randomSeed() (int64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<31-1))
	if err != nil {
		return 0, err
	}
	v := n.Int64()
	if v == 0 {
		return 1, nil
	}
	return v, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func (s *LobbyService) IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error) {
	return s.lobbies.IsMember(ctx, lobbyID, userID)
}

func (s *LobbyService) CloseLobby(ctx context.Context, hostID, lobbyID uuid.UUID) error {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return err
	}
	if lobby.HostUserID != hostID {
		return ErrNotLobbyHost
	}
	if err := s.lobbies.DeleteLobby(ctx, lobbyID); err != nil {
		return err
	}
	s.broadcastClosed(lobbyID, "closed_by_host")
	return nil
}

func (s *LobbyService) UpdateLobbyName(ctx context.Context, hostID, lobbyID uuid.UUID, name string) (*model.LobbyDetail, error) {
	lobby, err := s.lobbies.FindByID(ctx, lobbyID)
	if err != nil {
		return nil, err
	}
	if lobby.HostUserID != hostID {
		return nil, ErrNotLobbyHost
	}
	if lobby.Status != model.LobbyStatusWaiting {
		return nil, ErrLobbyNotWaiting
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidLobbyName
	}
	if len(name) > 64 {
		return nil, ErrLobbyNameTooLong
	}
	err = s.lobbies.UpdateLobbyName(ctx, lobbyID, name)
	if err != nil {
		return nil, err
	}
	s.broadcastUpdated(ctx, lobbyID)
	return s.lobbies.FindDetail(ctx, lobbyID)
}

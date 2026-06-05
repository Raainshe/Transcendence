package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/testutil"
	"backend/internal/ws"
)

func newLobbyService(t *testing.T, lobby *testutil.MockLobbyRepo, game *testutil.MockGameRepo, bus *testutil.MockBroadcaster) *service.LobbyService {
	t.Helper()
	if lobby == nil {
		lobby = &testutil.MockLobbyRepo{}
	}
	if game == nil {
		game = &testutil.MockGameRepo{}
	}
	if bus == nil {
		bus = &testutil.MockBroadcaster{}
	}
	return service.NewLobbyService(lobby, game, bus)
}

func waitingLobby(hostID uuid.UUID) *model.Lobby {
	return &model.Lobby{
		ID:         uuid.New(),
		HostUserID: hostID,
		InviteCode: "ABC123",
		MaxPlayers: 4,
		Status:     model.LobbyStatusWaiting,
		CreatedAt:  time.Now().UTC(),
	}
}

func lobbyDetail(lobby *model.Lobby, members []model.LobbyMemberView) *model.LobbyDetail {
	return &model.LobbyDetail{Lobby: *lobby, Members: members}
}

func TestLobbyService_StartLobby_NotHost(t *testing.T) {
	hostID := uuid.New()
	callerID := uuid.New()
	lobby := waitingLobby(hostID)

	repo := &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			if id == lobby.ID {
				return lobby, nil
			}
			return nil, repository.ErrNotFound
		},
	}
	svc := newLobbyService(t, repo, nil, nil)

	_, err := svc.StartLobby(context.Background(), callerID, lobby.ID)
	if err != service.ErrNotLobbyHost {
		t.Fatalf("err = %v, want ErrNotLobbyHost", err)
	}
}

func TestLobbyService_StartLobby_NotEnoughPlayers(t *testing.T) {
	hostID := uuid.New()
	lobby := waitingLobby(hostID)

	repo := &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			return lobby, nil
		},
		FindDetailFn: func(_ context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
			return lobbyDetail(lobby, []model.LobbyMemberView{
				{UserID: hostID, Username: "host", IsReady: true},
			}), nil
		},
	}
	svc := newLobbyService(t, repo, nil, nil)

	_, err := svc.StartLobby(context.Background(), hostID, lobby.ID)
	if err != service.ErrNotEnoughPlayers {
		t.Fatalf("err = %v, want ErrNotEnoughPlayers", err)
	}
}

func TestLobbyService_StartLobby_NotAllReady(t *testing.T) {
	hostID := uuid.New()
	guestID := uuid.New()
	lobby := waitingLobby(hostID)

	repo := &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			return lobby, nil
		},
		FindDetailFn: func(_ context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
			return lobbyDetail(lobby, []model.LobbyMemberView{
				{UserID: hostID, Username: "host", IsReady: true},
				{UserID: guestID, Username: "guest", IsReady: false},
			}), nil
		},
	}
	svc := newLobbyService(t, repo, nil, nil)

	_, err := svc.StartLobby(context.Background(), hostID, lobby.ID)
	if err != service.ErrNotAllReady {
		t.Fatalf("err = %v, want ErrNotAllReady", err)
	}
}

func TestLobbyService_LeaveLobby_HostDeletes(t *testing.T) {
	hostID := uuid.New()
	lobby := waitingLobby(hostID)
	deleted := false
	broadcasts := 0

	repo := &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			return lobby, nil
		},
		DeleteLobbyFn: func(_ context.Context, id uuid.UUID) error {
			if id == lobby.ID {
				deleted = true
			}
			return nil
		},
	}
	bus := &testutil.MockBroadcaster{
		BroadcastLobbyFn: func(_ uuid.UUID, _ ws.Envelope) {
			broadcasts++
		},
	}
	svc := newLobbyService(t, repo, nil, bus)

	if err := svc.LeaveLobby(context.Background(), hostID, lobby.ID); err != nil {
		t.Fatalf("LeaveLobby: %v", err)
	}
	if !deleted {
		t.Fatal("expected lobby to be deleted when host leaves")
	}
	if broadcasts != 1 {
		t.Fatalf("broadcasts = %d, want 1", broadcasts)
	}
}

func TestLobbyService_StartLobby_Success(t *testing.T) {
	hostID := uuid.New()
	guestID := uuid.New()
	lobby := waitingLobby(hostID)
	linked := false
	created := false

	repo := &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			return lobby, nil
		},
		FindDetailFn: func(_ context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
			return lobbyDetail(lobby, []model.LobbyMemberView{
				{UserID: hostID, Username: "host", IsReady: true},
				{UserID: guestID, Username: "guest", IsReady: true},
			}), nil
		},
		LinkGameFn: func(_ context.Context, lobbyID, gameID uuid.UUID, seed int64) error {
			linked = true
			return nil
		},
	}
	gameRepo := &testutil.MockGameRepo{
		CreateMultiplayerMatchFn: func(_ context.Context, _ *model.Game, players []model.GamePlayer) error {
			if len(players) != 2 {
				t.Fatalf("players = %d, want 2", len(players))
			}
			created = true
			return nil
		},
	}
	bus := &testutil.MockBroadcaster{}
	svc := newLobbyService(t, repo, gameRepo, bus)

	result, err := svc.StartLobby(context.Background(), hostID, lobby.ID)
	if err != nil {
		t.Fatalf("StartLobby: %v", err)
	}
	if result.GameID == uuid.Nil || result.SharedSeed == 0 {
		t.Fatalf("result = %+v, want game_id and shared_seed", result)
	}
	if !linked || !created {
		t.Fatalf("linked=%v created=%v, want both true", linked, created)
	}
	if len(bus.Messages) < 2 {
		t.Fatalf("broadcast messages = %d, want at least 2", len(bus.Messages))
	}
}

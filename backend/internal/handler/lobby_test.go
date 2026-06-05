package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/handler"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/testutil"
)

func newLobbyHandler(t *testing.T, lobbyRepo *testutil.MockLobbyRepo, gameRepo *testutil.MockGameRepo) *handler.LobbyHandler {
	t.Helper()
	if lobbyRepo == nil {
		lobbyRepo = &testutil.MockLobbyRepo{}
	}
	if gameRepo == nil {
		gameRepo = &testutil.MockGameRepo{}
	}
	svc := service.NewLobbyService(lobbyRepo, gameRepo, &testutil.MockBroadcaster{})
	return handler.NewLobbyHandler(svc)
}

func TestLobbyHandler_Create_Unauthorized(t *testing.T) {
	h := newLobbyHandler(t, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/lobbies", strings.NewReader(`{"max_players":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	serveProtected(http.MethodPost, "/lobbies", h.Create).ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestLobbyHandler_Create_InvalidMaxPlayers(t *testing.T) {
	h := newLobbyHandler(t, &testutil.MockLobbyRepo{
		FindWaitingLobbyByUserFn: func(_ context.Context, _ uuid.UUID) (*model.Lobby, error) {
			return nil, repository.ErrNotFound
		},
	}, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/lobbies", strings.NewReader(`{"max_players":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(userID, testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodPost, "/lobbies", h.Create).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestLobbyHandler_Get_BadUUID(t *testing.T) {
	h := newLobbyHandler(t, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/lobbies/not-a-uuid", nil)
	req = withChiParam(req, "id", "not-a-uuid")
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(uuid.New(), testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodGet, "/lobbies/{id}", h.Get).ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestLobbyHandler_Get_NotFound(t *testing.T) {
	h := newLobbyHandler(t, &testutil.MockLobbyRepo{
		FindDetailFn: func(_ context.Context, _ uuid.UUID) (*model.LobbyDetail, error) {
			return nil, repository.ErrNotFound
		},
	}, nil)

	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/lobbies/"+id.String(), nil)
	req = withChiParam(req, "id", id.String())
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(uuid.New(), testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodGet, "/lobbies/{id}", h.Get).ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestLobbyHandler_Start_ForbiddenForNonHost(t *testing.T) {
	hostID := uuid.New()
	lobbyID := uuid.New()
	h := newLobbyHandler(t, &testutil.MockLobbyRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			return &model.Lobby{
				ID:         lobbyID,
				HostUserID: hostID,
				Status:     model.LobbyStatusWaiting,
				CreatedAt:  time.Now().UTC(),
			}, nil
		},
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/lobbies/"+lobbyID.String()+"/start", nil)
	req = withChiParam(req, "id", lobbyID.String())
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(uuid.New(), testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodPost, "/lobbies/{id}/start", h.Start).ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (body: %s)", w.Code, w.Body.String())
	}
}

func TestLobbyHandler_JoinByCode_Success(t *testing.T) {
	userID := uuid.New()
	lobbyID := uuid.New()
	h := newLobbyHandler(t, &testutil.MockLobbyRepo{
		FindByInviteCodeFn: func(_ context.Context, code string) (*model.Lobby, error) {
			if code == "ABC123" {
				return &model.Lobby{
					ID:         lobbyID,
					HostUserID: uuid.New(),
					InviteCode: "ABC123",
					MaxPlayers: 4,
					Status:     model.LobbyStatusWaiting,
					CreatedAt:  time.Now().UTC(),
				}, nil
			}
			return nil, repository.ErrNotFound
		},
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Lobby, error) {
			if id == lobbyID {
				return &model.Lobby{
					ID:         lobbyID,
					HostUserID: uuid.New(),
					InviteCode: "ABC123",
					MaxPlayers: 4,
					Status:     model.LobbyStatusWaiting,
					CreatedAt:  time.Now().UTC(),
				}, nil
			}
			return nil, repository.ErrNotFound
		},
		IsMemberFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) {
			return false, nil
		},
		FindWaitingLobbyByUserFn: func(_ context.Context, _ uuid.UUID) (*model.Lobby, error) {
			return nil, repository.ErrNotFound
		},
		MemberCountFn: func(_ context.Context, _ uuid.UUID) (int, error) {
			return 1, nil
		},
		AddMemberFn: func(_ context.Context, _, _ uuid.UUID) error {
			return nil
		},
		FindDetailFn: func(_ context.Context, id uuid.UUID) (*model.LobbyDetail, error) {
			return &model.LobbyDetail{
				Lobby: model.Lobby{ID: id, InviteCode: "ABC123", Status: model.LobbyStatusWaiting},
				Members: []model.LobbyMemberView{
					{UserID: userID, Username: "guest", IsReady: false},
				},
			}, nil
		},
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/lobbies/join", strings.NewReader(`{"invite_code":"ABC123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(userID, testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodPost, "/lobbies/join", h.JoinByCode).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	assertJSONKeys(t, w.Body.Bytes(), []string{"lobby"})
}

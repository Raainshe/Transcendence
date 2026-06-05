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

func newMatchHandler(t *testing.T, gameRepo *testutil.MockGameRepo, lobbyRepo *testutil.MockLobbyRepo) *handler.MatchHandler {
	t.Helper()
	if gameRepo == nil {
		gameRepo = &testutil.MockGameRepo{}
	}
	if lobbyRepo == nil {
		lobbyRepo = &testutil.MockLobbyRepo{}
	}
	svc := service.NewMatchService(gameRepo, lobbyRepo)
	return handler.NewMatchHandler(svc)
}

func TestMatchHandler_Get_MemberSuccess(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()
	seed := int64(424242)

	gameRepo := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{
				ID:     gameID,
				Mode:   "multiplayer",
				Status: "in_progress",
			}, nil
		},
		IsGamePlayerFn: func(_ context.Context, gid, uid uuid.UUID) (bool, error) {
			return gid == gameID && uid == userID, nil
		},
		ListMatchPlayersFn: func(_ context.Context, gid uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{
				{UserID: userID, Username: "host"},
			}, nil
		},
	}
	lobbyRepo := &testutil.MockLobbyRepo{
		FindByGameIDFn: func(_ context.Context, gid uuid.UUID) (*model.Lobby, error) {
			return &model.Lobby{
				ID:         uuid.New(),
				GameID:     &gameID,
				SharedSeed: &seed,
				Status:     "closed",
				CreatedAt:  time.Now().UTC(),
			}, nil
		},
	}

	h := newMatchHandler(t, gameRepo, lobbyRepo)
	req := httptest.NewRequest(http.MethodGet, "/matches/"+gameID.String(), nil)
	req = withChiParam(req, "id", gameID.String())
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(userID, testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodGet, "/matches/{id}", h.Get).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, `"shared_seed":424242`) || !strings.Contains(body, `"status":"in_progress"`) {
		t.Fatalf("body = %s, want match metadata", body)
	}
}

func TestMatchHandler_Get_NonMemberReturnsNull(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()

	gameRepo := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: gameID, Mode: "multiplayer", Status: "in_progress"}, nil
		},
		IsGamePlayerFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}

	h := newMatchHandler(t, gameRepo, nil)
	req := httptest.NewRequest(http.MethodGet, "/matches/"+gameID.String(), nil)
	req = withChiParam(req, "id", gameID.String())
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(userID, testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodGet, "/matches/{id}", h.Get).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"match":null`) && !strings.Contains(body, `"match": null`) {
		t.Fatalf("body = %s, want null match", body)
	}
}

func TestMatchHandler_Get_NotFoundReturnsNull(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()

	gameRepo := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, _ uuid.UUID) (*model.Game, error) {
			return nil, repository.ErrNotFound
		},
	}

	h := newMatchHandler(t, gameRepo, nil)
	req := httptest.NewRequest(http.MethodGet, "/matches/"+gameID.String(), nil)
	req = withChiParam(req, "id", gameID.String())
	req.Header.Set("Authorization", "Bearer "+testutil.MakeTestToken(userID, testSecret))

	w := httptest.NewRecorder()
	serveProtected(http.MethodGet, "/matches/{id}", h.Get).ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"match":null`) && !strings.Contains(body, `"match": null`) {
		t.Fatalf("body = %s, want null match", body)
	}
}

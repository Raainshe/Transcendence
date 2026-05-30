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
	"backend/internal/service"
	"backend/internal/testutil"
)

func newGameHandler(t *testing.T, repo *testutil.MockGameRepo) *handler.GameHandler {
	t.Helper()
	return handler.NewGameHandler(service.NewGameService(repo))
}

func TestGameHandler_CreateGame(t *testing.T) {
	now := time.Now()
	validBody := `{"mode":"marathon","score":1000,"lines_cleared":40,"level_reached":5,"started_at":"` +
		now.Add(-5*time.Minute).Format(time.RFC3339) + `","finished_at":"` +
		now.Format(time.RFC3339) + `","is_winner":true}`

	tests := []struct {
		name          string
		body          string
		recordMatchFn func(context.Context, *model.Game, *model.GamePlayer) error
		wantStatus    int
		wantKeys      []string
	}{
		{
			name:       "valid mode",
			body:       validBody,
			wantStatus: http.StatusCreated,
			wantKeys:   []string{"game"},
		},
		{
			name:       "invalid mode",
			body:       `{"mode":"invalid","score":100,"lines_cleared":10,"level_reached":1,"started_at":"2024-01-01T00:00:00Z","finished_at":"2024-01-01T00:01:00Z"}`,
			wantStatus: http.StatusBadRequest,
			wantKeys:   []string{"error"},
		},
		{
			name:       "invalid JSON",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
			wantKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{RecordMatchFn: tt.recordMatchFn}
			h := newGameHandler(t, repo)

			req := httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			userID := uuid.New()
			token := testutil.MakeTestToken(userID, testSecret)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			serveProtected(http.MethodPost, "/games", h.CreateGame).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestGameHandler_ListGames(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantKeys   []string
	}{
		{name: "success no params", query: "", wantStatus: http.StatusOK, wantKeys: []string{"games", "total"}},
		{name: "success with valid user_id + pagination", query: "?user_id=" + uuid.New().String() + "&limit=5&offset=10", wantStatus: http.StatusOK, wantKeys: []string{"games", "total"}},
		{name: "bad user_id", query: "?user_id=not-a-uuid", wantStatus: http.StatusBadRequest, wantKeys: []string{"error"}},
		{name: "bad limit", query: "?limit=abc", wantStatus: http.StatusBadRequest, wantKeys: []string{"error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{}
			h := newGameHandler(t, repo)
			req := httptest.NewRequest(http.MethodGet, "/games"+tt.query, nil)
			w := httptest.NewRecorder()
			h.ListGames(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestGameHandler_GetGame(t *testing.T) {
	validID := uuid.New()
	tests := []struct {
		name             string
		idParam          string
		findGameDetailFn func(context.Context, uuid.UUID) (*model.GameDetail, error)
		wantStatus       int
		wantKeys         []string
	}{
		{
			name:    "success",
			idParam: validID.String(),
			findGameDetailFn: func(_ context.Context, _ uuid.UUID) (*model.GameDetail, error) {
				return &model.GameDetail{Game: model.Game{ID: validID, Mode: "marathon"}}, nil
			},
			wantStatus: http.StatusOK,
			wantKeys:   []string{"game"},
		},
		{name: "invalid UUID", idParam: "not-a-uuid", wantStatus: http.StatusBadRequest, wantKeys: []string{"error"}},
		{name: "not found", idParam: validID.String(), wantStatus: http.StatusNotFound, wantKeys: []string{"error"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{FindGameDetailFn: tt.findGameDetailFn}
			h := newGameHandler(t, repo)
			req := httptest.NewRequest(http.MethodGet, "/games/"+tt.idParam, nil)
			req = withChiParam(req, "id", tt.idParam)
			w := httptest.NewRecorder()
			h.GetGame(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestGameHandler_GetUserStats(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		idParam        string
		getUserStatsFn func(context.Context, uuid.UUID) (*model.UserStats, error)
		wantStatus     int
	}{
		{
			name:    "valid UUID",
			idParam: validID.String(),
			getUserStatsFn: func(_ context.Context, _ uuid.UUID) (*model.UserStats, error) {
				return &model.UserStats{GamesPlayed: 5, Wins: 3, BestScore: 9000}, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID",
			idParam:    "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{GetUserStatsFn: tt.getUserStatsFn}
			h := newGameHandler(t, repo)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.idParam+"/stats", nil)
			req = withChiParam(req, "id", tt.idParam)
			w := httptest.NewRecorder()
			h.GetUserStats(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestGameHandler_GetLeaderboard(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{
			name:       "no limit param",
			query:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "with limit",
			query:      "?limit=5",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{}
			h := newGameHandler(t, repo)

			req := httptest.NewRequest(http.MethodGet, "/leaderboard"+tt.query, nil)
			w := httptest.NewRecorder()
			h.GetLeaderboard(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

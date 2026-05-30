package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/testutil"
)

const testSecret = "test-secret"

// newUserHandler builds a UserHandler backed by mock repos.
func newUserHandler(t *testing.T, users *testutil.MockUserRepo, rels *testutil.MockRelationshipRepo) *handler.UserHandler {
	t.Helper()
	if users == nil {
		users = &testutil.MockUserRepo{}
	}
	if rels == nil {
		rels = &testutil.MockRelationshipRepo{}
	}
	svc := service.NewUserService(users, &testutil.MockFileRepo{}, rels, t.TempDir())
	return handler.NewUserHandler(svc)
}

// serveProtected wraps a handler method behind the JWT middleware on a chi router.
// The returned ServeHTTP can be called with a request that includes an Authorization header.
func serveProtected(method, pattern string, hfn http.HandlerFunc) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.JWTAuth(testSecret))
	r.Method(method, pattern, hfn)
	return r
}

// withChiParam injects a chi URL parameter into the request context (for handlers called directly).
func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// ── Public handlers ───────────────────────────────────────────────────────────

func TestUserHandler_ListUsers(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		listFn     func(context.Context, int, int) ([]model.User, error)
		countFn    func(context.Context) (int, error)
		wantStatus int
		wantKeys   []string
	}{
		{
			name:       "success defaults",
			wantStatus: http.StatusOK,
			wantKeys:   []string{"users", "total"},
		},
		{
			name:  "success with pagination params",
			query: "?limit=5&offset=10",
			listFn: func(_ context.Context, limit, offset int) ([]model.User, error) {
				if limit != 5 {
					return nil, nil
				}
				return []model.User{}, nil
			},
			wantStatus: http.StatusOK,
			wantKeys:   []string{"users", "total"},
		},
		{
			name: "db error returns 500",
			listFn: func(_ context.Context, _, _ int) ([]model.User, error) {
				return nil, context.DeadlineExceeded
			},
			wantStatus: http.StatusInternalServerError,
			wantKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{ListFn: tt.listFn, CountFn: tt.countFn}
			h := newUserHandler(t, repo, nil)

			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			w := httptest.NewRecorder()
			h.ListUsers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	user := testutil.NewTestUser()

	tests := []struct {
		name       string
		idParam    string
		findByIDFn func(context.Context, uuid.UUID) (*model.User, error)
		wantStatus int
	}{
		{
			name:    "success",
			idParam: user.ID.String(),
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
				return user, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid UUID",
			idParam:    "not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "user not found",
			idParam: uuid.New().String(),
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
				return nil, repository.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByIDFn: tt.findByIDFn}
			h := newUserHandler(t, repo, nil)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.idParam, nil)
			req = withChiParam(req, "id", tt.idParam)
			w := httptest.NewRecorder()
			h.GetUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ── Protected handlers ────────────────────────────────────────────────────────

func TestUserHandler_GetMe(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)

	tests := []struct {
		name       string
		findByIDFn func(context.Context, uuid.UUID) (*model.User, error)
		wantStatus int
	}{
		{
			name: "success",
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
				return user, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "user not found",
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
				return nil, repository.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByIDFn: tt.findByIDFn}
			h := newUserHandler(t, repo, nil)
			srv := serveProtected(http.MethodGet, "/users/me", h.GetMe)

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUserHandler_UpdateMe(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)

	tests := []struct {
		name             string
		body             string
		findByUsernameFn func(context.Context, string) (*model.User, error)
		wantStatus       int
	}{
		{
			name:       "success",
			body:       `{"username":"newname"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid JSON",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "username taken",
			body: `{"username":"taken"}`,
			findByUsernameFn: func(_ context.Context, _ string) (*model.User, error) {
				other := testutil.NewTestUser()
				return other, nil
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{
				FindByUsernameFn: tt.findByUsernameFn,
				UpdateFn: func(_ context.Context, id uuid.UUID, _ model.UpdateUserRequest) (*model.User, error) {
					return &model.User{ID: id}, nil
				},
			}
			h := newUserHandler(t, repo, nil)
			srv := serveProtected(http.MethodPatch, "/users/me", h.UpdateMe)

			req := httptest.NewRequest(http.MethodPatch, "/users/me", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestUserHandler_DeleteMe(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)

	tests := []struct {
		name       string
		deleteFn   func(context.Context, uuid.UUID) error
		wantStatus int
	}{
		{name: "success", wantStatus: http.StatusNoContent},
		{
			name: "db error",
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return context.DeadlineExceeded
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{DeleteFn: tt.deleteFn}
			h := newUserHandler(t, repo, nil)
			srv := serveProtected(http.MethodDelete, "/users/me", h.DeleteMe)

			req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUserHandler_GetFriends(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)

	tests := []struct {
		name           string
		listFriendsFn  func(context.Context, uuid.UUID) ([]model.User, error)
		wantStatus     int
		wantFriendsKey bool
	}{
		{
			name:           "returns empty list not null",
			wantStatus:     http.StatusOK,
			wantFriendsKey: true,
		},
		{
			name: "returns friends",
			listFriendsFn: func(_ context.Context, _ uuid.UUID) ([]model.User, error) {
				return []model.User{*testutil.NewTestUser()}, nil
			},
			wantStatus:     http.StatusOK,
			wantFriendsKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{ListFriendsFn: tt.listFriendsFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodGet, "/users/me/friends", h.GetFriends)

			req := httptest.NewRequest(http.MethodGet, "/users/me/friends", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantFriendsKey {
				assertJSONKeys(t, w.Body.Bytes(), []string{"friends"})
				// Ensure the value is an array, not null.
				var body map[string]json.RawMessage
				json.Unmarshal(w.Body.Bytes(), &body)
				if string(body["friends"]) == "null" {
					t.Error(`"friends" must not be null`)
				}
			}
		})
	}
}

func TestUserHandler_GetPendingRequests(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	rels := &testutil.MockRelationshipRepo{}
	h := newUserHandler(t, nil, rels)
	srv := serveProtected(http.MethodGet, "/users/me/friends/requests", h.GetPendingRequests)

	req := httptest.NewRequest(http.MethodGet, "/users/me/friends/requests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	assertJSONKeys(t, w.Body.Bytes(), []string{"requests"})
}

func TestUserHandler_GetBlockedUsers(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	rels := &testutil.MockRelationshipRepo{}
	h := newUserHandler(t, nil, rels)
	srv := serveProtected(http.MethodGet, "/users/me/blocks", h.GetBlockedUsers)

	req := httptest.NewRequest(http.MethodGet, "/users/me/blocks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	assertJSONKeys(t, w.Body.Bytes(), []string{"blocked"})
}

func TestUserHandler_AddFriend(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	friendID := uuid.New()

	tests := []struct {
		name       string
		idParam    string
		findFn     func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantStatus int
	}{
		{
			name:       "success",
			idParam:    friendID.String(),
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid UUID",
			idParam:    "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "self-request",
			idParam: user.ID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "relationship already exists",
			idParam: friendID.String(),
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{}, nil
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{FindFn: tt.findFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodPost, "/users/me/friends/{id}", h.AddFriend)

			req := httptest.NewRequest(http.MethodPost, "/users/me/friends/"+tt.idParam, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestUserHandler_AcceptFriend(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	requesterID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name              string
		idParam           string
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantStatus        int
	}{
		{
			name:    "success",
			idParam: requesterID.String(),
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipPending}, nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid UUID",
			idParam:    "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "request not found",
			idParam:    requesterID.String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{FindDirectionalFn: tt.findDirectionalFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodPatch, "/users/me/friends/{id}", h.AcceptFriend)

			req := httptest.NewRequest(http.MethodPatch, "/users/me/friends/"+tt.idParam, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestUserHandler_RemoveFriend(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	otherID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name       string
		idParam    string
		findFn     func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantStatus int
	}{
		{
			name:    "success",
			idParam: otherID.String(),
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipAccepted}, nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid UUID",
			idParam:    "bad-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "friendship not found",
			idParam:    otherID.String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{FindFn: tt.findFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodDelete, "/users/me/friends/{id}", h.RemoveFriend)

			req := httptest.NewRequest(http.MethodDelete, "/users/me/friends/"+tt.idParam, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestUserHandler_BlockUser(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	targetID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name              string
		idParam           string
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantStatus        int
	}{
		{
			name:       "success",
			idParam:    targetID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid UUID",
			idParam:    "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "self-block",
			idParam: user.ID.String(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "already blocked",
			idParam: targetID.String(),
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{FindDirectionalFn: tt.findDirectionalFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodPost, "/users/me/block/{id}", h.BlockUser)

			req := httptest.NewRequest(http.MethodPost, "/users/me/block/"+tt.idParam, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestUserHandler_UnblockUser(t *testing.T) {
	user := testutil.NewTestUser()
	token := testutil.MakeTestToken(user.ID, testSecret)
	targetID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name              string
		idParam           string
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantStatus        int
	}{
		{
			name:    "success",
			idParam: targetID.String(),
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid UUID",
			idParam:    "bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "user not blocked",
			idParam:    targetID.String(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rels := &testutil.MockRelationshipRepo{FindDirectionalFn: tt.findDirectionalFn}
			h := newUserHandler(t, nil, rels)
			srv := serveProtected(http.MethodDelete, "/users/me/block/{id}", h.UnblockUser)

			req := httptest.NewRequest(http.MethodDelete, "/users/me/block/"+tt.idParam, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

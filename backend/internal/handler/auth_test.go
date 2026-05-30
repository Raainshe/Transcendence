package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/handler"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestAuthHandler_Register(t *testing.T) {
	existingUser := testutil.NewTestUser()

	tests := []struct {
		name             string
		body             string
		findByEmailFn    func(context.Context, string) (*model.User, error)
		findByUsernameFn func(context.Context, string) (*model.User, error)
		wantStatus       int
		wantKeys         []string
	}{
		{
			name:       "success",
			body:       `{"username":"alice","email":"alice@example.com","password":"secret123"}`,
			wantStatus: http.StatusCreated,
			wantKeys:   []string{"user", "token"},
		},
		{
			name:       "invalid JSON",
			body:       `{bad json`,
			wantStatus: http.StatusBadRequest,
			wantKeys:   []string{"error"},
		},
		{
			name:       "missing required fields",
			body:       `{"email":"alice@example.com"}`,
			wantStatus: http.StatusBadRequest,
			wantKeys:   []string{"error"},
		},
		{
			name: "email already taken",
			body: `{"username":"alice","email":"taken@example.com","password":"secret123"}`,
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return existingUser, nil
			},
			wantStatus: http.StatusConflict,
			wantKeys:   []string{"error"},
		},
		{
			name: "username already taken",
			body: `{"username":"taken","email":"new@example.com","password":"secret123"}`,
			findByUsernameFn: func(_ context.Context, _ string) (*model.User, error) {
				return existingUser, nil
			},
			wantStatus: http.StatusConflict,
			wantKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{
				FindByEmailFn:    tt.findByEmailFn,
				FindByUsernameFn: tt.findByUsernameFn,
			}
			svc := service.NewAuthService(repo, "test-secret")
			h := handler.NewAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	password := "correct"
	hash := testutil.HashPassword(password)
	user := testutil.NewTestUser()
	user.PasswordHash = &hash

	tests := []struct {
		name          string
		body          string
		findByEmailFn func(context.Context, string) (*model.User, error)
		wantStatus    int
		wantKeys      []string
	}{
		{
			name: "success",
			body: `{"email":"test@example.com","password":"correct"}`,
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return user, nil
			},
			wantStatus: http.StatusOK,
			wantKeys:   []string{"user", "token"},
		},
		{
			name:       "invalid JSON",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
			wantKeys:   []string{"error"},
		},
		{
			name:       "invalid credentials",
			body:       `{"email":"nobody@example.com","password":"wrong"}`,
			wantStatus: http.StatusUnauthorized,
			wantKeys:   []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByEmailFn: tt.findByEmailFn}
			svc := service.NewAuthService(repo, "test-secret")
			h := handler.NewAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.Login(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestAuthHandler_Refresh(t *testing.T) {
	const secret = "test-secret"
	validToken := testutil.MakeTestToken(testutil.NewTestUser().ID, secret)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantKeys   []string
	}{
		{
			name:       "no authorization header",
			wantStatus: http.StatusUnauthorized,
			wantKeys:   []string{"error"},
		},
		{
			name:       "invalid token",
			authHeader: "Bearer garbage.token.value",
			wantStatus: http.StatusUnauthorized,
			wantKeys:   []string{"error"},
		},
		{
			name:       "valid token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			wantKeys:   []string{"token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewAuthService(&testutil.MockUserRepo{}, secret)
			h := handler.NewAuthHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			h.Refresh(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			assertJSONKeys(t, w.Body.Bytes(), tt.wantKeys)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	svc := service.NewAuthService(&testutil.MockUserRepo{}, "test-secret")
	h := handler.NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// assertJSONKeys checks that all given keys are present in the JSON body.
func assertJSONKeys(t *testing.T, body []byte, keys []string) {
	t.Helper()
	if len(keys) == 0 {
		return
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		t.Errorf("response is not valid JSON: %v (body: %s)", err, body)
		return
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("expected key %q in response, body: %s", k, body)
		}
	}
}

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestRateLimit(t *testing.T) {
	t.Run("general limit & user isolation", func(t *testing.T) {
		userA := uuid.New()
		userB := uuid.New()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		limited := middleware.RateLimit(1, 2)(next)

		s := func(uid uuid.UUID) *service.APIKeyService {
			repo := &testutil.MockAPIKeyRepo{
				FindByHashFn: func(ctx context.Context, keyHash string) (*model.APIKey, error) {
					return &model.APIKey{UserID: uid, RevokedAt: nil}, nil
				},
			}
			return service.NewAPIKeyService(repo)
		}
		handlerA := middleware.APIKeyAuth(s(userA))(limited)
		handlerB := middleware.APIKeyAuth(s(userB))(limited)

		req := func(h http.Handler) int {
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-API-Key", "tra_anything")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			return w.Code
		}

		if c := req(handlerA); c != 200 {
			t.Errorf("request 1 for User A: got %d, want 200", c)
		}
		if c := req(handlerA); c != 200 {
			t.Errorf("request 2 for User A: got %d, want 200", c)
		}
		if c := req(handlerA); c != 429 {
			t.Errorf("request 3 for User A: got %d, want 429", c)
		}
		if c := req(handlerB); c != 200 {
			t.Errorf("request 1 for User B: got %d, want 200", c)
		}
	})
}

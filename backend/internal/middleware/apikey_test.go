package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestAPIKeyAuth(t *testing.T) {
	userID := uuid.New()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := middleware.UserIDFromContext(r.Context())
		if gotID != userID {
			t.Errorf("context userID = %v, want %v", gotID, userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		apiHeader    string
		findByHashFn func(ctx context.Context, keyHash string) (*model.APIKey, error)
		wantStatus   int
	}{
		{
			name:       "no x-api-key header",
			apiHeader:  "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "no prefix",
			apiHeader:  "357537f53",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unknown api-key",
			apiHeader:  "tra_3634745748ffh35",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "revoked api-key",
			apiHeader: "tra_revokedkey",
			findByHashFn: func(ctx context.Context, keyHash string) (*model.APIKey, error) {
				now := time.Now()
				return &model.APIKey{UserID: userID, RevokedAt: &now}, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "valid api key",
			apiHeader: "tra_validkey",
			findByHashFn: func(ctx context.Context, keyHash string) (*model.APIKey, error) {
				return &model.APIKey{UserID: userID, RevokedAt: nil}, nil
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockAPIKeyRepo{FindByHashFn: tt.findByHashFn}
			svc := service.NewAPIKeyService(repo)
			handler := middleware.APIKeyAuth(svc)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.apiHeader != "" {
				req.Header.Set("X-API-Key", tt.apiHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

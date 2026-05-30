package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/testutil"
)

func TestJWTAuth(t *testing.T) {
	const secret = "test-secret"
	userID := uuid.New()

	// next is a simple handler that writes 200 when reached.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the userID was placed in context correctly.
		gotID := middleware.UserIDFromContext(r.Context())
		if gotID != userID {
			t.Errorf("context userID = %v, want %v", gotID, userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.JWTAuth(secret, nil)(next)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "no authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-bearer scheme",
			authHeader: "Basic dXNlcjpwYXNz",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "garbage token",
			authHeader: "Bearer not.a.token",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong signing secret",
			authHeader: "Bearer " + testutil.MakeTestToken(userID, "wrong-secret"),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + testutil.MakeExpiredToken(userID, secret),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token",
			authHeader: "Bearer " + testutil.MakeTestToken(userID, secret),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

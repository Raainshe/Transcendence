package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"backend/internal/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

// JWTAuth validates Bearer tokens and (optionally) reports the authenticated
// user id to the onSeen callback for last-seen tracking. The callback runs in
// its own goroutine — it must not block the request and failures are ignored.
func JWTAuth(secret string, onSeen func(uuid.UUID)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			userID, err := auth.ParseUserID(secret, tokenStr)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			if onSeen != nil {
				go onSeen(userID)
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(userIDKey).(uuid.UUID)
	return id
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

type claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			var c claims
			token, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, c.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(userIDKey).(uuid.UUID)
	return id
}

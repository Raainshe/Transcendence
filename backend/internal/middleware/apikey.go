package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/service"
)

// APIKeyAuth validates X-API-Key headers via the API-key service and stashes
// the resolved user id in context. Public routes use this
// as their auth so handlers can read the caller via UserIDFromContext.
func APIKeyAuth(keys *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiHeader := r.Header.Get("X-API-Key")
			if !strings.HasPrefix(apiHeader, "tra_") {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid X-API-Key header")
				return
			}

			userID, err := keys.Authenticate(r.Context(), apiHeader)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid or revoked API key")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package middleware

import "net/http"

// MaxBodyBytes caps the size of incoming request bodies.
// Routes that need a larger budget (e.g. file uploads) must override r.Body
// with their own http.MaxBytesReader before reading.
func MaxBodyBytes(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, limit)
			}
			next.ServeHTTP(w, r)
		})
	}
}

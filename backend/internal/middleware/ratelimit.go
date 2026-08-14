package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	type client struct {
		limiter    *rate.Limiter
		lastSeenAt time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeenAt) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromContext(r.Context())
			if userID == uuid.Nil {
				writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			key := userID.String()

			mu.Lock()

			if _, ok := clients[key]; !ok {
				clients[key] = &client{
					limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			}
			clients[key].lastSeenAt = time.Now()
			allowed := clients[key].limiter.Allow()
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "1")
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return

			}
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitIP(rps float64, burst int) func(http.Handler) http.Handler {
	type client struct {
		limiter    *rate.Limiter
		lastSeenAt time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeenAt) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				key = host
			}

			mu.Lock()

			if _, ok := clients[key]; !ok {
				clients[key] = &client{
					limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			}
			clients[key].lastSeenAt = time.Now()
			allowed := clients[key].limiter.Allow()
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "1")
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return

			}
			next.ServeHTTP(w, r)
		})
	}
}

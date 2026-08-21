package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

type rateLimitClient struct {
	limiter    *rate.Limiter
	lastSeenAt time.Time
}

type limiterStore struct {
	mu      sync.Mutex
	clients map[string]*rateLimitClient
	rps     float64
	burst   int
}

func newLimiterStore(rps float64, burst int) *limiterStore {
	rl := &limiterStore{clients: make(map[string]*rateLimitClient), rps: rps, burst: burst}
	go rl.cleanupLoop()
	return rl
}

func retryAfterSeconds(rps float64) string {
	if rps <= 0 {
		return "1"
	}
	secs := int(math.Ceil(1 / rps))
	if secs < 1 {
		secs = 1
	}
	return strconv.Itoa(secs)
}

func (rl *limiterStore) cleanupLoop() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for key, c := range rl.clients {
			if time.Since(c.lastSeenAt) > 5*time.Minute {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *limiterStore) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, ok := rl.clients[key]
	if !ok {
		c = &rateLimitClient{limiter: rate.NewLimiter(rate.Limit(rl.rps), rl.burst)}
		rl.clients[key] = c
	}
	c.lastSeenAt = time.Now()
	return c.limiter.Allow()
}

func (rl *limiterStore) middleware(keyFunc func(*http.Request) (string, bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, ok := keyFunc(r)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if !rl.allow(key) {
				w.Header().Set("Retry-After", retryAfterSeconds(rl.rps))
				writeJSONError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	rl := newLimiterStore(rps, burst)
	return rl.middleware(func(r *http.Request) (string, bool) {
		userID := UserIDFromContext(r.Context())
		if userID == uuid.Nil {
			return "", false
		}
		return userID.String(), true
	})
}

// if I add proxy remember to use X-Forwarded-For
func RateLimitIP(rps float64, burst int) func(http.Handler) http.Handler {
	rl := newLimiterStore(rps, burst)
	return rl.middleware(func(r *http.Request) (string, bool) {
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			return host, true
		}
		return r.RemoteAddr, true
	})
}

package ws

import (
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

const playerStateMinInterval = 100 * time.Millisecond

var (
	ErrInvalidBoard       = errors.New("invalid board encoding")
	ErrRateLimited        = errors.New("rate limited")
	ErrInvalidPlayerState = errors.New("invalid player state")
)

// PlayerStateUpload is the client-sent payload for player.state messages.
type PlayerStateUpload struct {
	Score int    `json:"score"`
	Lines int    `json:"lines"`
	Level int    `json:"level"`
	Alive bool   `json:"alive"`
	Board string `json:"board"`
}

// PlayerStateBroadcast is the server-sent payload for player.state envelopes.
type PlayerStateBroadcast struct {
	UserID string `json:"user_id"`
	Score  int    `json:"score"`
	Lines  int    `json:"lines"`
	Level  int    `json:"level"`
	Alive  bool   `json:"alive"`
	Board  string `json:"board"`
}

// CachedPlayerState stores the last validated state for a player in a match.
type CachedPlayerState struct {
	UserID uuid.UUID
	Score  int
	Lines  int
	Level  int
	Alive  bool
	Board  string
}

func ValidatePlayerStateUpload(upload PlayerStateUpload) error {
	if upload.Score < 0 || upload.Score > MaxPlayerStateScore {
		return ErrInvalidPlayerState
	}
	if upload.Lines < 0 || upload.Lines > MaxPlayerStateScore {
		return ErrInvalidPlayerState
	}
	if upload.Level < 1 || upload.Level > 999 {
		return ErrInvalidPlayerState
	}
	if upload.Board == "" {
		return ErrInvalidBoard
	}
	raw, err := base64.StdEncoding.DecodeString(upload.Board)
	if err != nil {
		return ErrInvalidBoard
	}
	if len(raw) != MatrixCellCount {
		return ErrInvalidBoard
	}
	for _, b := range raw {
		if b > 7 {
			return ErrInvalidBoard
		}
	}
	return nil
}

type MatchStateStore struct {
	mu     sync.RWMutex
	states map[uuid.UUID]map[uuid.UUID]CachedPlayerState
}

func NewMatchStateStore() *MatchStateStore {
	return &MatchStateStore{
		states: make(map[uuid.UUID]map[uuid.UUID]CachedPlayerState),
	}
}

func (s *MatchStateStore) Set(gameID, userID uuid.UUID, state CachedPlayerState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.states[gameID] == nil {
		s.states[gameID] = make(map[uuid.UUID]CachedPlayerState)
	}
	s.states[gameID][userID] = state
}

func (s *MatchStateStore) List(gameID uuid.UUID) []CachedPlayerState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	players := s.states[gameID]
	if len(players) == 0 {
		return nil
	}
	out := make([]CachedPlayerState, 0, len(players))
	for _, st := range players {
		out = append(out, st)
	}
	return out
}

func (s *MatchStateStore) Evict(gameID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, gameID)
}

type rateLimitKey struct {
	gameID uuid.UUID
	userID uuid.UUID
}

type RateLimiter struct {
	mu       sync.Mutex
	lastSeen map[rateLimitKey]time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		lastSeen: make(map[rateLimitKey]time.Time),
	}
}

func (r *RateLimiter) Allow(gameID, userID uuid.UUID) bool {
	key := rateLimitKey{gameID: gameID, userID: userID}
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	if last, ok := r.lastSeen[key]; ok && now.Sub(last) < playerStateMinInterval {
		return false
	}
	r.lastSeen[key] = now
	return true
}

func (r *RateLimiter) EvictMatch(gameID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for key := range r.lastSeen {
		if key.gameID == gameID {
			delete(r.lastSeen, key)
		}
	}
}

package ws

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// DisconnectGracePeriod is how long a disconnected player may reconnect before forfeit.
var DisconnectGracePeriod = 30 * time.Second

type disconnectEntry struct {
	timer *time.Timer
}

// DisconnectTracker tracks WS disconnect grace timers per match player.
type DisconnectTracker struct {
	mu        sync.Mutex
	pending   map[uuid.UUID]map[uuid.UUID]*disconnectEntry
	onTimeout func(gameID, userID uuid.UUID)
}

func NewDisconnectTracker(onTimeout func(gameID, userID uuid.UUID)) *DisconnectTracker {
	return &DisconnectTracker{
		pending:   make(map[uuid.UUID]map[uuid.UUID]*disconnectEntry),
		onTimeout: onTimeout,
	}
}

// MarkDisconnected starts a grace timer. Returns true when newly marked.
func (d *DisconnectTracker) MarkDisconnected(gameID, userID uuid.UUID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	byGame := d.pending[gameID]
	if byGame == nil {
		byGame = make(map[uuid.UUID]*disconnectEntry)
		d.pending[gameID] = byGame
	}
	if _, exists := byGame[userID]; exists {
		return false
	}

	gameIDCopy := gameID
	userIDCopy := userID
	entry := &disconnectEntry{
		timer: time.AfterFunc(DisconnectGracePeriod, func() {
			d.mu.Lock()
			byGame := d.pending[gameIDCopy]
			if byGame != nil {
				delete(byGame, userIDCopy)
				if len(byGame) == 0 {
					delete(d.pending, gameIDCopy)
				}
			}
			d.mu.Unlock()
			if d.onTimeout != nil {
				d.onTimeout(gameIDCopy, userIDCopy)
			}
		}),
	}
	byGame[userID] = entry
	return true
}

// MarkReconnected cancels a pending timer. Returns true when the player was disconnected.
func (d *DisconnectTracker) MarkReconnected(gameID, userID uuid.UUID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	byGame := d.pending[gameID]
	if byGame == nil {
		return false
	}
	entry, ok := byGame[userID]
	if !ok {
		return false
	}
	if entry.timer != nil {
		entry.timer.Stop()
	}
	delete(byGame, userID)
	if len(byGame) == 0 {
		delete(d.pending, gameID)
	}
	return true
}

// IsDisconnected reports whether a player is in the grace window.
func (d *DisconnectTracker) IsDisconnected(gameID, userID uuid.UUID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	byGame := d.pending[gameID]
	if byGame == nil {
		return false
	}
	_, ok := byGame[userID]
	return ok
}

// ListDisconnected returns user IDs currently in the grace window for a match.
func (d *DisconnectTracker) ListDisconnected(gameID uuid.UUID) []uuid.UUID {
	d.mu.Lock()
	defer d.mu.Unlock()
	byGame := d.pending[gameID]
	if len(byGame) == 0 {
		return nil
	}
	out := make([]uuid.UUID, 0, len(byGame))
	for userID := range byGame {
		out = append(out, userID)
	}
	return out
}

// ClearUser removes tracking for one player (e.g. after elimination).
func (d *DisconnectTracker) ClearUser(gameID, userID uuid.UUID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	byGame := d.pending[gameID]
	if byGame == nil {
		return
	}
	if entry, ok := byGame[userID]; ok && entry.timer != nil {
		entry.timer.Stop()
	}
	delete(byGame, userID)
	if len(byGame) == 0 {
		delete(d.pending, gameID)
	}
}

// Evict cancels all timers for a finished match.
func (d *DisconnectTracker) Evict(gameID uuid.UUID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	byGame := d.pending[gameID]
	for _, entry := range byGame {
		if entry.timer != nil {
			entry.timer.Stop()
		}
	}
	delete(d.pending, gameID)
}

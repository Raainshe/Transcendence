package ws

import (
	"sync"

	"github.com/google/uuid"
)

type eliminationEntry struct {
	UserID    uuid.UUID
	Reason    string
	Placement int
}

type matchLifecycle struct {
	playerIDs       []uuid.UUID
	eliminatedOrder []eliminationEntry
	finished        bool
}

// MatchLifecycle tracks eliminations and placements per in-progress match.
type MatchLifecycle struct {
	mu      sync.RWMutex
	matches map[uuid.UUID]*matchLifecycle
}

func NewMatchLifecycle() *MatchLifecycle {
	return &MatchLifecycle{
		matches: make(map[uuid.UUID]*matchLifecycle),
	}
}

func (l *MatchLifecycle) InitMatch(gameID uuid.UUID, playerIDs []uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.matches[gameID]; ok {
		return
	}
	ids := make([]uuid.UUID, len(playerIDs))
	copy(ids, playerIDs)
	l.matches[gameID] = &matchLifecycle{
		playerIDs: ids,
	}
}

func (l *MatchLifecycle) RegisterElimination(
	gameID, userID uuid.UUID,
	reason string,
) (placement int, aliveRemaining int, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	m, exists := l.matches[gameID]
	if !exists || m.finished {
		return 0, 0, false
	}

	for _, e := range m.eliminatedOrder {
		if e.UserID == userID {
			return e.Placement, len(m.playerIDs) - len(m.eliminatedOrder), false
		}
	}

	aliveBefore := len(m.playerIDs) - len(m.eliminatedOrder)
	aliveRemaining = aliveBefore - 1
	placement = aliveRemaining + 1

	m.eliminatedOrder = append(m.eliminatedOrder, eliminationEntry{
		UserID:    userID,
		Reason:    reason,
		Placement: placement,
	})
	return placement, aliveRemaining, true
}

func (l *MatchLifecycle) IsEliminated(gameID, userID uuid.UUID) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return false
	}
	for _, e := range m.eliminatedOrder {
		if e.UserID == userID {
			return true
		}
	}
	return false
}

func (l *MatchLifecycle) AliveRemaining(gameID uuid.UUID) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return 0
	}
	return len(m.playerIDs) - len(m.eliminatedOrder)
}

func (l *MatchLifecycle) Survivor(gameID uuid.UUID) (uuid.UUID, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return uuid.Nil, false
	}
	if len(m.playerIDs)-len(m.eliminatedOrder) != 1 {
		return uuid.Nil, false
	}
	eliminated := make(map[uuid.UUID]struct{}, len(m.eliminatedOrder))
	for _, e := range m.eliminatedOrder {
		eliminated[e.UserID] = struct{}{}
	}
	for _, id := range m.playerIDs {
		if _, out := eliminated[id]; !out {
			return id, true
		}
	}
	return uuid.Nil, false
}

func (l *MatchLifecycle) PlayerIDs(gameID uuid.UUID) []uuid.UUID {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return nil
	}
	out := make([]uuid.UUID, len(m.playerIDs))
	copy(out, m.playerIDs)
	return out
}

func (l *MatchLifecycle) EliminatedEntries(gameID uuid.UUID) []eliminationEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return nil
	}
	out := make([]eliminationEntry, len(m.eliminatedOrder))
	copy(out, m.eliminatedOrder)
	return out
}

func (l *MatchLifecycle) EliminationPlacement(gameID, userID uuid.UUID) int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return 0
	}
	for _, e := range m.eliminatedOrder {
		if e.UserID == userID {
			return e.Placement
		}
	}
	return 0
}

func (l *MatchLifecycle) EliminationReason(gameID, userID uuid.UUID) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	if !ok {
		return ""
	}
	for _, e := range m.eliminatedOrder {
		if e.UserID == userID {
			return e.Reason
		}
	}
	return ""
}

func (l *MatchLifecycle) MarkFinished(gameID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if m, ok := l.matches[gameID]; ok {
		m.finished = true
	}
}

func (l *MatchLifecycle) IsFinished(gameID uuid.UUID) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	m, ok := l.matches[gameID]
	return ok && m.finished
}

func (l *MatchLifecycle) Evict(gameID uuid.UUID) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.matches, gameID)
}

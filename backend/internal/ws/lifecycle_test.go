package ws_test

import (
	"testing"

	"github.com/google/uuid"

	"backend/internal/ws"
)

func TestMatchLifecycle_RegisterElimination_Placements(t *testing.T) {
	life := ws.NewMatchLifecycle()
	gameID := uuid.New()
	a := uuid.New()
	b := uuid.New()
	c := uuid.New()
	life.InitMatch(gameID, []uuid.UUID{a, b, c})

	p1, alive, ok := life.RegisterElimination(gameID, a, "topOut")
	if !ok || p1 != 3 || alive != 2 {
		t.Fatalf("first elimination = (%d, %d, %v), want (3, 2, true)", p1, alive, ok)
	}

	p2, alive, ok := life.RegisterElimination(gameID, b, "lockOut")
	if !ok || p2 != 2 || alive != 1 {
		t.Fatalf("second elimination = (%d, %d, %v), want (2, 1, true)", p2, alive, ok)
	}

	survivor, ok := life.Survivor(gameID)
	if !ok || survivor != c {
		t.Fatalf("survivor = %v, %v, want %v true", survivor, ok, c)
	}

	_, _, ok = life.RegisterElimination(gameID, a, "topOut")
	if ok {
		t.Fatal("duplicate elimination should be idempotent")
	}
}

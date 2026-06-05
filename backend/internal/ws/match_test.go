package ws_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/ws"
)

func validBoardB64() string {
	return base64.StdEncoding.EncodeToString(make([]byte, ws.MatrixCellCount))
}

func TestValidatePlayerStateUpload_Valid(t *testing.T) {
	err := ws.ValidatePlayerStateUpload(ws.PlayerStateUpload{
		Score: 1200,
		Lines: 10,
		Level: 3,
		Alive: true,
		Board: validBoardB64(),
	})
	if err != nil {
		t.Fatalf("ValidatePlayerStateUpload: %v", err)
	}
}

func TestValidatePlayerStateUpload_InvalidBoardLength(t *testing.T) {
	short := base64.StdEncoding.EncodeToString(make([]byte, 10))
	err := ws.ValidatePlayerStateUpload(ws.PlayerStateUpload{
		Score: 0,
		Lines: 0,
		Level: 1,
		Alive: true,
		Board: short,
	})
	if err != ws.ErrInvalidBoard {
		t.Fatalf("err = %v, want ErrInvalidBoard", err)
	}
}

func TestValidatePlayerStateUpload_InvalidCellValue(t *testing.T) {
	raw := make([]byte, ws.MatrixCellCount)
	raw[0] = 8
	board := base64.StdEncoding.EncodeToString(raw)
	err := ws.ValidatePlayerStateUpload(ws.PlayerStateUpload{
		Score: 0,
		Lines: 0,
		Level: 1,
		Alive: true,
		Board: board,
	})
	if err != ws.ErrInvalidBoard {
		t.Fatalf("err = %v, want ErrInvalidBoard", err)
	}
}

func TestValidatePlayerStateUpload_InvalidScore(t *testing.T) {
	err := ws.ValidatePlayerStateUpload(ws.PlayerStateUpload{
		Score: -1,
		Lines: 0,
		Level: 1,
		Alive: true,
		Board: validBoardB64(),
	})
	if err != ws.ErrInvalidPlayerState {
		t.Fatalf("err = %v, want ErrInvalidPlayerState", err)
	}
}

func TestRateLimiter_AllowsAfterInterval(t *testing.T) {
	rl := ws.NewRateLimiter()
	gameID := uuid.New()
	userID := uuid.New()

	if !rl.Allow(gameID, userID) {
		t.Fatal("first update should be allowed")
	}
	if rl.Allow(gameID, userID) {
		t.Fatal("immediate second update should be rate limited")
	}

	time.Sleep(110 * time.Millisecond)
	if !rl.Allow(gameID, userID) {
		t.Fatal("update after interval should be allowed")
	}
}

func TestMatchStateStore_ListAndEvict(t *testing.T) {
	store := ws.NewMatchStateStore()
	gameID := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	store.Set(gameID, userA, ws.CachedPlayerState{UserID: userA, Score: 1})
	store.Set(gameID, userB, ws.CachedPlayerState{UserID: userB, Score: 2})

	states := store.List(gameID)
	if len(states) != 2 {
		t.Fatalf("states = %d, want 2", len(states))
	}

	store.Evict(gameID)
	if len(store.List(gameID)) != 0 {
		t.Fatal("expected empty list after evict")
	}
}

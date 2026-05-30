//go:build integration

package itest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func recordGame(t *testing.T, token, mode string, score int, win bool) {
	t.Helper()
	now := time.Now().UTC()
	payload, err := json.Marshal(map[string]any{
		"mode":          mode,
		"score":         score,
		"lines_cleared": 40,
		"level_reached": 5,
		"started_at":    now.Add(-5 * time.Minute).Format(time.RFC3339),
		"finished_at":   now.Format(time.RFC3339),
		"is_winner":     win,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/games", token, string(payload))
	mustStatus(t, resp, raw, http.StatusCreated)
}

func TestGames_RecordAndStats(t *testing.T) {
	truncate(t)
	userID, token := registerUser(t, "alice", "alice@example.com", "secret12")

	recordGame(t, token, "marathon", 1000, true)

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/"+userID.String()+"/stats", "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	var stats struct {
		GamesPlayed int `json:"games_played"`
		Wins        int `json:"wins"`
		BestScore   int `json:"best_score"`
	}
	decodeJSON(t, raw, &stats)
	if stats.GamesPlayed != 1 || stats.Wins != 1 || stats.BestScore != 1000 {
		t.Errorf("stats = %+v, want games=1 wins=1 best=1000", stats)
	}
}

func TestGames_InvalidMode(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	now := time.Now().UTC().Format(time.RFC3339)
	body := fmt.Sprintf(`{"mode":"nonsense","score":1,"lines_cleared":0,"level_reached":1,"started_at":%q,"finished_at":%q}`, now, now)
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/games", token, body)
	mustStatus(t, resp, raw, http.StatusBadRequest)
}

func TestLeaderboard_OrderedByScore(t *testing.T) {
	truncate(t)
	_, aTok := registerUser(t, "alice", "alice@example.com", "secret12")
	_, bTok := registerUser(t, "bob", "bob@example.com", "secret12")
	_, cTok := registerUser(t, "carol", "carol@example.com", "secret12")

	recordGame(t, aTok, "marathon", 500, false)
	recordGame(t, bTok, "marathon", 1500, true)
	recordGame(t, cTok, "marathon", 1000, false)

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/leaderboard", "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	var entries []struct {
		Rank     int64  `json:"rank"`
		Username string `json:"username"`
		Score    int    `json:"score"`
	}
	decodeJSON(t, raw, &entries)
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3 (body: %s)", len(entries), raw)
	}
	wantOrder := []string{"bob", "carol", "alice"}
	for i, want := range wantOrder {
		if entries[i].Username != want {
			t.Errorf("rank %d = %s, want %s", i+1, entries[i].Username, want)
		}
	}
}

func TestGames_ListAndDetail(t *testing.T) {
	truncate(t)
	aliceID, aliceTok := registerUser(t, "alice", "alice@example.com", "secret12")
	bobID, bobTok := registerUser(t, "bob", "bob@example.com", "secret12")

	recordGame(t, aliceTok, "marathon", 100, true)
	recordGame(t, aliceTok, "sprint", 200, false)
	recordGame(t, bobTok, "marathon", 300, true)

	// List all → 3 games, total 3
	resp, raw := doJSON(t, http.MethodGet, "/api/v1/games", "", "")
	mustStatus(t, resp, raw, http.StatusOK)
	var all struct {
		Games []struct{ ID string `json:"id"` } `json:"games"`
		Total int                                `json:"total"`
	}
	decodeJSON(t, raw, &all)
	if len(all.Games) != 3 || all.Total != 3 {
		t.Fatalf("list all: got %d games (total %d), want 3/3 (body: %s)", len(all.Games), all.Total, raw)
	}

	// Filter by Alice → 2 games
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/games?user_id="+aliceID.String(), "", "")
	mustStatus(t, resp, raw, http.StatusOK)
	var alice struct {
		Games []struct{ ID string `json:"id"` } `json:"games"`
		Total int                                `json:"total"`
	}
	decodeJSON(t, raw, &alice)
	if len(alice.Games) != 2 || alice.Total != 2 {
		t.Fatalf("list by alice: got %d games (total %d), want 2/2", len(alice.Games), alice.Total)
	}
	_ = bobID

	// Fetch detail of the first game; player record should be present
	first := alice.Games[0].ID
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/games/"+first, "", "")
	mustStatus(t, resp, raw, http.StatusOK)
	var detail struct {
		Game struct {
			ID      string `json:"id"`
			Players []struct {
				UserID string `json:"user_id"`
			} `json:"players"`
		} `json:"game"`
	}
	decodeJSON(t, raw, &detail)
	if detail.Game.ID != first || len(detail.Game.Players) != 1 || detail.Game.Players[0].UserID != aliceID.String() {
		t.Errorf("detail = %+v, want id=%s with 1 alice player", detail.Game, first)
	}

	// Unknown id → 404
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/games/"+uuid.NewString(), "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

func TestLeaderboard_LimitClamping(t *testing.T) {
	truncate(t)
	_, tok := registerUser(t, "alice", "alice@example.com", "secret12")
	for i := 0; i < 5; i++ {
		recordGame(t, tok, "marathon", 100+i, false)
	}

	// limit=999 should be clamped to 100 by the service; with 5 rows we get 5 back.
	resp, raw := doJSON(t, http.MethodGet, "/api/v1/leaderboard?limit=999", "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	var entries []map[string]any
	decodeJSON(t, raw, &entries)
	if len(entries) != 5 {
		t.Errorf("len(entries) = %d, want 5", len(entries))
	}
}

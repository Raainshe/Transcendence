//go:build integration

package itest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// doJSON's counterpart for the public API, authenticating
// via X-API-Key instead of a JWT Bearer token.
func doPublicJSON(t *testing.T, method, path, apiKey, body string) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, raw
}

// generates a real API key for the given user via the internal
// endpoint and returns the raw key text.
func createAPIKey(t *testing.T, token, name string) string {
	t.Helper()
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/users/me/api-keys", token, fmt.Sprintf(`{"name":%q}`, name))
	mustStatus(t, resp, raw, http.StatusCreated)
	var out struct {
		APIKey struct {
			Key string `json:"key"`
		} `json:"api_key"`
	}
	decodeJSON(t, raw, &out)
	if out.APIKey.Key == "" {
		t.Fatalf("create API key: empty key in response %s", raw)
	}
	return out.APIKey.Key
}

func TestPublicAPI_RequiresAPIKey(t *testing.T) {
	truncate(t)
	resp, raw := doPublicJSON(t, http.MethodGet, "/api/public/v1/leaderboard", "", "")
	mustStatus(t, resp, raw, http.StatusUnauthorized)
}

func TestPublicAPI_LeaderboardAndStats(t *testing.T) {
	truncate(t)
	userID, token := registerUser(t, "alice", "alice@example.com", "secret12")
	recordGame(t, token, "multiplayer", 1000, true)
	apiKey := createAPIKey(t, token, "test integration")

	resp, raw := doPublicJSON(t, http.MethodGet, "/api/public/v1/leaderboard", apiKey, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var lb struct {
		Leaderboard []struct {
			UserID uuid.UUID `json:"user_id"`
			Score  int       `json:"score"`
		} `json:"leaderboard"`
	}
	decodeJSON(t, raw, &lb)
	if len(lb.Leaderboard) != 1 || lb.Leaderboard[0].UserID != userID || lb.Leaderboard[0].Score != 1000 {
		t.Errorf("leaderboard = %+v, want one entry for %s with score 1000", lb.Leaderboard, userID)
	}

	resp, raw = doPublicJSON(t, http.MethodGet, "/api/public/v1/users/"+userID.String()+"/stats", apiKey, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var stats struct {
		Stats struct {
			GamesPlayed int `json:"games_played"`
			Wins        int `json:"wins"`
			BestScore   int `json:"best_score"`
		} `json:"stats"`
	}
	decodeJSON(t, raw, &stats)
	if stats.Stats.GamesPlayed != 1 || stats.Stats.Wins != 1 || stats.Stats.BestScore != 1000 {
		t.Errorf("stats = %+v, want games=1 wins=1 best=1000", stats.Stats)
	}
}

func TestPublicAPI_LobbyLifecycle(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "host", "host@example.com", "secret12")
	apiKey := createAPIKey(t, token, "test integration")

	resp, raw := doPublicJSON(t, http.MethodPost, "/api/public/v1/lobbies", apiKey, `{"max_players":4}`)
	mustStatus(t, resp, raw, http.StatusCreated)
	var created struct {
		Lobby struct {
			ID uuid.UUID `json:"id"`
		} `json:"lobby"`
	}
	decodeJSON(t, raw, &created)
	if created.Lobby.ID == uuid.Nil {
		t.Fatalf("create response missing lobby id: %s", raw)
	}
	lobbyID := created.Lobby.ID.String()

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/lobbies/"+lobbyID, token, "")
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doPublicJSON(t, http.MethodPut, "/api/public/v1/lobbies/"+lobbyID, apiKey, `{"name":"Fun"}`)
	mustStatus(t, resp, raw, http.StatusOK)
	var renamed struct {
		Lobby struct {
			Name string `json:"name"`
		} `json:"lobby"`
	}
	decodeJSON(t, raw, &renamed)
	if renamed.Lobby.Name != "Fun" {
		t.Errorf("lobby name = %q, want %q", renamed.Lobby.Name, "Fun")
	}

	resp, raw = doPublicJSON(t, http.MethodDelete, "/api/public/v1/lobbies/"+lobbyID, apiKey, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/lobbies/"+lobbyID, token, "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

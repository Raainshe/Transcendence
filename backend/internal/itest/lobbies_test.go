//go:build integration

package itest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestLobbies_FullFlow(t *testing.T) {
	truncate(t)

	hostID, hostToken := registerUser(t, "host", "host@example.com", "secret12")
	guestID, guestToken := registerUser(t, "guest", "guest@example.com", "secret12")
	_ = hostID
	_ = guestID

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/lobbies", hostToken, `{"max_players":4}`)
	mustStatus(t, resp, raw, http.StatusCreated)

	var created struct {
		Lobby struct {
			ID         uuid.UUID `json:"id"`
			InviteCode string    `json:"invite_code"`
		} `json:"lobby"`
		InviteCode string `json:"invite_code"`
	}
	decodeJSON(t, raw, &created)
	lobbyID := created.Lobby.ID
	code := created.InviteCode
	if code == "" {
		code = created.Lobby.InviteCode
	}
	if lobbyID == uuid.Nil || code == "" {
		t.Fatalf("create response missing lobby id or invite code: %s", raw)
	}

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/join", guestToken, fmt.Sprintf(`{"invite_code":%q}`, code))
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+lobbyID.String()+"/ready", hostToken, `{"ready":true}`)
	mustStatus(t, resp, raw, http.StatusOK)
	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+lobbyID.String()+"/ready", guestToken, `{"ready":true}`)
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+lobbyID.String()+"/start", hostToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	var start struct {
		GameID     uuid.UUID `json:"game_id"`
		SharedSeed int64     `json:"shared_seed"`
		Players    []struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"players"`
	}
	decodeJSON(t, raw, &start)
	if start.GameID == uuid.Nil || start.SharedSeed == 0 || len(start.Players) != 2 {
		t.Fatalf("start response = %+v, want game_id, shared_seed, 2 players", start)
	}

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/games/"+start.GameID.String(), "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	var game struct {
		Game struct {
			Mode    string `json:"mode"`
			Status  string `json:"status"`
			Players []struct {
				UserID uuid.UUID `json:"user_id"`
				Score  int       `json:"score"`
			} `json:"players"`
		} `json:"game"`
	}
	decodeJSON(t, raw, &game)
	if game.Game.Mode != "multiplayer" || game.Game.Status != "in_progress" {
		t.Fatalf("game = %+v, want multiplayer in_progress", game.Game)
	}
	if len(game.Game.Players) != 2 {
		t.Fatalf("players = %d, want 2", len(game.Game.Players))
	}
	for _, p := range game.Game.Players {
		if p.Score != 0 {
			t.Fatalf("player score = %d, want 0 at start", p.Score)
		}
	}

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/lobbies/"+lobbyID.String(), hostToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	var lobbySnap struct {
		Lobby struct {
			Status     string     `json:"status"`
			GameID     *uuid.UUID `json:"game_id"`
			SharedSeed *int64     `json:"shared_seed"`
			Members    []struct {
				IsReady bool `json:"is_ready"`
			} `json:"members"`
		} `json:"lobby"`
	}
	decodeJSON(t, raw, &lobbySnap)
	if lobbySnap.Lobby.Status != "closed" || lobbySnap.Lobby.GameID == nil {
		t.Fatalf("lobby snapshot = %+v, want closed with game_id", lobbySnap.Lobby)
	}
}

func TestLobbies_HostLeaveClosesLobby(t *testing.T) {
	truncate(t)

	_, hostToken := registerUser(t, "host2", "host2@example.com", "secret12")
	_, guestToken := registerUser(t, "guest2", "guest2@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/lobbies", hostToken, `{"max_players":2}`)
	mustStatus(t, resp, raw, http.StatusCreated)

	var created struct {
		Lobby struct {
			ID         uuid.UUID `json:"id"`
			InviteCode string    `json:"invite_code"`
		} `json:"lobby"`
	}
	decodeJSON(t, raw, &created)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/join", guestToken,
		fmt.Sprintf(`{"invite_code":%q}`, created.Lobby.InviteCode))
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodDelete, "/api/v1/lobbies/"+created.Lobby.ID.String()+"/leave", hostToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/lobbies/"+created.Lobby.ID.String(), guestToken, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d after host leave, want 404 (body: %s)", resp.StatusCode, raw)
	}
}

func TestLobbies_StartRequiresAllReady(t *testing.T) {
	truncate(t)

	_, hostToken := registerUser(t, "host3", "host3@example.com", "secret12")
	_, guestToken := registerUser(t, "guest3", "guest3@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/lobbies", hostToken, `{"max_players":2}`)
	mustStatus(t, resp, raw, http.StatusCreated)

	var created struct {
		Lobby struct {
			ID         uuid.UUID `json:"id"`
			InviteCode string    `json:"invite_code"`
		} `json:"lobby"`
	}
	decodeJSON(t, raw, &created)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/join", guestToken,
		fmt.Sprintf(`{"invite_code":%q}`, created.Lobby.InviteCode))
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+created.Lobby.ID.String()+"/ready", hostToken, `{"ready":true}`)
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+created.Lobby.ID.String()+"/start", hostToken, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 when guest not ready (body: %s)", resp.StatusCode, raw)
	}

	var errBody struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &errBody)
	if errBody.Error == "" {
		t.Fatalf("expected error message in body: %s", raw)
	}
}

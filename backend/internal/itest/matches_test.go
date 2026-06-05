//go:build integration

package itest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func validBoardB64() string {
	return base64.StdEncoding.EncodeToString(make([]byte, 400))
}

func startMatchFromLobby(t *testing.T) (gameID uuid.UUID, hostToken, guestToken string) {
	t.Helper()
	truncate(t)

	_, hostToken = registerUser(t, "mhost", "mhost@example.com", "secret12")
	_, guestToken = registerUser(t, "mguest", "mguest@example.com", "secret12")

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
	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+created.Lobby.ID.String()+"/ready", guestToken, `{"ready":true}`)
	mustStatus(t, resp, raw, http.StatusOK)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/lobbies/"+created.Lobby.ID.String()+"/start", hostToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	var start struct {
		GameID uuid.UUID `json:"game_id"`
	}
	decodeJSON(t, raw, &start)
	if start.GameID == uuid.Nil {
		t.Fatalf("start response missing game_id: %s", raw)
	}
	return start.GameID, hostToken, guestToken
}

func dialIntegrationWS(t *testing.T, token string) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(srv.URL, "http://", "ws://", 1) + "/api/v1/ws?token=" + token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestMatches_GetMetadata(t *testing.T) {
	gameID, hostToken, guestToken := startMatchFromLobby(t)

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/matches/"+gameID.String(), hostToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	var hostView struct {
		Match *struct {
			GameID     uuid.UUID `json:"game_id"`
			Status     string    `json:"status"`
			SharedSeed int64     `json:"shared_seed"`
			Players    []struct {
				UserID uuid.UUID `json:"user_id"`
			} `json:"players"`
		} `json:"match"`
	}
	decodeJSON(t, raw, &hostView)
	if hostView.Match == nil || hostView.Match.GameID != gameID {
		t.Fatalf("host match view = %+v, want metadata", hostView.Match)
	}
	if hostView.Match.Status != "in_progress" || hostView.Match.SharedSeed == 0 {
		t.Fatalf("match = %+v, want in_progress with shared_seed", hostView.Match)
	}
	if len(hostView.Match.Players) != 2 {
		t.Fatalf("players = %d, want 2", len(hostView.Match.Players))
	}

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/matches/"+gameID.String(), guestToken, "")
	mustStatus(t, resp, raw, http.StatusOK)

	_, outsiderToken := registerUser(t, "outsider", "outsider@example.com", "secret12")
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/matches/"+gameID.String(), outsiderToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var outsiderView struct {
		Match *struct {
			GameID uuid.UUID `json:"game_id"`
		} `json:"match"`
	}
	decodeJSON(t, raw, &outsiderView)
	if outsiderView.Match != nil {
		t.Fatalf("outsider match = %+v, want null", outsiderView.Match)
	}
}

func TestMatches_PlayerStateRelay(t *testing.T) {
	gameID, hostToken, guestToken := startMatchFromLobby(t)

	hostConn := dialIntegrationWS(t, hostToken)
	guestConn := dialIntegrationWS(t, guestToken)

	subscribe := func(conn *websocket.Conn) {
		t.Helper()
		if err := conn.WriteJSON(map[string]string{
			"type": "subscribe",
			"room": "match:" + gameID.String(),
		}); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	subscribe(hostConn)
	subscribe(guestConn)

	upload := map[string]any{
		"type": "player.state",
		"room": "match:" + gameID.String(),
		"payload": map[string]any{
			"score": 900,
			"lines": 12,
			"level": 4,
			"alive": true,
			"board": validBoardB64(),
		},
	}
	if err := hostConn.WriteJSON(upload); err != nil {
		t.Fatalf("write player.state: %v", err)
	}

	_ = guestConn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := guestConn.ReadMessage()
	if err != nil {
		t.Fatalf("guest read: %v", err)
	}

	var env struct {
		Type    string `json:"type"`
		Payload struct {
			UserID string `json:"user_id"`
			Score  int    `json:"score"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(msg, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Type != "player.state" || env.Payload.Score != 900 {
		t.Fatalf("guest received %+v, want player.state score 900", env)
	}
}

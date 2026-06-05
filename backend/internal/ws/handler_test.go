package ws_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"backend/internal/model"
	"backend/internal/testutil"
	"backend/internal/ws"
)

const testJWTSecret = "test-jwt-secret-for-ws-handler"

func validBoardB64WS() string {
	return base64.StdEncoding.EncodeToString(make([]byte, ws.MatrixCellCount))
}

func startWSTestServer(t *testing.T, lobbies ws.MemberChecker, games ws.GamePlayerChecker) (string, *ws.Hub) {
	t.Helper()
	hub := ws.NewHub()
	handler := ws.NewHandler(hub, testJWTSecret, lobbies, games)
	srv := httptest.NewServer(http.HandlerFunc(handler.ServeHTTP))
	t.Cleanup(srv.Close)
	return strings.Replace(srv.URL, "http://", "ws://", 1), hub
}

func dialWS(t *testing.T, baseURL, token string) *websocket.Conn {
	t.Helper()
	conn, resp, err := websocket.DefaultDialer.Dial(baseURL+"?token="+token, nil)
	if err != nil {
		t.Fatalf("dial ws: %v (status %v)", err, resp)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func readEnvelope(t *testing.T, conn *websocket.Conn) ws.Envelope {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	var env ws.Envelope
	if err := json.Unmarshal(msg, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return env
}

func TestWSHandler_MatchSubscribeDeniedForNonMember(t *testing.T) {
	gameID := uuid.New()
	userID := uuid.New()
	token := testutil.MakeTestToken(userID, testJWTSecret)

	games := &testutil.MockGameRepo{
		IsGamePlayerFn: func(_ context.Context, gid, uid uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	url, _ := startWSTestServer(t, &testutil.MockLobbyRepo{}, games)
	conn := dialWS(t, url, token)

	if err := conn.WriteJSON(map[string]string{
		"type": "subscribe",
		"room": ws.MatchRoomID(gameID.String()),
	}); err != nil {
		t.Fatalf("write subscribe: %v", err)
	}

	env := readEnvelope(t, conn)
	if env.Type != ws.TypeError {
		t.Fatalf("type = %s, want error", env.Type)
	}
}

func TestWSHandler_PlayerStateRelayAndReplay(t *testing.T) {
	gameID := uuid.New()
	hostID := uuid.New()
	guestID := uuid.New()
	hostToken := testutil.MakeTestToken(hostID, testJWTSecret)
	guestToken := testutil.MakeTestToken(guestID, testJWTSecret)

	inProgress := &model.Game{
		ID:     gameID,
		Mode:   "multiplayer",
		Status: "in_progress",
	}

	games := &testutil.MockGameRepo{
		IsGamePlayerFn: func(_ context.Context, gid, uid uuid.UUID) (bool, error) {
			if gid != gameID {
				return false, nil
			}
			return uid == hostID || uid == guestID, nil
		},
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			if id == gameID {
				return inProgress, nil
			}
			return nil, nil
		},
	}

	url, hub := startWSTestServer(t, &testutil.MockLobbyRepo{}, games)
	hostConn := dialWS(t, url, hostToken)
	guestConn := dialWS(t, url, guestToken)

	subscribe := func(conn *websocket.Conn) {
		t.Helper()
		if err := conn.WriteJSON(map[string]string{
			"type": "subscribe",
			"room": ws.MatchRoomID(gameID.String()),
		}); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	subscribe(hostConn)
	subscribe(guestConn)

	upload := map[string]any{
		"type": ws.TypePlayerState,
		"room": ws.MatchRoomID(gameID.String()),
		"payload": map[string]any{
			"score": 500,
			"lines": 4,
			"level": 2,
			"alive": true,
			"board": validBoardB64WS(),
		},
	}
	if err := hostConn.WriteJSON(upload); err != nil {
		t.Fatalf("write player.state: %v", err)
	}

	env := readEnvelope(t, guestConn)
	if env.Type != ws.TypePlayerState {
		t.Fatalf("guest envelope type = %s, want player.state", env.Type)
	}
	var broadcast ws.PlayerStateBroadcast
	if err := json.Unmarshal(env.Payload, &broadcast); err != nil {
		t.Fatalf("unmarshal broadcast: %v", err)
	}
	if broadcast.UserID != hostID.String() || broadcast.Score != 500 {
		t.Fatalf("broadcast = %+v, want host score 500", broadcast)
	}

	// Replay: third connection subscribes after state cached.
	lateToken := testutil.MakeTestToken(guestID, testJWTSecret)
	lateConn := dialWS(t, url, lateToken)
	subscribe(lateConn)
	replay := readEnvelope(t, lateConn)
	if replay.Type != ws.TypePlayerState {
		t.Fatalf("replay type = %s, want player.state", replay.Type)
	}

	// Host should not receive their own broadcast.
	_ = hostConn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	if _, _, err := hostConn.ReadMessage(); err == nil {
		t.Fatal("host should not receive own player.state broadcast")
	}

	_ = hub
}

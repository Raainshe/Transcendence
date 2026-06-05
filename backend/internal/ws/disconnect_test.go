package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"backend/internal/model"
	"backend/internal/testutil"
	"backend/internal/ws"
)

func multiplayerGameMocks(gameID, hostID, guestID uuid.UUID) *testutil.MockGameRepo {
	inProgress := &model.Game{ID: gameID, Mode: "multiplayer", Status: "in_progress"}
	return &testutil.MockGameRepo{
		IsGamePlayerFn: func(_ context.Context, gid, uid uuid.UUID) (bool, error) {
			return gid == gameID && (uid == hostID || uid == guestID), nil
		},
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			if id == gameID {
				return inProgress, nil
			}
			return nil, nil
		},
		ListMatchPlayersFn: func(_ context.Context, _ uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{
				{UserID: hostID, Username: "host"},
				{UserID: guestID, Username: "guest"},
			}, nil
		},
	}
}

func subscribeMatch(t *testing.T, conn *websocket.Conn, gameID uuid.UUID) {
	t.Helper()
	if err := conn.WriteJSON(map[string]string{
		"type": "subscribe",
		"room": ws.MatchRoomID(gameID.String()),
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
}

func drainUntilType(t *testing.T, conn *websocket.Conn, typ string) ws.Envelope {
	t.Helper()
	for {
		env := readEnvelope(t, conn)
		if env.Type == typ {
			return env
		}
	}
}

func TestWSHandler_DisconnectBroadcastAndReconnect(t *testing.T) {
	gameID := uuid.New()
	hostID := uuid.New()
	guestID := uuid.New()
	hostToken := testutil.MakeTestToken(hostID, testJWTSecret)
	guestToken := testutil.MakeTestToken(guestID, testJWTSecret)

	url, hub := startWSTestServer(t, &testutil.MockLobbyRepo{}, multiplayerGameMocks(gameID, hostID, guestID), nil)
	t.Cleanup(func() {
		hub.Disconnect().Evict(gameID)
	})
	hostConn := dialWS(t, url, hostToken)
	guestConn := dialWS(t, url, guestToken)

	subscribeMatch(t, hostConn, gameID)
	subscribeMatch(t, guestConn, gameID)

	_ = hostConn.Close()

	env := drainUntilType(t, guestConn, ws.TypePlayerDisconnected)
	var payload model.PlayerConnectionBroadcast
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("unmarshal disconnected: %v", err)
	}
	if payload.UserID != hostID.String() {
		t.Fatalf("disconnected user = %s, want host", payload.UserID)
	}

	hostConn2 := dialWS(t, url, hostToken)
	subscribeMatch(t, hostConn2, gameID)

	env = drainUntilType(t, guestConn, ws.TypePlayerReconnected)
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		t.Fatalf("unmarshal reconnected: %v", err)
	}
	if payload.UserID != hostID.String() {
		t.Fatalf("reconnected user = %s, want host", payload.UserID)
	}
}

func TestWSHandler_EmptyMatchRoomDoesNotEvictInProgress(t *testing.T) {
	gameID := uuid.New()
	hostID := uuid.New()
	guestID := uuid.New()
	hostToken := testutil.MakeTestToken(hostID, testJWTSecret)
	guestToken := testutil.MakeTestToken(guestID, testJWTSecret)

	url, hub := startWSTestServer(t, &testutil.MockLobbyRepo{}, multiplayerGameMocks(gameID, hostID, guestID), nil)
	t.Cleanup(func() {
		hub.Disconnect().Evict(gameID)
	})
	hostConn := dialWS(t, url, hostToken)
	guestConn := dialWS(t, url, guestToken)

	subscribeMatch(t, hostConn, gameID)
	subscribeMatch(t, guestConn, gameID)

	upload := map[string]any{
		"type": ws.TypePlayerState,
		"room": ws.MatchRoomID(gameID.String()),
		"payload": map[string]any{
			"score": 42,
			"lines": 1,
			"level": 1,
			"alive": true,
			"board": validBoardB64WS(),
		},
	}
	if err := hostConn.WriteJSON(upload); err != nil {
		t.Fatalf("write state: %v", err)
	}
	_ = readEnvelope(t, guestConn)

	_ = hostConn.Close()
	_ = guestConn.Close()
	time.Sleep(50 * time.Millisecond)

	hub.Lifecycle().InitMatch(gameID, []uuid.UUID{hostID, guestID})
	states := hub.MatchStates().List(gameID)
	if len(states) == 0 {
		t.Fatal("match state cache was evicted while match in progress")
	}
}

func TestWSHandler_DisconnectForfeitEndsMatch(t *testing.T) {
	orig := ws.DisconnectGracePeriod
	ws.DisconnectGracePeriod = 50 * time.Millisecond
	t.Cleanup(func() { ws.DisconnectGracePeriod = orig })

	gameID := uuid.New()
	hostID := uuid.New()
	guestID := uuid.New()
	hostToken := testutil.MakeTestToken(hostID, testJWTSecret)
	guestToken := testutil.MakeTestToken(guestID, testJWTSecret)

	ended := make(chan model.EndMatchInput, 1)
	matches := &mockMatchEnder{
		endFn: func(_ context.Context, in model.EndMatchInput) (*model.MatchEndedPayload, error) {
			ended <- in
			return &model.MatchEndedPayload{WinnerUserID: in.SurvivorID}, nil
		},
	}

	url, hub := startWSTestServer(t, &testutil.MockLobbyRepo{}, multiplayerGameMocks(gameID, hostID, guestID), matches)
	t.Cleanup(func() {
		hub.Disconnect().Evict(gameID)
	})
	hostConn := dialWS(t, url, hostToken)
	guestConn := dialWS(t, url, guestToken)

	subscribeMatch(t, hostConn, gameID)
	subscribeMatch(t, guestConn, gameID)

	_ = hostConn.Close()
	drainUntilType(t, guestConn, ws.TypePlayerDisconnected)

	select {
	case in := <-ended:
		if in.SurvivorID == nil || *in.SurvivorID != guestID {
			t.Fatalf("survivor = %+v, want guest", in.SurvivorID)
		}
		if len(in.Eliminations) != 1 || in.Eliminations[0].UserID != hostID || in.Eliminations[0].Reason != "forfeit" {
			t.Fatalf("eliminations = %+v, want host forfeit", in.Eliminations)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("forfeit disconnect did not end match")
	}
}

package ws

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"backend/internal/auth"
	"backend/internal/model"
)

type MemberChecker interface {
	IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error)
}

type GamePlayerChecker interface {
	IsGamePlayer(ctx context.Context, gameID, userID uuid.UUID) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Game, error)
	ListMatchPlayers(ctx context.Context, gameID uuid.UUID) ([]model.MatchPlayerView, error)
}

type MatchEnder interface {
	EndMatch(ctx context.Context, in model.EndMatchInput) (*model.MatchEndedPayload, error)
}

type ChatSender interface {
	SendMessage(ctx context.Context, senderID, recipientID uuid.UUID, body string) (*model.Message, error)
}

type incomingMessage struct {
	Type    string          `json:"type"`
	Room    string          `json:"room"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub       *Hub
	jwtSecret string
	lobbies   MemberChecker
	games     GamePlayerChecker
	matches   MatchEnder
	chat ChatSender
}

func NewHandler(hub *Hub, jwtSecret string, lobbies MemberChecker, games GamePlayerChecker, matches MatchEnder, chat ChatSender) *Handler {
	h := &Handler{hub: hub, jwtSecret: jwtSecret, lobbies: lobbies, games: games, matches: matches, chat: chat}
	hub.SetDisconnectTimeoutHandler(h.forfeitOnDisconnect)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	userID, err := auth.ParseUserID(h.jwtSecret, token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, userID)
	h.hub.Register(client)

	go client.WritePump()
	client.ReadPump(func(c *Client, msg []byte) {
		h.handleMessage(c, msg)
	})
}

func (h *Handler) handleMessage(c *Client, msg []byte) {
	var incoming incomingMessage
	if err := json.Unmarshal(msg, &incoming); err != nil {
		h.sendError(c, "invalid message")
		return
	}

	switch incoming.Type {
	case "subscribe":
		h.handleSubscribe(c, strings.TrimSpace(incoming.Room))
	case TypePlayerState:
		h.handlePlayerState(c, strings.TrimSpace(incoming.Room), incoming.Payload)
	case TypePlayerEliminated:
		h.handlePlayerEliminated(c, strings.TrimSpace(incoming.Room), incoming.Payload)
	default:
		h.sendError(c, "unsupported message type")
	}
}

func (h *Handler) handleSubscribe(c *Client, room string) {
	switch {
	case strings.HasPrefix(room, "lobby:"):
		h.handleLobbySubscribe(c, room)
	case strings.HasPrefix(room, "match:"):
		h.handleMatchSubscribe(c, room)
	default:
		h.sendError(c, "invalid room")
	}
}

func (h *Handler) handleLobbySubscribe(c *Client, room string) {
	lobbyID, err := uuid.Parse(strings.TrimPrefix(room, "lobby:"))
	if err != nil {
		h.sendError(c, "invalid lobby id")
		return
	}

	ok, err := h.lobbies.IsMember(context.Background(), lobbyID, c.UserID())
	if err != nil || !ok {
		h.sendError(c, "not a lobby member")
		return
	}

	h.hub.Subscribe(c, room)
}

func (h *Handler) handleMatchSubscribe(c *Client, room string) {
	gameID, err := uuid.Parse(strings.TrimPrefix(room, "match:"))
	if err != nil {
		h.sendError(c, "invalid match id")
		return
	}

	if !h.verifyMatchParticipant(c, gameID) {
		return
	}

	wasDisconnected := h.hub.Disconnect().MarkReconnected(gameID, c.UserID())

	h.hub.Subscribe(c, room)
	h.hub.ReplayMatchStates(gameID, c)
	h.hub.ReplayDisconnected(gameID, c)

	if wasDisconnected {
		env, err := NewEnvelope(TypePlayerReconnected, model.PlayerConnectionBroadcast{
			UserID: c.UserID().String(),
		})
		if err == nil {
			h.hub.BroadcastMatch(gameID, env)
		}
	}
}

func (h *Handler) handlePlayerState(c *Client, room string, rawPayload json.RawMessage) {
	if !strings.HasPrefix(room, "match:") {
		h.sendError(c, "invalid room")
		return
	}

	gameID, err := uuid.Parse(strings.TrimPrefix(room, "match:"))
	if err != nil {
		h.sendError(c, "invalid match id")
		return
	}

	if !h.verifyMatchParticipant(c, gameID) {
		return
	}

	var upload PlayerStateUpload
	if err := json.Unmarshal(rawPayload, &upload); err != nil {
		h.sendError(c, "invalid player state")
		return
	}
	if err := ValidatePlayerStateUpload(upload); err != nil {
		switch {
		case errors.Is(err, ErrInvalidBoard):
			h.sendError(c, "invalid board")
		case errors.Is(err, ErrInvalidPlayerState):
			h.sendError(c, "invalid player state")
		default:
			h.sendError(c, "invalid player state")
		}
		return
	}

	if !h.hub.AllowPlayerState(gameID, c.UserID()) {
		return
	}

	cached := CachedPlayerState{
		UserID: c.UserID(),
		Score:  upload.Score,
		Lines:  upload.Lines,
		Level:  upload.Level,
		Alive:  upload.Alive,
		Board:  upload.Board,
	}
	h.hub.StorePlayerState(gameID, c.UserID(), cached)

	if !upload.Alive && !h.hub.Lifecycle().IsFinished(gameID) && !h.hub.Lifecycle().IsEliminated(gameID, c.UserID()) {
		h.processElimination(c, gameID, c.UserID(), model.PlayerEliminatedUpload{
			Reason: "blockOut",
			Score:  upload.Score,
			Lines:  upload.Lines,
			Level:  upload.Level,
		})
	}

	broadcast := PlayerStateBroadcast{
		UserID: c.UserID().String(),
		Score:  upload.Score,
		Lines:  upload.Lines,
		Level:  upload.Level,
		Alive:  upload.Alive,
		Board:  upload.Board,
	}
	env, err := NewEnvelope(TypePlayerState, broadcast)
	if err != nil {
		return
	}
	h.hub.BroadcastMatchExcept(gameID, c, env)
}

func (h *Handler) verifyMatchParticipant(c *Client, gameID uuid.UUID) bool {
	ok, err := h.games.IsGamePlayer(context.Background(), gameID, c.UserID())
	if err != nil || !ok {
		h.sendError(c, "not a match participant")
		return false
	}

	game, err := h.games.FindByID(context.Background(), gameID)
	if err != nil {
		h.sendError(c, "match not found")
		return false
	}
	if game.Status != "in_progress" {
		h.sendError(c, "match is not in progress")
		return false
	}
	return true
}

func (h *Handler) sendError(c *Client, message string) {
	env, err := NewEnvelope(TypeError, map[string]string{"error": message})
	if err != nil {
		return
	}
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	c.Send(data)
}

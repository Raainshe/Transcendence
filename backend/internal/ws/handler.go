package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"backend/internal/auth"
)

type MemberChecker interface {
	IsMember(ctx context.Context, lobbyID, userID uuid.UUID) (bool, error)
}

type subscribeMessage struct {
	Type string `json:"type"`
	Room string `json:"room"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub       *Hub
	jwtSecret string
	lobbies   MemberChecker
}

func NewHandler(hub *Hub, jwtSecret string, lobbies MemberChecker) *Handler {
	return &Handler{hub: hub, jwtSecret: jwtSecret, lobbies: lobbies}
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
	var sub subscribeMessage
	if err := json.Unmarshal(msg, &sub); err != nil {
		h.sendError(c, "invalid message")
		return
	}
	if sub.Type != "subscribe" {
		h.sendError(c, "unsupported message type")
		return
	}

	room := strings.TrimSpace(sub.Room)
	if !strings.HasPrefix(room, "lobby:") {
		h.sendError(c, "invalid room")
		return
	}
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

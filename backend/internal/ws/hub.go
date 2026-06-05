package ws

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu          sync.RWMutex
	rooms       map[string]map[*Client]struct{}
	clients     map[*Client]struct{}
	matchStates *MatchStateStore
	rateLimiter *RateLimiter
}

func NewHub() *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Client]struct{}),
		clients:     make(map[*Client]struct{}),
		matchStates: NewMatchStateStore(),
		rateLimiter: NewRateLimiter(),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	roomsToCheck := make([]string, 0, len(c.rooms))
	for room := range c.rooms {
		roomsToCheck = append(roomsToCheck, room)
	}
	delete(h.clients, c)
	for room := range c.rooms {
		if members, ok := h.rooms[room]; ok {
			delete(members, c)
			if len(members) == 0 {
				delete(h.rooms, room)
			}
		}
	}
	h.mu.Unlock()

	for _, room := range roomsToCheck {
		if strings.HasPrefix(room, "match:") {
			h.evictMatchRoomIfEmpty(room)
		}
	}
}

func (h *Hub) evictMatchRoomIfEmpty(room string) {
	gameID, err := uuid.Parse(strings.TrimPrefix(room, "match:"))
	if err != nil {
		return
	}
	h.mu.RLock()
	members := h.rooms[room]
	empty := len(members) == 0
	h.mu.RUnlock()
	if empty {
		h.matchStates.Evict(gameID)
		h.rateLimiter.EvictMatch(gameID)
	}
}

func (h *Hub) Subscribe(c *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c.rooms == nil {
		c.rooms = make(map[string]struct{})
	}
	c.rooms[room] = struct{}{}
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*Client]struct{})
	}
	h.rooms[room][c] = struct{}{}
}

func (h *Hub) Broadcast(room string, env Envelope) {
	h.mu.RLock()
	members := h.rooms[room]
	clients := make([]*Client, 0, len(members))
	for c := range members {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	for _, c := range clients {
		c.Send(data)
	}
}

// BroadcastLobby sends an envelope to lobby:{lobbyID}.
func (h *Hub) BroadcastLobby(lobbyID uuid.UUID, env Envelope) {
	h.Broadcast(LobbyRoomID(lobbyID.String()), env)
}

// BroadcastMatchExcept fans out to match:{gameID} excluding sender.
func (h *Hub) BroadcastMatchExcept(gameID uuid.UUID, sender *Client, env Envelope) {
	room := MatchRoomID(gameID.String())
	h.mu.RLock()
	members := h.rooms[room]
	clients := make([]*Client, 0, len(members))
	for c := range members {
		if c != sender {
			clients = append(clients, c)
		}
	}
	h.mu.RUnlock()

	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	for _, c := range clients {
		c.Send(data)
	}
}

// SubscribeAllInLobbyToMatch subscribes every client in lobby:{lobbyID} to match:{gameID}.
func (h *Hub) SubscribeAllInLobbyToMatch(lobbyID, gameID uuid.UUID) {
	lobbyRoom := LobbyRoomID(lobbyID.String())
	matchRoom := MatchRoomID(gameID.String())

	h.mu.Lock()
	defer h.mu.Unlock()

	members := h.rooms[lobbyRoom]
	if len(members) == 0 {
		return
	}
	if h.rooms[matchRoom] == nil {
		h.rooms[matchRoom] = make(map[*Client]struct{})
	}
	for c := range members {
		if c.rooms == nil {
			c.rooms = make(map[string]struct{})
		}
		c.rooms[matchRoom] = struct{}{}
		h.rooms[matchRoom][c] = struct{}{}
	}
}

// SendToClient delivers a single envelope to one client.
func (h *Hub) SendToClient(c *Client, env Envelope) {
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	c.Send(data)
}

// StorePlayerState caches the latest validated state for a player.
func (h *Hub) StorePlayerState(gameID, userID uuid.UUID, state CachedPlayerState) {
	h.matchStates.Set(gameID, userID, state)
}

// ReplayMatchStates sends cached player.state envelopes for all other players to subscriber.
func (h *Hub) ReplayMatchStates(gameID uuid.UUID, subscriber *Client) {
	states := h.matchStates.List(gameID)
	for _, st := range states {
		if st.UserID == subscriber.UserID() {
			continue
		}
		env, err := NewEnvelope(TypePlayerState, PlayerStateBroadcast{
			UserID: st.UserID.String(),
			Score:  st.Score,
			Lines:  st.Lines,
			Level:  st.Level,
			Alive:  st.Alive,
			Board:  st.Board,
		})
		if err != nil {
			continue
		}
		h.SendToClient(subscriber, env)
	}
}

// AllowPlayerState returns false when the sender is rate-limited.
func (h *Hub) AllowPlayerState(gameID, userID uuid.UUID) bool {
	return h.rateLimiter.Allow(gameID, userID)
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	rooms  map[string]struct{}
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 32),
		userID: userID,
		rooms:  make(map[string]struct{}),
	}
}

func (c *Client) Send(msg []byte) {
	select {
	case c.send <- msg:
	default:
		// drop if client is slow
	}
}

func (c *Client) UserID() uuid.UUID {
	return c.userID
}

func (c *Client) ReadPump(onMessage func(*Client, []byte)) {
	defer func() {
		c.hub.Unregister(c)
		_ = c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if onMessage != nil {
			onMessage(c, msg)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		_ = c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

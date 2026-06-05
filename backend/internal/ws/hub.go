package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]struct{}
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms:   make(map[string]map[*Client]struct{}),
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	for room := range c.rooms {
		if members, ok := h.rooms[room]; ok {
			delete(members, c)
			if len(members) == 0 {
				delete(h.rooms, room)
			}
		}
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

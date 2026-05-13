package ws

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"sync"

	"github.com/denis/web-backgammon/internal/db"
)

// DBRepos bundles the DB repositories the hub and rooms need.
type DBRepos struct {
	Rooms   *db.RoomRepo
	Players *db.PlayerRepo
	Games   *db.GameRepo
}

// Hub manages all active WS rooms.
type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]*Room   // roomCode → Room
	sessions map[string]*Client // sessionToken → Client
	repos    DBRepos
	origins  []string // allowed origins for WS upgrade
}

func NewHub(repos DBRepos, origins []string) *Hub {
	return &Hub{
		rooms:    make(map[string]*Room),
		sessions: make(map[string]*Client),
		repos:    repos,
		origins:  originHostPatterns(origins),
	}
}

func originHostPatterns(origins []string) []string {
	patterns := make([]string, 0, len(origins))
	for _, origin := range origins {
		u, err := url.Parse(origin)
		if err == nil && u.Host != "" {
			patterns = append(patterns, u.Host)
			continue
		}
		patterns = append(patterns, origin)
	}
	return patterns
}

// register attaches a new client to its room, creating the Room if this is
// the first connection. Returns the room so handler.go can start client.run().
func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sessions[c.sessionToken] = c

	room, ok := h.rooms[c.roomCode]
	if !ok {
		room = newRoom(c.roomCode, h)
		h.rooms[c.roomCode] = room
	}
	room.addClient(c)
}

// unregister removes a client from its room and starts the grace period.
func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sessions, c.sessionToken)

	room, ok := h.rooms[c.roomCode]
	if !ok {
		return
	}
	room.removeClient(c)
	// Clean up empty rooms immediately.
	if room.isEmpty() {
		delete(h.rooms, c.roomCode)
	}
}

// dispatch routes a raw inbound JSON frame to the client's room.
func (h *Hub) dispatch(c *Client, data []byte) {
	var env Message
	if err := json.Unmarshal(data, &env); err != nil {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: "invalid_json"}))
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[c.roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}

	room.handleMessage(c, env)
}

// broadcast sends a pre-encoded message to all connected clients in a room.
func (h *Hub) broadcast(roomCode string, msg []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}
	room.broadcast(msg)
}

// broadcastExcept sends to all clients in a room except the excluded one.
func (h *Hub) broadcastExcept(roomCode string, except *Client, msg []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}
	room.broadcastExcept(except, msg)
}

func (h *Hub) logStats() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	slog.Info("hub stats", "rooms", len(h.rooms), "sessions", len(h.sessions))
}

package ws

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"nhooyr.io/websocket"

	"github.com/go-chi/chi/v5"
)

// Handler returns an http.Handler that upgrades GET /ws/{roomCode} to WebSocket.
func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCode := strings.ToUpper(chi.URLParam(r, "roomCode"))

		// Validate session cookie.
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sessionToken := cookie.Value

		ctx := context.Background()

		// Look up player.
		player, err := hub.repos.Players.FindBySession(ctx, sessionToken)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Look up room in DB to get its UUID.
		room, err := hub.repos.Rooms.FindByCode(ctx, roomCode)
		if err != nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		if room.Status != "playing" && room.Status != "waiting" {
			http.Error(w, "room not available", http.StatusGone)
			return
		}

		// Upgrade to WebSocket.
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // TODO: restrict to ALLOWED_ORIGINS in production
		})
		if err != nil {
			slog.Error("ws accept", "err", err)
			return
		}

		c := &Client{
			hub:          hub,
			conn:         conn,
			roomCode:     roomCode,
			sessionToken: sessionToken,
			name:         player.Name,
			send:         make(chan []byte, 64),
		}

		// Check for reconnect first.
		hub.mu.Lock()
		wsRoom, exists := hub.rooms[roomCode]
		if exists && wsRoom.reconnect(c) {
			hub.sessions[sessionToken] = c
			hub.mu.Unlock()
			c.run(r.Context())
			return
		}
		// Set dbRoomID on the ws room (create it if needed).
		if !exists {
			wsRoom = newRoom(roomCode, hub)
			hub.rooms[roomCode] = wsRoom
		}
		wsRoom.dbRoomID = room.ID
		hub.mu.Unlock()

		hub.register(c)
		c.run(r.Context())
	}
}

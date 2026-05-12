package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/microcosm-cc/bluemonday"

	"github.com/denis/web-backgammon/internal/db"
	"github.com/denis/web-backgammon/internal/ws"
)

type Server struct {
	rooms     *db.RoomRepo
	players   *db.PlayerRepo
	games     *db.GameRepo
	sanitizer *bluemonday.Policy
	origins   []string
	hub       *ws.Hub
}

func NewServer(rooms *db.RoomRepo, players *db.PlayerRepo, games *db.GameRepo, origins []string, hub *ws.Hub) *Server {
	return &Server{
		rooms:     rooms,
		players:   players,
		games:     games,
		sanitizer: bluemonday.StrictPolicy(),
		origins:   origins,
		hub:       hub,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(loggingMiddleware)
	r.Use(s.corsMiddleware)

	createRoomLimiter := newIPLimiter(5.0/60, 5)  // 5 req/min burst 5
	joinRoomLimiter := newIPLimiter(10.0/60, 10)  // 10 req/min burst 10

	r.Get("/api/health", s.health)
	r.With(createRoomLimiter.middleware).Post("/api/rooms", s.createRoom)
	r.Get("/api/rooms/{code}", s.getRoom)
	r.With(joinRoomLimiter.middleware).Post("/api/rooms/{code}/join", s.joinRoom)
	r.With(s.requireSession).Get("/api/games/{roomId}/state", s.getGameState)
	r.With(s.requireSession).Get("/api/games/{roomId}/history", s.getGameHistory)
	r.Get("/ws/{roomCode}", ws.Handler(s.hub))

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type createRoomRequest struct {
	CreatorName string `json:"creatorName"`
}

type createRoomResponse struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	URL          string `json:"url"`
	SessionToken string `json:"sessionToken"`
}

type joinRoomRequest struct {
	Name string `json:"name"`
}

type joinRoomResponse struct {
	PlayerID     string  `json:"playerId"`
	Color        *string `json:"color"`
	SessionToken string  `json:"sessionToken"`
}

type roomInfoResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Status      string `json:"status"`
	PlayerCount int    `json:"playerCount"`
}

func (s *Server) createRoom(w http.ResponseWriter, r *http.Request) {
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	name := s.sanitizer.Sanitize(strings.TrimSpace(req.CreatorName))
	if len(name) == 0 || len([]rune(name)) > 40 {
		http.Error(w, "name must be 1–40 characters", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	room, err := s.rooms.Create(ctx)
	if err != nil {
		slog.Error("create room", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	player, err := s.players.Create(ctx, room.ID, name)
	if err != nil {
		slog.Error("create player", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, sessionCookie(player.SessionToken))
	writeJSON(w, http.StatusCreated, createRoomResponse{
		ID:           room.ID,
		Code:         room.Code,
		URL:          "/game/" + room.Code,
		SessionToken: player.SessionToken,
	})
}

func (s *Server) getRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))
	ctx := r.Context()

	room, err := s.rooms.FindByCode(ctx, code)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	count, err := s.players.CountInRoom(ctx, room.ID)
	if err != nil {
		slog.Error("count players", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, roomInfoResponse{
		ID:          room.ID,
		Code:        room.Code,
		Status:      room.Status,
		PlayerCount: count,
	})
}

func (s *Server) joinRoom(w http.ResponseWriter, r *http.Request) {
	code := strings.ToUpper(chi.URLParam(r, "code"))

	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	name := s.sanitizer.Sanitize(strings.TrimSpace(req.Name))
	if len(name) == 0 || len([]rune(name)) > 40 {
		http.Error(w, "name must be 1–40 characters", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	room, err := s.rooms.FindByCode(ctx, code)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	if room.Status != "waiting" {
		http.Error(w, "room is not available", http.StatusGone)
		return
	}
	count, err := s.players.CountInRoom(ctx, room.ID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if count >= 2 {
		http.Error(w, "room is full", http.StatusGone)
		return
	}

	player, err := s.players.Create(ctx, room.ID, name)
	if err != nil {
		slog.Error("create player on join", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := s.rooms.UpdateStatus(ctx, room.ID, "playing"); err != nil {
		slog.Error("update room status", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, sessionCookie(player.SessionToken))
	writeJSON(w, http.StatusOK, joinRoomResponse{
		PlayerID:     player.ID,
		Color:        nil, // assigned by WS layer in Phase 3
		SessionToken: player.SessionToken,
	})
}

func sessionCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     "session_token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
		Path:     "/",
	}
}

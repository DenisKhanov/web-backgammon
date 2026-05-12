package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type gameStateResponse struct {
	ID            string          `json:"id"`
	Phase         string          `json:"phase"`
	CurrentTurn   *string         `json:"currentTurn"`
	Dice          []int           `json:"dice"`
	RemainingDice []int           `json:"remainingDice"`
	BoardState    json.RawMessage `json:"boardState"`
	MoveCount     int             `json:"moveCount"`
	Winner        *string         `json:"winner,omitempty"`
	IsMars        bool            `json:"isMars"`
}

func (s *Server) getGameState(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomId")
	player := playerFromCtx(r.Context())
	if player == nil || player.RoomID != roomID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	g, err := s.games.FindByRoomID(r.Context(), roomID)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, gameStateResponse{
		ID:            g.ID,
		Phase:         g.Phase,
		CurrentTurn:   g.CurrentTurn,
		Dice:          g.Dice,
		RemainingDice: g.RemainingDice,
		BoardState:    json.RawMessage(g.BoardState),
		MoveCount:     g.MoveCount,
		Winner:        g.Winner,
		IsMars:        g.IsMars,
	})
}

func (s *Server) getGameHistory(w http.ResponseWriter, r *http.Request) {
	// Phase 3 will implement move log retrieval.
	// For now return an empty list so the endpoint is live.
	writeJSON(w, http.StatusOK, []any{})
}

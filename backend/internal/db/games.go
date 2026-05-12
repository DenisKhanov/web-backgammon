package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GameRepo struct{ pool *pgxpool.Pool }

func NewGameRepo(pool *pgxpool.Pool) *GameRepo { return &GameRepo{pool: pool} }

// Create inserts a new game record with the initial board state (as JSON bytes).
// Called by the WS layer (Phase 3) once both players connect.
func (r *GameRepo) Create(ctx context.Context, roomID string, boardJSON []byte) (*GameRecord, error) {
	var g GameRecord
	err := r.pool.QueryRow(ctx, `
		INSERT INTO games (room_id, board_state, phase)
		VALUES ($1, $2, 'rolling_first')
		RETURNING id, room_id, board_state, current_turn, dice, remaining_dice,
		          phase, winner, is_mars, turn_started_at, move_count, created_at, updated_at`,
		roomID, boardJSON,
	).Scan(&g.ID, &g.RoomID, &g.BoardState, &g.CurrentTurn, &g.Dice, &g.RemainingDice,
		&g.Phase, &g.Winner, &g.IsMars, &g.TurnStartedAt, &g.MoveCount, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert game: %w", err)
	}
	return &g, nil
}

// FindByRoomID retrieves the active game for a room. Returns error if not found.
func (r *GameRepo) FindByRoomID(ctx context.Context, roomID string) (*GameRecord, error) {
	var g GameRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, room_id, board_state, current_turn, dice, remaining_dice,
		       phase, winner, is_mars, turn_started_at, move_count, created_at, updated_at
		FROM games WHERE room_id = $1`, roomID,
	).Scan(&g.ID, &g.RoomID, &g.BoardState, &g.CurrentTurn, &g.Dice, &g.RemainingDice,
		&g.Phase, &g.Winner, &g.IsMars, &g.TurnStartedAt, &g.MoveCount, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find game by room: %w", err)
	}
	return &g, nil
}

// UpdateState persists the full game state snapshot (called by WS after each move).
func (r *GameRepo) UpdateState(ctx context.Context, gameID string,
	boardJSON []byte, currentTurn string, dice, remainingDice []int,
	phase string, winner *string, isMars bool, moveCount int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE games SET
			board_state    = $1,
			current_turn   = $2,
			dice           = $3,
			remaining_dice = $4,
			phase          = $5,
			winner         = $6,
			is_mars        = $7,
			move_count     = $8,
			updated_at     = NOW()
		WHERE id = $9`,
		boardJSON, currentTurn, dice, remainingDice, phase, winner, isMars, moveCount, gameID)
	return err
}

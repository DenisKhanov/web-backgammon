package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PlayerRepo struct{ pool *pgxpool.Pool }

func NewPlayerRepo(pool *pgxpool.Pool) *PlayerRepo { return &PlayerRepo{pool: pool} }

func (r *PlayerRepo) Create(ctx context.Context, roomID, name string) (*Player, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}
	var p Player
	err = r.pool.QueryRow(ctx, `
		INSERT INTO players (room_id, name, session_token)
		VALUES ($1, $2, $3)
		RETURNING id, room_id, name, color, session_token, joined_at, last_seen_at`,
		roomID, name, token,
	).Scan(&p.ID, &p.RoomID, &p.Name, &p.Color, &p.SessionToken, &p.JoinedAt, &p.LastSeenAt)
	if err != nil {
		return nil, fmt.Errorf("insert player: %w", err)
	}
	return &p, nil
}

func (r *PlayerRepo) FindBySession(ctx context.Context, token string) (*Player, error) {
	var p Player
	err := r.pool.QueryRow(ctx, `
		SELECT id, room_id, name, color, session_token, joined_at, last_seen_at
		FROM players WHERE session_token = $1`, token,
	).Scan(&p.ID, &p.RoomID, &p.Name, &p.Color, &p.SessionToken, &p.JoinedAt, &p.LastSeenAt)
	if err != nil {
		return nil, fmt.Errorf("find player by session: %w", err)
	}
	return &p, nil
}

func (r *PlayerRepo) FindByRoom(ctx context.Context, roomID string) ([]*Player, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, room_id, name, color, session_token, joined_at, last_seen_at
		FROM players WHERE room_id = $1 ORDER BY joined_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var players []*Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.RoomID, &p.Name, &p.Color,
			&p.SessionToken, &p.JoinedAt, &p.LastSeenAt); err != nil {
			return nil, err
		}
		players = append(players, &p)
	}
	return players, rows.Err()
}

func (r *PlayerRepo) CountInRoom(ctx context.Context, roomID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM players WHERE room_id = $1", roomID).Scan(&count)
	return count, err
}

// generateSessionToken returns a 64-character hex string (32 random bytes).
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

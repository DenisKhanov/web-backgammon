package db

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo struct{ pool *pgxpool.Pool }

func NewRoomRepo(pool *pgxpool.Pool) *RoomRepo { return &RoomRepo{pool: pool} }

func (r *RoomRepo) Create(ctx context.Context) (*Room, error) {
	code, err := generateRoomCode()
	if err != nil {
		return nil, fmt.Errorf("generate room code: %w", err)
	}
	expiresAt := time.Now().Add(24 * time.Hour)
	var room Room
	err = r.pool.QueryRow(ctx, `
		INSERT INTO rooms (code, status, expires_at)
		VALUES ($1, 'waiting', $2)
		RETURNING id, code, status, created_at, expires_at`,
		code, expiresAt,
	).Scan(&room.ID, &room.Code, &room.Status, &room.CreatedAt, &room.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("insert room: %w", err)
	}
	return &room, nil
}

func (r *RoomRepo) FindByCode(ctx context.Context, code string) (*Room, error) {
	var room Room
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, status, created_at, expires_at
		FROM rooms WHERE code = $1`, code,
	).Scan(&room.ID, &room.Code, &room.Status, &room.CreatedAt, &room.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("find room by code: %w", err)
	}
	return &room, nil
}

func (r *RoomRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx,
		"UPDATE rooms SET status = $1 WHERE id = $2", status, id)
	return err
}

// generateRoomCode returns 8 uppercase base32 characters (A-Z, 2-7).
// 5 random bytes → 40 bits → 8 × 5-bit base32 chars, no padding.
func generateRoomCode() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(b), nil
}

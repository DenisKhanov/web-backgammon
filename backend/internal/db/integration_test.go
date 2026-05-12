//go:build integration

package db_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/denis/web-backgammon/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgC, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgC.Terminate(ctx) })

	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := db.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Resolve migrations directory from this test file's location.
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../../migrations")
	require.NoError(t, db.RunMigrations(ctx, pool, migrationsDir))

	return pool
}

// TestRoomRepo_CreateAndFind creates a room and retrieves it by code.
func TestRoomRepo_CreateAndFind(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	rooms := db.NewRoomRepo(pool)

	room, err := rooms.Create(ctx)
	require.NoError(t, err)
	assert.Len(t, room.Code, 8)
	assert.Equal(t, "waiting", room.Status)
	assert.True(t, room.ExpiresAt.After(time.Now()))

	found, err := rooms.FindByCode(ctx, room.Code)
	require.NoError(t, err)
	assert.Equal(t, room.ID, found.ID)
}

// TestRoomRepo_UpdateStatus changes room status to "playing".
func TestRoomRepo_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	rooms := db.NewRoomRepo(pool)

	room, err := rooms.Create(ctx)
	require.NoError(t, err)

	require.NoError(t, rooms.UpdateStatus(ctx, room.ID, "playing"))

	found, err := rooms.FindByCode(ctx, room.Code)
	require.NoError(t, err)
	assert.Equal(t, "playing", found.Status)
}

// TestPlayerRepo_CreateAndFindBySession creates a player and retrieves by session token.
func TestPlayerRepo_CreateAndFindBySession(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)

	room, err := rooms.Create(ctx)
	require.NoError(t, err)

	p, err := players.Create(ctx, room.ID, "Алексей")
	require.NoError(t, err)
	assert.Len(t, p.SessionToken, 64)
	assert.Equal(t, "Алексей", p.Name)
	assert.Nil(t, p.Color)

	found, err := players.FindBySession(ctx, p.SessionToken)
	require.NoError(t, err)
	assert.Equal(t, p.ID, found.ID)
}

// TestPlayerRepo_CountInRoom counts players after two joins.
func TestPlayerRepo_CountInRoom(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)

	room, _ := rooms.Create(ctx)

	count, err := players.CountInRoom(ctx, room.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	players.Create(ctx, room.ID, "Алексей")
	players.Create(ctx, room.ID, "Мария")

	count, err = players.CountInRoom(ctx, room.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestGameRepo_CreateAndFind stores and retrieves a game record.
func TestGameRepo_CreateAndFind(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()
	rooms := db.NewRoomRepo(pool)
	games := db.NewGameRepo(pool)

	room, _ := rooms.Create(ctx)
	boardJSON, _ := json.Marshal(map[string]any{"points": []any{}, "borneOff": []int{0, 0, 0}})

	g, err := games.Create(ctx, room.ID, boardJSON)
	require.NoError(t, err)
	assert.Equal(t, "rolling_first", g.Phase)

	found, err := games.FindByRoomID(ctx, room.ID)
	require.NoError(t, err)
	assert.Equal(t, g.ID, found.ID)
}

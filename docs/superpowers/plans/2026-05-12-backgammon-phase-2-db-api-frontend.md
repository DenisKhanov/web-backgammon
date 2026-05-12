# Web-Backgammon Phase 2 — DB Layer + REST API + Frontend Scaffold

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the PostgreSQL persistence layer with versioned migrations, a chi-based REST API for room and player management, and a Next.js frontend scaffold (neumorphic UI, Zustand stores, landing page, room waiting page) — delivering a runnable lobby where two players can find each other before the WebSocket game begins in Phase 3.

**Architecture:** Phase 2A — custom pgx/v5 migration runner + three repository structs (RoomRepo, PlayerRepo, GameRepo) integration-tested via testcontainers. Phase 2B — chi v5 HTTP server with logging, CORS, per-IP rate-limiting, session-cookie auth middleware, and five REST handlers. Phase 2C — Zustand stores (game/chat/ui), neumorphic Tailwind extensions, Button/Input/Card components, landing page with create/join forms, room-waiting polling page.

**Tech Stack:** Go 1.26 (chi/v5, pgx/v5, google/uuid, bluemonday, x/time/rate, testcontainers-go/modules/postgres), Next.js 14 (App Router), Zustand, Framer Motion, Tailwind CSS v3.

**Ссылки:** `docs/specs/backgammon-design.md` — секции 1, 5, 6, 7, 8.

---

## File Structure

```
backend/
├── migrations/
│   ├── 001_rooms.up.sql
│   ├── 001_rooms.down.sql
│   ├── 002_players.up.sql
│   ├── 002_players.down.sql
│   ├── 003_games.up.sql
│   ├── 003_games.down.sql
│   ├── 004_moves_results_chat.up.sql
│   └── 004_moves_results_chat.down.sql
├── internal/
│   ├── config/
│   │   └── config.go           # Config struct from env
│   ├── db/
│   │   ├── db.go               # Pool init + migration runner
│   │   ├── model.go            # Room, Player, GameRecord structs
│   │   ├── rooms.go            # RoomRepo
│   │   ├── players.go          # PlayerRepo
│   │   ├── games.go            # GameRepo
│   │   └── integration_test.go # testcontainers tests (build tag: integration)
│   └── api/
│       ├── server.go           # Server struct + chi router wiring
│       ├── middleware.go       # logging, CORS, rate-limit, requireSession
│       ├── rooms.go            # createRoom, getRoom, joinRoom handlers
│       ├── games.go            # getGameState, getGameHistory handlers
│       ├── health.go           # GET /api/health
│       └── integration_test.go # handler tests (build tag: integration)
└── cmd/server/main.go          # updated: config → DB → migrations → server

frontend/
├── tailwind.config.ts          # extend: neo-raised/neo-inset shadows
├── src/
│   ├── lib/
│   │   └── types.ts            # shared TypeScript types
│   ├── stores/
│   │   ├── gameStore.ts
│   │   ├── chatStore.ts
│   │   └── uiStore.ts
│   ├── components/ui/
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   └── Card.tsx
│   └── app/
│       ├── page.tsx            # landing: create-room + join-by-code
│       └── room/[code]/
│           └── page.tsx        # waiting-for-opponent page

backend/Dockerfile              # multi-stage: go build → distroless/static
docker-compose.yml              # add backend service
```

---

## Phase 2A: Database Layer

### Task 1: Add Go backend dependencies

**Files:** Modify `backend/go.mod` (via `go get`)

- [ ] **Step 1: Add all runtime and test dependencies**

```bash
cd backend
go get github.com/go-chi/chi/v5@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/google/uuid@latest
go get github.com/joho/godotenv@latest
go get golang.org/x/time@latest
go get github.com/microcosm-cc/bluemonday@latest
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/testcontainers/testcontainers-go/modules/postgres@latest
go mod tidy
```

- [ ] **Step 2: Verify build still compiles**

```bash
cd backend && go build ./...
```

Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
cd backend
git add go.mod go.sum
git commit -m "chore(backend): add chi, pgx/v5, uuid, bluemonday, x/time, testcontainers deps"
```

---

### Task 2: Config package

**Files:** Create `backend/internal/config/config.go`

- [ ] **Step 1: Write the config file**

```go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	Port           string
	AllowedOrigins []string
	MigrationsDir  string
}

func Load() (*Config, error) {
	// Load .env if present; ignore error (file may not exist in production).
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://bg_user:bg_pass@localhost:5433/backgammon?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	originsRaw := os.Getenv("ALLOWED_ORIGINS")
	if originsRaw == "" {
		originsRaw = "http://localhost:3000"
	}
	origins := strings.Split(originsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &Config{
		DatabaseURL:    dsn,
		Port:           port,
		AllowedOrigins: origins,
		MigrationsDir:  migrationsDir,
	}, nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/config/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/config/
git commit -m "feat(config): add Config struct loaded from env with defaults"
```

---

### Task 3: SQL migration files

**Files:** Create all files in `backend/migrations/`

- [ ] **Step 1: Create migrations directory and 001_rooms**

```bash
mkdir -p backend/migrations
```

`backend/migrations/001_rooms.up.sql`:
```sql
CREATE TABLE rooms (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code       VARCHAR(8) UNIQUE NOT NULL,
  status     VARCHAR(20) NOT NULL DEFAULT 'waiting',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_rooms_code     ON rooms(code);
CREATE INDEX idx_rooms_expires  ON rooms(expires_at) WHERE status != 'finished';
```

`backend/migrations/001_rooms.down.sql`:
```sql
DROP TABLE IF EXISTS rooms;
```

- [ ] **Step 2: Create 002_players**

`backend/migrations/002_players.up.sql`:
```sql
CREATE TABLE players (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id       UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  name          VARCHAR(40) NOT NULL,
  color         VARCHAR(5),
  session_token VARCHAR(64) UNIQUE NOT NULL,
  joined_at     TIMESTAMPTZ DEFAULT NOW(),
  last_seen_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_players_room    ON players(room_id);
CREATE INDEX idx_players_session ON players(session_token);
```

`backend/migrations/002_players.down.sql`:
```sql
DROP TABLE IF EXISTS players;
```

- [ ] **Step 3: Create 003_games**

`backend/migrations/003_games.up.sql`:
```sql
CREATE TABLE games (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id        UUID UNIQUE NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  board_state    JSONB NOT NULL,
  current_turn   VARCHAR(5),
  dice           INTEGER[] NOT NULL DEFAULT '{}',
  remaining_dice INTEGER[] NOT NULL DEFAULT '{}',
  phase          VARCHAR(20) NOT NULL DEFAULT 'rolling_first',
  winner         VARCHAR(5),
  is_mars        BOOLEAN DEFAULT FALSE,
  turn_started_at TIMESTAMPTZ DEFAULT NOW(),
  move_count     INTEGER DEFAULT 0,
  created_at     TIMESTAMPTZ DEFAULT NOW(),
  updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_games_room ON games(room_id);
```

`backend/migrations/003_games.down.sql`:
```sql
DROP TABLE IF EXISTS games;
```

- [ ] **Step 4: Create 004_moves_results_chat**

`backend/migrations/004_moves_results_chat.up.sql`:
```sql
CREATE TABLE moves (
  id           BIGSERIAL PRIMARY KEY,
  game_id      UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  move_number  INTEGER NOT NULL,
  player_color VARCHAR(5) NOT NULL,
  dice_rolled  INTEGER[] NOT NULL,
  moves_data   JSONB NOT NULL,
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(game_id, move_number)
);

CREATE INDEX idx_moves_game ON moves(game_id);

CREATE TABLE game_results (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_id     UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  winner_name VARCHAR(40) NOT NULL,
  loser_name  VARCHAR(40) NOT NULL,
  is_mars     BOOLEAN DEFAULT FALSE,
  total_moves INTEGER NOT NULL,
  duration_sec INTEGER NOT NULL,
  finished_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_results_finished ON game_results(finished_at DESC);

CREATE TABLE chat_messages (
  id         BIGSERIAL PRIMARY KEY,
  room_id    UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  player_id  UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
  text       VARCHAR(500) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_chat_room ON chat_messages(room_id, created_at);
```

`backend/migrations/004_moves_results_chat.down.sql`:
```sql
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS game_results;
DROP TABLE IF EXISTS moves;
```

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/
git commit -m "feat(db): add SQL migrations for rooms, players, games, moves, results, chat"
```

---

### Task 4: DB pool and migration runner

**Files:** Create `backend/internal/db/db.go`

- [ ] **Step 1: Write db.go**

```go
package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates and validates a pgx connection pool.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}
	cfg.MaxConns = 25
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping DB: %w", err)
	}
	return pool, nil
}

// RunMigrations applies all pending *.up.sql files from dir in lexicographic order.
// Applied versions are tracked in the schema_migrations table.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		_ = rows.Scan(&v)
		applied[v] = true
	}
	rows.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var upFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fname := range upFiles {
		version := strings.TrimSuffix(fname, ".up.sql")
		if applied[version] {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(dir, fname))
		if err != nil {
			return fmt.Errorf("read %s: %w", fname, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", fname, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", fname, err)
		}
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations(version) VALUES($1)", version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", fname, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", fname, err)
		}
	}
	return nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/db/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/db.go
git commit -m "feat(db): add pgxpool Connect and custom SQL migration runner"
```

---

### Task 5: DB models

**Files:** Create `backend/internal/db/model.go`

- [ ] **Step 1: Write model.go**

```go
package db

import "time"

type Room struct {
	ID        string
	Code      string
	Status    string // waiting | playing | finished | abandoned
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Player struct {
	ID           string
	RoomID       string
	Name         string
	Color        *string // nil until assigned; "white" | "black"
	SessionToken string
	JoinedAt     time.Time
	LastSeenAt   time.Time
}

type GameRecord struct {
	ID            string
	RoomID        string
	BoardState    []byte // JSONB raw bytes
	CurrentTurn   *string
	Dice          []int
	RemainingDice []int
	Phase         string
	Winner        *string
	IsMars        bool
	TurnStartedAt time.Time
	MoveCount     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/db/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/model.go
git commit -m "feat(db): add Room, Player, GameRecord model types"
```

---

### Task 6: Room repository

**Files:** Create `backend/internal/db/rooms.go`

- [ ] **Step 1: Write rooms.go**

```go
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
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/db/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/rooms.go
git commit -m "feat(db): add RoomRepo with Create, FindByCode, UpdateStatus"
```

---

### Task 7: Player repository

**Files:** Create `backend/internal/db/players.go`

- [ ] **Step 1: Write players.go**

```go
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
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/db/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/players.go
git commit -m "feat(db): add PlayerRepo with Create, FindBySession, FindByRoom, CountInRoom"
```

---

### Task 8: Game repository

**Files:** Create `backend/internal/db/games.go`

The GameRepo is written now but only called from Phase 3 (WS layer). Its `Create` takes `boardJSON []byte` so the `db` package does not import `internal/game`.

- [ ] **Step 1: Write games.go**

```go
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
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/db/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/db/games.go
git commit -m "feat(db): add GameRepo with Create, FindByRoomID, UpdateState"
```

---

### Task 9: DB integration tests

**Files:** Create `backend/internal/db/integration_test.go`

Build tag `integration` keeps these out of plain `go test ./...`.

- [ ] **Step 1: Write integration_test.go**

```go
//go:build integration

package db_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/denis/web-backgammon/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *db.Pool {
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
```

Note: `db.Pool` is an alias — replace with `*pgxpool.Pool` if the db package doesn't export a `Pool` type. Adjust the import to use the concrete type from `pgxpool`.

Actually, `db.Connect` returns `*pgxpool.Pool`. Import pgx in the test:

```go
import "github.com/jackc/pgx/v5/pgxpool"
```

And change `*db.Pool` to `*pgxpool.Pool` everywhere in the test helper.

- [ ] **Step 2: Fix the test helper signature to use `*pgxpool.Pool`**

Replace `func setupTestDB(t *testing.T) *db.Pool {` with `func setupTestDB(t *testing.T) *pgxpool.Pool {`

And add `"github.com/jackc/pgx/v5/pgxpool"` to imports.

- [ ] **Step 3: Run integration tests** (requires Docker)

```bash
cd backend && go test -tags integration -v -timeout 120s ./internal/db/...
```

Expected output ends with:
```
PASS
ok  	github.com/denis/web-backgammon/internal/db
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/db/integration_test.go
git commit -m "test(db): add integration tests for RoomRepo, PlayerRepo, GameRepo via testcontainers"
```

---

## Phase 2B: REST API

### Task 10: API server setup

**Files:** Create `backend/internal/api/server.go`

- [ ] **Step 1: Write server.go**

```go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/microcosm-cc/bluemonday"

	"github.com/denis/web-backgammon/internal/db"
)

type Server struct {
	rooms     *db.RoomRepo
	players   *db.PlayerRepo
	games     *db.GameRepo
	sanitizer *bluemonday.Policy
	origins   []string
}

func NewServer(rooms *db.RoomRepo, players *db.PlayerRepo, games *db.GameRepo, origins []string) *Server {
	return &Server{
		rooms:     rooms,
		players:   players,
		games:     games,
		sanitizer: bluemonday.StrictPolicy(),
		origins:   origins,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(loggingMiddleware)
	r.Use(s.corsMiddleware)

	createRoomLimiter := newIPLimiter(5.0/60, 5)   // 5 req/min burst 5
	joinRoomLimiter   := newIPLimiter(10.0/60, 10)  // 10 req/min burst 10

	r.Get("/api/health", s.health)
	r.With(createRoomLimiter.middleware).Post("/api/rooms", s.createRoom)
	r.Get("/api/rooms/{code}", s.getRoom)
	r.With(joinRoomLimiter.middleware).Post("/api/rooms/{code}/join", s.joinRoom)
	r.With(s.requireSession).Get("/api/games/{roomId}/state", s.getGameState)
	r.With(s.requireSession).Get("/api/games/{roomId}/history", s.getGameHistory)

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/api/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/server.go
git commit -m "feat(api): add Server struct and chi router with middleware chain"
```

---

### Task 11: Middleware

**Files:** Create `backend/internal/api/middleware.go`

- [ ] **Step 1: Write middleware.go**

```go
package api

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/denis/web-backgammon/internal/db"
)

// --- Logging ---

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"dur", time.Since(start).String(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// --- CORS ---

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range s.origins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				break
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Rate limiting ---

type ipLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newIPLimiter(r rate.Limit, b int) *ipLimiter {
	return &ipLimiter{visitors: make(map[string]*rate.Limiter), r: r, b: b}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.visitors[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.r, l.b)
	l.visitors[ip] = lim
	return lim
}

func (l *ipLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !l.get(ip).Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Session auth ---

type contextKey string

const playerCtxKey contextKey = "player"

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		player, err := s.players.FindBySession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), playerCtxKey, player)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func playerFromCtx(ctx context.Context) *db.Player {
	p, _ := ctx.Value(playerCtxKey).(*db.Player)
	return p
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/api/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/middleware.go
git commit -m "feat(api): add logging, CORS, IP rate-limit, and session-auth middleware"
```

---

### Task 12: Room handlers

**Files:** Create `backend/internal/api/rooms.go`

- [ ] **Step 1: Write rooms.go**

```go
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
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/api/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/rooms.go
git commit -m "feat(api): add createRoom, getRoom, joinRoom handlers"
```

---

### Task 13: Game state and health handlers

**Files:** Create `backend/internal/api/games.go` and `backend/internal/api/health.go`

- [ ] **Step 1: Write games.go**

```go
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
```

- [ ] **Step 2: Write health.go**

```go
package api

import "net/http"

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

- [ ] **Step 3: Verify**

```bash
cd backend && go build ./internal/api/...
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/api/games.go backend/internal/api/health.go
git commit -m "feat(api): add getGameState (resync endpoint) and health check"
```

---

### Task 14: API integration tests

**Files:** Create `backend/internal/api/integration_test.go`

- [ ] **Step 1: Write integration_test.go**

```go
//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/denis/web-backgammon/internal/api"
	"github.com/denis/web-backgammon/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestServer(t *testing.T) (http.Handler, *pgxpool.Pool) {
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

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../../migrations")
	require.NoError(t, db.RunMigrations(ctx, pool, migrationsDir))

	srv := api.NewServer(
		db.NewRoomRepo(pool),
		db.NewPlayerRepo(pool),
		db.NewGameRepo(pool),
		[]string{"http://localhost:3000"},
	)
	return srv.Router(), pool
}

func TestHealth(t *testing.T) {
	handler, _ := setupTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	assert.Equal(t, "ok", body["status"])
}

func TestCreateRoom(t *testing.T) {
	handler, _ := setupTestServer(t)
	body := bytes.NewBufferString(`{"creatorName":"Алексей"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp["code"], 8)
	assert.NotEmpty(t, resp["sessionToken"])
	assert.NotEmpty(t, resp["id"])

	// Cookie must be set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie)
	assert.True(t, sessionCookie.HttpOnly)
}

func TestGetRoom(t *testing.T) {
	handler, _ := setupTestServer(t)

	// Create a room first
	body := bytes.NewBufferString(`{"creatorName":"Алексей"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)

	// Now fetch the room
	req2 := httptest.NewRequest(http.MethodGet, "/api/rooms/"+createResp["code"], nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var roomResp map[string]any
	json.NewDecoder(w2.Body).Decode(&roomResp)
	assert.Equal(t, "waiting", roomResp["status"])
	assert.Equal(t, float64(1), roomResp["playerCount"])
}

func TestJoinRoom(t *testing.T) {
	handler, _ := setupTestServer(t)

	// Create room
	body := bytes.NewBufferString(`{"creatorName":"Алексей"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)

	// Join room
	body2 := bytes.NewBufferString(`{"name":"Мария"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/rooms/"+createResp["code"]+"/join", body2)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var joinResp map[string]any
	json.NewDecoder(w2.Body).Decode(&joinResp)
	assert.NotEmpty(t, joinResp["playerId"])
	assert.Nil(t, joinResp["color"])

	// Room status must now be "playing"
	req3 := httptest.NewRequest(http.MethodGet, "/api/rooms/"+createResp["code"], nil)
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	var roomResp map[string]any
	json.NewDecoder(w3.Body).Decode(&roomResp)
	assert.Equal(t, "playing", roomResp["status"])
	assert.Equal(t, float64(2), roomResp["playerCount"])
}

func TestJoinRoom_Full(t *testing.T) {
	handler, _ := setupTestServer(t)

	// Create room
	body := bytes.NewBufferString(`{"creatorName":"Алексей"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)
	code := createResp["code"]

	// Join once (second player)
	body2 := bytes.NewBufferString(`{"name":"Мария"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/api/rooms/"+code+"/join", body2)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Join again (third player — must be rejected)
	body3 := bytes.NewBufferString(`{"name":"Иван"}`)
	req3 := httptest.NewRequest(http.MethodPost, "/api/rooms/"+code+"/join", body3)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	handler.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusGone, w3.Code)
}

func TestGetGameState_NotFound(t *testing.T) {
	handler, _ := setupTestServer(t)

	// Create room and player to get a session token
	body := bytes.NewBufferString(`{"creatorName":"Алексей"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	var createResp map[string]string
	json.NewDecoder(w.Body).Decode(&createResp)
	cookies := w.Result().Cookies()

	// Request game state — no game exists yet → 404
	req2 := httptest.NewRequest(http.MethodGet, "/api/games/"+createResp["id"]+"/state", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}
```

- [ ] **Step 2: Run API integration tests**

```bash
cd backend && go test -tags integration -v -timeout 120s ./internal/api/...
```

Expected: all 6 tests pass.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/integration_test.go
git commit -m "test(api): add integration tests for all REST endpoints via testcontainers"
```

---

### Task 15: Updated main.go and backend Dockerfile

**Files:** Modify `backend/cmd/server/main.go`, create `backend/Dockerfile`

- [ ] **Step 1: Write main.go**

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/denis/web-backgammon/internal/api"
	"github.com/denis/web-backgammon/internal/config"
	"github.com/denis/web-backgammon/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("connect to DB", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}

	srv := api.NewServer(
		db.NewRoomRepo(pool),
		db.NewPlayerRepo(pool),
		db.NewGameRepo(pool),
		cfg.AllowedOrigins,
	)

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify the full binary builds**

```bash
cd backend && go build -o /tmp/backgammon-server ./cmd/server/main.go
```

Expected: no errors.

- [ ] **Step 3: Write backend/Dockerfile**

```dockerfile
# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Runtime stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

- [ ] **Step 4: Commit**

```bash
git add backend/cmd/server/main.go backend/Dockerfile
git commit -m "feat(server): wire config, DB pool, migrations, and chi API server in main.go"
```

---

## Phase 2C: Frontend Scaffold

### Task 16: Install frontend runtime dependencies

**Files:** Modify `frontend/package.json`

- [ ] **Step 1: Install Zustand and Framer Motion**

```bash
cd frontend && npm install zustand framer-motion
```

- [ ] **Step 2: Verify the build still works**

```bash
cd frontend && npm run build
```

Expected: build completes without errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/package.json frontend/package-lock.json
git commit -m "feat(frontend): add zustand and framer-motion dependencies"
```

---

### Task 17: Tailwind neumorphic extensions

**Files:** Modify `frontend/tailwind.config.ts`

The existing config already has the neumorphic color palette. Add box-shadow utilities and board-specific colours.

- [ ] **Step 1: Update tailwind.config.ts**

```ts
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        neo: {
          bg:     "#e0e5ec",
          light:  "#ffffff",
          dark:   "#a3b1c6",
          accent: "#6c63ff",
        },
        board: {
          green: "#2d5016",
          wood:  "#8B4513",
        },
        checker: {
          white: "#f0f0f0",
          black: "#3a3a3a",
        },
      },
      boxShadow: {
        "neo-raised": "6px 6px 12px #a3b1c6, -6px -6px 12px #ffffff",
        "neo-inset":  "inset 6px 6px 12px #a3b1c6, inset -6px -6px 12px #ffffff",
        "neo-sm":     "3px 3px 6px #a3b1c6, -3px -3px 6px #ffffff",
      },
    },
  },
  plugins: [],
};

export default config;
```

- [ ] **Step 2: Verify the build picks up new classes**

```bash
cd frontend && npm run build
```

Expected: success.

- [ ] **Step 3: Commit**

```bash
git add frontend/tailwind.config.ts
git commit -m "feat(frontend): extend Tailwind with neo-raised/neo-inset shadows and board colours"
```

---

### Task 18: TypeScript types

**Files:** Create `frontend/src/lib/types.ts`

- [ ] **Step 1: Create src/lib/ directory and types.ts**

```bash
mkdir -p frontend/src/lib
```

`frontend/src/lib/types.ts`:
```ts
export type Color = 'white' | 'black';

export type GamePhase =
  | 'waiting'
  | 'rolling_first'
  | 'playing'
  | 'bearing_off'
  | 'finished';

export interface Point {
  owner: 0 | 1 | 2; // 0=none, 1=white, 2=black
  checkers: number;
}

export interface Board {
  Points: Point[];   // [25] — index 0 unused, 1–24 are board points
  BorneOff: number[]; // [3] — index 1=white, 2=black
}

export interface Move {
  from: number;
  to: number;
  die: number;
}

export interface GameState {
  id: string;
  phase: GamePhase;
  currentTurn: Color | null;
  dice: number[];
  remainingDice: number[];
  boardState: Board;
  moveCount: number;
  winner?: Color;
  isMars: boolean;
}

export interface ChatMessage {
  from: string;
  text: string;
  time: string; // "HH:MM"
}

export interface Room {
  id: string;
  code: string;
  status: string;
  playerCount: number;
}

export interface CreateRoomResponse {
  id: string;
  code: string;
  url: string;
  sessionToken: string;
}

export interface JoinRoomResponse {
  playerId: string;
  color: Color | null;
  sessionToken: string;
}
```

- [ ] **Step 2: Verify TypeScript compiles**

```bash
cd frontend && npm run typecheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/types.ts
git commit -m "feat(frontend): add shared TypeScript types for game, room, and chat"
```

---

### Task 19: Zustand stores

**Files:** Create `frontend/src/stores/gameStore.ts`, `chatStore.ts`, `uiStore.ts`

- [ ] **Step 1: Create stores directory**

```bash
mkdir -p frontend/src/stores
```

`frontend/src/stores/gameStore.ts`:
```ts
import { create } from 'zustand';
import type { Board, Color, GamePhase, Move } from '@/lib/types';

interface GameStore {
  board: Board | null;
  dice: number[];
  remainingDice: number[];
  turn: Color | null;
  phase: GamePhase;
  myColor: Color | null;
  selectedChecker: number | null;
  validMoves: Move[];
  timeLeft: number;

  // Setters called by the WS hook (Phase 3)
  setGameState: (state: Partial<GameStore>) => void;
  selectChecker: (point: number | null) => void;
  setMyColor: (color: Color) => void;
  reset: () => void;
}

const initialState = {
  board: null,
  dice: [],
  remainingDice: [],
  turn: null,
  phase: 'waiting' as GamePhase,
  myColor: null,
  selectedChecker: null,
  validMoves: [],
  timeLeft: 60,
};

export const useGameStore = create<GameStore>((set) => ({
  ...initialState,
  setGameState: (state) => set((prev) => ({ ...prev, ...state })),
  selectChecker: (point) => set({ selectedChecker: point }),
  setMyColor: (color) => set({ myColor: color }),
  reset: () => set(initialState),
}));
```

`frontend/src/stores/chatStore.ts`:
```ts
import { create } from 'zustand';
import type { ChatMessage } from '@/lib/types';

interface ChatStore {
  messages: ChatMessage[];
  addMessage: (msg: ChatMessage) => void;
  clear: () => void;
}

export const useChatStore = create<ChatStore>((set) => ({
  messages: [],
  addMessage: (msg) =>
    set((state) => ({ messages: [...state.messages, msg] })),
  clear: () => set({ messages: [] }),
}));
```

`frontend/src/stores/uiStore.ts`:
```ts
import { create } from 'zustand';

interface UIStore {
  showChat: boolean;
  animationsEnabled: boolean;
  soundEnabled: boolean;
  toggleChat: () => void;
  toggleAnimations: () => void;
  toggleSound: () => void;
}

export const useUIStore = create<UIStore>((set) => ({
  showChat: false,
  animationsEnabled: true,
  soundEnabled: true,
  toggleChat: () => set((s) => ({ showChat: !s.showChat })),
  toggleAnimations: () =>
    set((s) => ({ animationsEnabled: !s.animationsEnabled })),
  toggleSound: () => set((s) => ({ soundEnabled: !s.soundEnabled })),
}));
```

- [ ] **Step 2: Verify TypeScript compilation**

```bash
cd frontend && npm run typecheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/
git commit -m "feat(frontend): add Zustand stores for game, chat, and UI state"
```

---

### Task 20: UI components

**Files:** Create `frontend/src/components/ui/Button.tsx`, `Input.tsx`, `Card.tsx`

- [ ] **Step 1: Create component directory**

```bash
mkdir -p frontend/src/components/ui
```

`frontend/src/components/ui/Button.tsx`:
```tsx
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'raised' | 'inset';
  children: React.ReactNode;
}

export default function Button({
  variant = 'raised',
  children,
  className = '',
  ...props
}: ButtonProps) {
  const base =
    'px-6 py-3 rounded-xl font-semibold transition-all duration-150 bg-neo-bg text-neo-accent select-none';
  const raised = 'shadow-neo-raised active:shadow-neo-inset';
  const inset = 'shadow-neo-inset';

  return (
    <button
      className={`${base} ${variant === 'raised' ? raised : inset} ${className} disabled:opacity-50 disabled:cursor-not-allowed`}
      {...props}
    >
      {children}
    </button>
  );
}
```

`frontend/src/components/ui/Input.tsx`:
```tsx
interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

export default function Input({ className = '', ...props }: InputProps) {
  return (
    <input
      className={`w-full px-4 py-3 rounded-xl bg-neo-bg shadow-neo-inset outline-none
        text-gray-700 placeholder-gray-400 focus:ring-2 focus:ring-neo-accent/50
        transition-shadow ${className}`}
      {...props}
    />
  );
}
```

`frontend/src/components/ui/Card.tsx`:
```tsx
interface CardProps {
  title?: string;
  children: React.ReactNode;
  className?: string;
}

export default function Card({ title, children, className = '' }: CardProps) {
  return (
    <div
      className={`bg-neo-bg shadow-neo-raised rounded-2xl p-6 w-full max-w-sm ${className}`}
    >
      {title && (
        <h2 className="text-lg font-semibold text-gray-600 mb-4">{title}</h2>
      )}
      {children}
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compilation**

```bash
cd frontend && npm run typecheck
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ui/
git commit -m "feat(frontend): add neumorphic Button, Input, Card UI components"
```

---

### Task 21: Landing page

**Files:** Modify `frontend/src/app/page.tsx`, update `frontend/src/app/globals.css`

- [ ] **Step 1: Update globals.css to set neumorphic background**

`frontend/src/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

body {
  background-color: #e0e5ec;
  min-height: 100vh;
}
```

- [ ] **Step 2: Write the landing page**

`frontend/src/app/page.tsx`:
```tsx
'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Card from '@/components/ui/Card';
import type { CreateRoomResponse } from '@/lib/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export default function HomePage() {
  const router = useRouter();
  const [creatorName, setCreatorName] = useState('');
  const [joinCode, setJoinCode] = useState('');
  const [joinName, setJoinName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/api/rooms`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ creatorName }),
      });
      if (!res.ok) {
        setError(await res.text());
        return;
      }
      const data: CreateRoomResponse = await res.json();
      router.push(`/room/${data.code}`);
    } catch {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  }

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError('');
    const code = joinCode.toUpperCase();
    try {
      const res = await fetch(`${API_URL}/api/rooms/${code}/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ name: joinName }),
      });
      if (res.status === 410) {
        setError('Комната заполнена или уже завершена');
        return;
      }
      if (res.status === 404) {
        setError('Комната не найдена');
        return;
      }
      if (!res.ok) {
        setError(await res.text());
        return;
      }
      router.push(`/room/${code}`);
    } catch {
      setError('Ошибка соединения с сервером');
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-8 p-6">
      <h1 className="text-4xl font-bold text-neo-accent tracking-tight">
        Длинные нарды
      </h1>

      {error && (
        <p className="text-red-500 text-sm text-center max-w-sm">{error}</p>
      )}

      <Card title="Создать игру">
        <form onSubmit={handleCreate} className="flex flex-col gap-3">
          <Input
            placeholder="Ваше имя (до 40 символов)"
            value={creatorName}
            onChange={(e) => setCreatorName(e.target.value)}
            maxLength={40}
            required
          />
          <Button type="submit" disabled={loading}>
            {loading ? 'Создаём...' : 'Создать комнату'}
          </Button>
        </form>
      </Card>

      <Card title="Войти по коду">
        <form onSubmit={handleJoin} className="flex flex-col gap-3">
          <Input
            placeholder="Код комнаты (8 символов)"
            value={joinCode}
            onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
            maxLength={8}
            required
          />
          <Input
            placeholder="Ваше имя (до 40 символов)"
            value={joinName}
            onChange={(e) => setJoinName(e.target.value)}
            maxLength={40}
            required
          />
          <Button type="submit" disabled={loading}>
            {loading ? 'Входим...' : 'Войти'}
          </Button>
        </form>
      </Card>
    </main>
  );
}
```

- [ ] **Step 3: Verify the build**

```bash
cd frontend && npm run build
```

Expected: build succeeds, no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/app/page.tsx frontend/src/app/globals.css
git commit -m "feat(frontend): add landing page with create-room and join-by-code forms"
```

---

### Task 22: Room waiting page

**Files:** Create `frontend/src/app/room/[code]/page.tsx`

- [ ] **Step 1: Create the route directory**

```bash
mkdir -p "frontend/src/app/room/[code]"
```

- [ ] **Step 2: Write the waiting page**

`frontend/src/app/room/[code]/page.tsx`:
```tsx
'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Card from '@/components/ui/Card';
import type { Room } from '@/lib/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';

export default function RoomPage() {
  const params = useParams();
  const code = (params.code as string).toUpperCase();
  const [room, setRoom] = useState<Room | null>(null);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const poll = async () => {
      try {
        const res = await fetch(`${API_URL}/api/rooms/${code}`, {
          credentials: 'include',
        });
        if (!res.ok) {
          setError('Комната не найдена');
          return;
        }
        const data: Room = await res.json();
        setRoom(data);
      } catch {
        setError('Ошибка соединения');
      }
    };

    poll();
    const interval = setInterval(poll, 2000);
    return () => clearInterval(interval);
  }, [code]);

  async function copyLink() {
    const url = `${window.location.origin}/room/${code}`;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <main className="min-h-screen bg-neo-bg flex flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-3xl font-bold text-neo-accent">Ожидание соперника</h1>

      {error && <p className="text-red-500 text-center">{error}</p>}

      {!error && !room && (
        <p className="text-gray-500">Загрузка...</p>
      )}

      {room && (
        <Card title="Комната">
          <p className="text-2xl font-mono text-center text-neo-accent tracking-widest mb-2">
            {room.code}
          </p>

          <div className="flex justify-between text-sm text-gray-600 mb-4">
            <span>Игроков: {room.playerCount}/2</span>
            <span>
              {room.status === 'waiting'
                ? '⏳ Ожидание'
                : '▶ Игра началась!'}
            </span>
          </div>

          <button
            onClick={copyLink}
            className="w-full text-sm text-neo-accent underline cursor-pointer hover:no-underline"
          >
            {copied ? '✓ Скопировано!' : 'Скопировать ссылку для соперника'}
          </button>
        </Card>
      )}

      {room && room.playerCount < 2 && (
        <p className="text-sm text-gray-500 animate-pulse">
          Ожидаем второго игрока...
        </p>
      )}
    </main>
  );
}
```

- [ ] **Step 3: Verify the build**

```bash
cd frontend && npm run build
```

Expected: success, no TypeScript errors.

- [ ] **Step 4: Commit**

```bash
git add "frontend/src/app/room/"
git commit -m "feat(frontend): add room waiting page with 2s polling and copy-link button"
```

---

## Infrastructure

### Task 23: docker-compose.yml with backend service

**Files:** Modify `docker-compose.yml`, create `frontend/Dockerfile`

- [ ] **Step 1: Write frontend/Dockerfile**

`frontend/Dockerfile`:
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

- [ ] **Step 2: Enable Next.js standalone output**

Add `output: 'standalone'` to `frontend/next.config.js` (create the file if it doesn't exist):

```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
};

module.exports = nextConfig;
```

- [ ] **Step 3: Update docker-compose.yml**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: backgammon-postgres
    environment:
      POSTGRES_DB: ${POSTGRES_DB:-backgammon}
      POSTGRES_USER: ${POSTGRES_USER:-bg_user}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-bg_pass}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-bg_user}"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build: ./backend
    container_name: backgammon-backend
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER:-bg_user}:${POSTGRES_PASSWORD:-bg_pass}@postgres:5432/${POSTGRES_DB:-backgammon}?sslmode=disable
      PORT: "8080"
      ALLOWED_ORIGINS: ${ALLOWED_ORIGINS:-http://localhost:3000}
      MIGRATIONS_DIR: migrations
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy

  frontend:
    build: ./frontend
    container_name: backgammon-frontend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL:-http://localhost:8080}
    ports:
      - "3000:3000"
    depends_on:
      - backend

volumes:
  postgres_data:
```

- [ ] **Step 4: Verify docker compose config is valid**

```bash
docker compose config --quiet
```

Expected: no output (valid config).

- [ ] **Step 5: Smoke test — start DB + backend locally**

```bash
# Terminal 1: start postgres
docker compose up postgres -d

# Terminal 2: run backend
cd backend && DATABASE_URL="postgres://bg_user:bg_pass@localhost:5433/backgammon?sslmode=disable" go run ./cmd/server/main.go
```

Expected backend output includes:
```
INFO server starting addr=:8080
```

In a third terminal:
```bash
curl -s http://localhost:8080/api/health
```
Expected: `{"status":"ok"}`

```bash
curl -s -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -d '{"creatorName":"Тест"}'
```
Expected: `{"id":"...","code":"XXXXXXXX","url":"/game/XXXXXXXX","sessionToken":"..."}`

```bash
docker compose down
```

- [ ] **Step 6: Commit**

```bash
git add docker-compose.yml frontend/Dockerfile frontend/next.config.js
git commit -m "feat(infra): add backend service to docker-compose, frontend Dockerfile with standalone output"
```

---

## Phase 2 Complete

- [ ] **Tag the milestone**

```bash
git tag phase-2-db-api-frontend
git push origin master --tags
```

- [ ] **Run all unit tests (no integration)**

```bash
cd backend && go test ./...
cd frontend && npm run typecheck && npm run build
```

Expected: all green.

- [ ] **Run integration tests** (requires Docker)

```bash
cd backend && go test -tags integration -timeout 180s ./...
```

Expected: all green.

---

## Self-Review Checklist

**Spec coverage:**
- ✓ Секция 5 (БД): все 6 таблиц с индексами.
- ✓ Секция 6 (REST API): все 5 маршрутов + WS-заглушка (Phase 3).
- ✓ Секция 6 (Rate limiting): `POST /api/rooms` 5/мин, `POST .../join` 10/мин.
- ✓ Секция 7 (Безопасность): `session_token` HttpOnly SameSite=Strict, bluemonday для имён, сервер — источник истины.
- ✓ Секция 1 (File structure): `internal/db/`, `internal/api/`, `migrations/`.
- ✓ Секция 4 (Frontend): Zustand stores, TypeScript типы, neumorphic Tailwind, landing, /room/[code].

**Gaps:**
- `GET /api/games/:roomId/history` возвращает пустой массив — полная реализация в Phase 3 (нет таблицы moves до Phase 3 WS-слоя).
- Цвета игроков (`color` field) не назначаются REST-слоем — назначаются WS-слоем в Phase 3 при броске на первый ход.
- E2E и performance тесты — Phase 6.

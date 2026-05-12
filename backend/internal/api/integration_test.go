//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		nil, // hub not needed for REST-only tests
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

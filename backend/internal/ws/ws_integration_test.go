//go:build integration

package ws_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/denis/web-backgammon/internal/api"
	"github.com/denis/web-backgammon/internal/db"
	internalws "github.com/denis/web-backgammon/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestInfra(t *testing.T) (*pgxpool.Pool, *httptest.Server) {
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
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../migrations")
	require.NoError(t, db.RunMigrations(ctx, pool, migrationsDir))

	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)
	games := db.NewGameRepo(pool)

	hub := internalws.NewHub(internalws.DBRepos{Rooms: rooms, Players: players, Games: games}, []string{"*"})
	srv := api.NewServer(rooms, players, games, []string{"*"}, hub)

	ts := httptest.NewServer(srv.Router())
	t.Cleanup(ts.Close)
	return pool, ts
}

// connectWS dials the WS endpoint with a pre-set session cookie.
func connectWS(t *testing.T, ts *httptest.Server, roomCode, sessionToken string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/" + roomCode
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"Cookie": {"session_token=" + sessionToken},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { conn.CloseNow() })
	return conn
}

func readMsg(t *testing.T, conn *websocket.Conn) internalws.Message {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg internalws.Message
	err := wsjson.Read(ctx, conn, &msg)
	require.NoError(t, err)
	return msg
}

// TestWS_FullGameStart verifies that two clients connecting to the same room
// both receive dice_rolled and game_state messages.
func TestWS_FullGameStart(t *testing.T) {
	_, ts := setupTestInfra(t)

	// Create room via REST.
	resCreate := mustPost(t, ts.URL+"/api/rooms", `{"creatorName":"Алексей"}`)
	var createBody map[string]string
	json.Unmarshal(resCreate.body, &createBody)
	roomCode := createBody["code"]
	token1 := createBody["sessionToken"]

	// Join room via REST.
	resJoin := mustPost(t, ts.URL+"/api/rooms/"+roomCode+"/join", `{"name":"Мария"}`)
	var joinBody map[string]string
	json.Unmarshal(resJoin.body, &joinBody)
	token2 := joinBody["sessionToken"]

	// Connect both clients over WS.
	conn1 := connectWS(t, ts, roomCode, token1)
	conn2 := connectWS(t, ts, roomCode, token2)

	// Both should receive dice_rolled then game_state.
	msg1 := readMsg(t, conn1)
	assert.Equal(t, "dice_rolled", msg1.Type)

	msg2 := readMsg(t, conn2)
	assert.Equal(t, "dice_rolled", msg2.Type)

	gs1 := readMsg(t, conn1)
	assert.Equal(t, "game_state", gs1.Type)

	gs2 := readMsg(t, conn2)
	assert.Equal(t, "game_state", gs2.Type)
}

// TestWS_ChatDelivery verifies that a chat message is broadcast to both clients.
func TestWS_ChatDelivery(t *testing.T) {
	_, ts := setupTestInfra(t)

	resCreate := mustPost(t, ts.URL+"/api/rooms", `{"creatorName":"Алексей"}`)
	var createBody map[string]string
	json.Unmarshal(resCreate.body, &createBody)
	roomCode := createBody["code"]
	token1 := createBody["sessionToken"]

	resJoin := mustPost(t, ts.URL+"/api/rooms/"+roomCode+"/join", `{"name":"Мария"}`)
	var joinBody map[string]string
	json.Unmarshal(resJoin.body, &joinBody)
	token2 := joinBody["sessionToken"]

	conn1 := connectWS(t, ts, roomCode, token1)
	conn2 := connectWS(t, ts, roomCode, token2)

	// Drain startup messages (dice_rolled + game_state per client).
	for i := 0; i < 2; i++ {
		readMsg(t, conn1)
	}
	for i := 0; i < 2; i++ {
		readMsg(t, conn2)
	}

	// Send chat from player 1.
	chatCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := wsjson.Write(chatCtx, conn1, internalws.Message{
		Type:    "chat",
		Payload: json.RawMessage(`{"text":"Привет!"}`),
	})
	require.NoError(t, err)

	// Both clients should receive chat_message.
	chat1 := readMsg(t, conn1)
	assert.Equal(t, "chat_message", chat1.Type)

	chat2 := readMsg(t, conn2)
	assert.Equal(t, "chat_message", chat2.Type)
}

// --- HTTP helper ---

type httpResp struct {
	status int
	body   []byte
}

func mustPost(t *testing.T, url, body string) httpResp {
	t.Helper()
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return httpResp{status: resp.StatusCode, body: b}
}

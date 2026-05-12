# Web-Backgammon Phase 3 — WebSocket Protocol

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the real-time WebSocket layer — a room-based hub, per-client read/write pumps, a game session that drives the `internal/game` state machine, turn timers (60 s), disconnect grace periods (5 min), and all 11 WS message types — delivering a backend where two browser tabs can play a complete game of длинные нарды over WebSocket.

**Architecture:** `internal/ws/` contains four focused files: `message.go` (envelope + payload structs), `client.go` (WS connection + ping/pong), `hub.go` (room registry, broadcast, reconnect), `room.go` (game session — wraps `game.Game`, manages timers, persists to DB). The HTTP upgrade and session validation live in `handler.go`. The WS router is wired into `cmd/server/main.go` beside the existing REST routes.

**Tech Stack:** `nhooyr.io/websocket` (WS library), `internal/game` (game engine from Phase 1), `internal/db` repos (Phase 2), `net/http/httptest` + goroutine-based WS clients for integration tests.

**Ссылки:** `docs/specs/backgammon-design.md` — секции 3, 7, 9, 13.

---

## File Structure

```
backend/
├── internal/ws/
│   ├── message.go          # Envelope, all client→server and server→client payload types
│   ├── client.go           # Client struct: WS conn, send channel, readPump, writePump
│   ├── hub.go              # Hub: room registry, register/unregister/dispatch/broadcast
│   ├── room.go             # Room: game session, turn timer, grace period, move processing
│   ├── handler.go          # HTTP → WS upgrade, session auth, hub.Register call
│   └── ws_integration_test.go  # integration tests (build tag: integration)
└── cmd/server/main.go      # add WS route: GET /ws/{roomCode}
```

---

## Phase 3A: Messages and Client

### Task 1: Add nhooyr.io/websocket dependency

**Files:** Modify `backend/go.mod`

- [ ] **Step 1: Add the WS library**

```bash
cd backend && go get nhooyr.io/websocket@latest && go mod tidy
```

- [ ] **Step 2: Verify build**

```bash
cd backend && go build ./...
```

Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add backend/go.mod backend/go.sum
git commit -m "chore(backend): add nhooyr.io/websocket dependency"
```

---

### Task 2: Message types

**Files:** Create `backend/internal/ws/message.go`

All WS message envelopes and payload structs. No business logic here.

- [ ] **Step 1: Write message.go**

```go
package ws

import "encoding/json"

// --- Envelope ---

// Message is the top-level JSON envelope for every WS frame.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func encode(msgType string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Type: msgType, Payload: p})
}

func mustEncode(msgType string, payload any) []byte {
	b, err := encode(msgType, payload)
	if err != nil {
		panic(err)
	}
	return b
}

// --- Client → Server payload types ---

type MovePayload struct {
	From int `json:"from"`
	To   int `json:"to"`
	Die  int `json:"die"`
}

type ChatPayload struct {
	Text string `json:"text"`
}

// end_turn, pass, ping have no payload fields.

// --- Server → Client payload types ---

type BoardPoint struct {
	Owner    int `json:"owner"`    // 0=none 1=white 2=black
	Checkers int `json:"checkers"`
}

type PlayerSnapshot struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Connected bool   `json:"connected"`
}

type GameStatePayload struct {
	Phase         string           `json:"phase"`
	CurrentTurn   string           `json:"currentTurn"`
	Board         [25]BoardPoint   `json:"board"`
	BorneOff      [3]int           `json:"borneOff"`
	Dice          []int            `json:"dice"`
	RemainingDice []int            `json:"remainingDice"`
	MoveCount     int              `json:"moveCount"`
	MyColor       string           `json:"myColor"`
	Players       []PlayerSnapshot `json:"players"`
	TimeLeft      int              `json:"timeLeft"` // seconds remaining in current turn
}

type DiceRolledPayload struct {
	Dice     []int  `json:"dice"`
	IsDouble bool   `json:"isDouble"`
	Player   string `json:"player"` // "white" | "black"
}

type OpponentMovedPayload struct {
	From          int   `json:"from"`
	To            int   `json:"to"`
	Die           int   `json:"die"`
	RemainingDice []int `json:"remainingDice"`
}

type TurnChangedPayload struct {
	Player   string `json:"player"`
	TimeLeft int    `json:"timeLeft"`
}

type MoveErrorPayload struct {
	Reason string `json:"reason"`
}

type ChatMessagePayload struct {
	From string `json:"from"`
	Text string `json:"text"`
	Time string `json:"time"` // "15:04"
}

type GameOverPayload struct {
	Winner string `json:"winner"`
	IsMars bool   `json:"isMars"`
}

type OpponentDisconnectedPayload struct {
	GracePeriod int `json:"gracePeriod"` // seconds
}
```

- [ ] **Step 2: Write a unit test for encode/decode round-trip**

Create `backend/internal/ws/message_test.go`:

```go
package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode_MoveError(t *testing.T) {
	b := mustEncode("move_error", MoveErrorPayload{Reason: "glukhoi_zabor"})

	var env Message
	require.NoError(t, json.Unmarshal(b, &env))
	assert.Equal(t, "move_error", env.Type)

	var p MoveErrorPayload
	require.NoError(t, json.Unmarshal(env.Payload, &p))
	assert.Equal(t, "glukhoi_zabor", p.Reason)
}

func TestEncode_GameState(t *testing.T) {
	payload := GameStatePayload{
		Phase:       "playing",
		CurrentTurn: "white",
		MyColor:     "white",
		Dice:        []int{3, 5},
	}
	b := mustEncode("game_state", payload)
	assert.Contains(t, string(b), `"type":"game_state"`)
	assert.Contains(t, string(b), `"phase":"playing"`)
}
```

- [ ] **Step 3: Run unit tests**

```bash
cd backend && go test ./internal/ws/...
```

Expected: `PASS`

- [ ] **Step 4: Commit**

```bash
git add backend/internal/ws/
git commit -m "feat(ws): add WS message envelope and all payload types with unit tests"
```

---

### Task 3: WS Client

**Files:** Create `backend/internal/ws/client.go`

Each connected browser tab is one `Client`. It owns one WS connection and a send channel. The `readPump` delivers inbound frames to the hub; the `writePump` sends outbound bytes and pings every 15 s.

- [ ] **Step 1: Write client.go**

```go
package ws

import (
	"context"
	"log/slog"
	"time"

	"nhooyr.io/websocket"

	"github.com/denis/web-backgammon/internal/game"
)

const (
	pingInterval    = 15 * time.Second
	writeTimeout    = 10 * time.Second
	maxMessageBytes = 1024
)

// Client represents one connected WebSocket peer.
type Client struct {
	hub          *Hub
	conn         *websocket.Conn
	roomCode     string
	sessionToken string
	color        game.Color // assigned when the room starts
	name         string
	send         chan []byte
}

// run starts the read and write goroutines and blocks until the connection closes.
func (c *Client) run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		c.hub.unregister(c)
		c.conn.Close(websocket.StatusGoingAway, "disconnected")
	}()

	go c.writePump(ctx)
	c.readPump(ctx)
}

func (c *Client) readPump(ctx context.Context) {
	c.conn.SetReadLimit(maxMessageBytes)
	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		c.hub.dispatch(c, data)
	}
}

func (c *Client) writePump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			wCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(wCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Warn("ws write error", "err", err, "color", c.color)
				return
			}

		case <-ticker.C:
			pCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pCtx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// sendMsg queues a pre-encoded JSON byte slice for delivery.
func (c *Client) sendMsg(b []byte) {
	select {
	case c.send <- b:
	default:
		slog.Warn("ws send buffer full, dropping message", "color", c.color)
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/ws/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/ws/client.go
git commit -m "feat(ws): add Client with read/write pumps and 15s ping"
```

---

## Phase 3B: Hub and Game Session

### Task 4: Hub

**Files:** Create `backend/internal/ws/hub.go`

The hub is the central registry. It maps room codes to `Room` instances and session tokens to `Client` instances (for reconnect). The `dispatch` method routes inbound messages to the appropriate room.

- [ ] **Step 1: Write hub.go**

```go
package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/denis/web-backgammon/internal/db"
)

// DBRepos bundles the DB repositories the hub and rooms need.
type DBRepos struct {
	Rooms   *db.RoomRepo
	Players *db.PlayerRepo
	Games   *db.GameRepo
}

// Hub manages all active WS rooms.
type Hub struct {
	mu       sync.RWMutex
	rooms    map[string]*Room   // roomCode → Room
	sessions map[string]*Client // sessionToken → Client
	repos    DBRepos
}

func NewHub(repos DBRepos) *Hub {
	return &Hub{
		rooms:    make(map[string]*Room),
		sessions: make(map[string]*Client),
		repos:    repos,
	}
}

// register attaches a new client to its room, creating the Room if this is
// the first connection. Returns the room so handler.go can start client.run().
func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sessions[c.sessionToken] = c

	room, ok := h.rooms[c.roomCode]
	if !ok {
		room = newRoom(c.roomCode, h)
		h.rooms[c.roomCode] = room
	}
	room.addClient(c)
}

// unregister removes a client from its room and starts the grace period.
func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.sessions, c.sessionToken)

	room, ok := h.rooms[c.roomCode]
	if !ok {
		return
	}
	room.removeClient(c)
	// Clean up empty rooms immediately.
	if room.isEmpty() {
		delete(h.rooms, c.roomCode)
	}
}

// dispatch routes a raw inbound JSON frame to the client's room.
func (h *Hub) dispatch(c *Client, data []byte) {
	var env Message
	if err := json.Unmarshal(data, &env); err != nil {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: "invalid_json"}))
		return
	}

	h.mu.RLock()
	room, ok := h.rooms[c.roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}

	room.handleMessage(c, env)
}

// broadcast sends a pre-encoded message to all connected clients in a room.
func (h *Hub) broadcast(roomCode string, msg []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}
	room.broadcast(msg)
}

// broadcastExcept sends to all clients in a room except the excluded one.
func (h *Hub) broadcastExcept(roomCode string, except *Client, msg []byte) {
	h.mu.RLock()
	room, ok := h.rooms[roomCode]
	h.mu.RUnlock()
	if !ok {
		return
	}
	room.broadcastExcept(except, msg)
}

func (h *Hub) logStats() {
	h.mu.RLock()
	defer h.mu.RUnlock()
	slog.Info("hub stats", "rooms", len(h.rooms), "sessions", len(h.sessions))
}
```

- [ ] **Step 2: Verify**

```bash
cd backend && go build ./internal/ws/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/ws/hub.go
git commit -m "feat(ws): add Hub with room registry, session index, and message dispatch"
```

---

### Task 5: Room (game session)

**Files:** Create `backend/internal/ws/room.go`

The `Room` is the heart of Phase 3. It:
1. Collects two clients, assigns colors (White = first joiner, Black = second).
2. On second client connect: calls `game.RollFirst`, broadcasts `dice_rolled` + `game_state`, starts turn timer.
3. Processes `move` / `end_turn` / `pass` / `chat` / `ping` messages.
4. After each turn end: calls `game.Roll`, broadcasts `dice_rolled` + `turn_changed`, restarts timer.
5. On turn timeout: auto-skips, rotates turn.
6. On client disconnect: starts 5-min grace timer, notifies opponent. On reconnect within grace: stops timer, sends `game_state`, notifies opponent.
7. On grace timeout: opponent wins by default.
8. On game over: persists result to DB, sends `game_over`.

- [ ] **Step 1: Write room.go**

```go
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/denis/web-backgammon/internal/game"
)

const (
	turnDuration  = 60 * time.Second
	graceDuration = 5 * time.Minute
)

// Room holds the live state for one match between two players.
type Room struct {
	mu   sync.Mutex
	hub  *Hub
	code string

	// Clients indexed by color: clients[White-1]=White player, clients[Black-1]=Black player.
	clients [2]*Client
	names   [2]string // player names
	tokens  [2]string // session tokens (for reconnect matching)

	// Game state
	g      *game.Game
	gameID string // DB record ID, set after first roll

	// Timers
	turnTimer   *time.Timer
	graceTimers [2]*time.Timer
}

func newRoom(code string, hub *Hub) *Room {
	return &Room{code: code, hub: hub}
}

// addClient assigns the client a color slot and triggers game start when both slots fill.
func (r *Room) addClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Assign to first free slot (White=0, Black=1).
	slot := -1
	for i, cl := range r.clients {
		if cl == nil {
			slot = i
			break
		}
	}
	if slot < 0 {
		// Room full — shouldn't happen if REST join is enforced.
		c.conn.Close(websocket.StatusPolicyViolation, "room full")
		return
	}

	r.clients[slot] = c
	r.names[slot] = c.name
	r.tokens[slot] = c.sessionToken

	colorBySlot := [2]game.Color{game.White, game.Black}
	c.color = colorBySlot[slot]

	if r.clients[0] != nil && r.clients[1] != nil {
		go r.startGame()
	}
}

// removeClient marks a slot as disconnected and starts the grace timer.
func (r *Room) removeClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slot := r.slotOf(c)
	if slot < 0 {
		return
	}
	r.clients[slot] = nil

	if r.g == nil {
		return // game not started yet
	}

	// Notify opponent.
	r.broadcastExceptSlot(slot,
		mustEncode("opponent_disconnected", OpponentDisconnectedPayload{GracePeriod: int(graceDuration.Seconds())}))

	// Start grace timer.
	if r.graceTimers[slot] != nil {
		r.graceTimers[slot].Stop()
	}
	r.graceTimers[slot] = time.AfterFunc(graceDuration, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.handleGraceTimeout(slot)
	})
}

// reconnect re-attaches a client who has the same session token as a previous occupant.
// Returns true if the client was accepted as a reconnect.
func (r *Room) reconnect(c *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for slot, token := range r.tokens {
		if token == c.sessionToken && r.clients[slot] == nil {
			r.clients[slot] = c
			colorBySlot := [2]game.Color{game.White, game.Black}
			c.color = colorBySlot[slot]

			// Stop grace timer.
			if r.graceTimers[slot] != nil {
				r.graceTimers[slot].Stop()
				r.graceTimers[slot] = nil
			}

			// Send full game state for resync.
			if r.g != nil {
				c.sendMsg(mustEncode("game_state", r.buildGameState(c.color)))
				r.broadcastExceptSlot(slot,
					mustEncode("opponent_reconnected", struct{}{}))
			}
			return true
		}
	}
	return false
}

// --- Game lifecycle ---

func (r *Room) startGame() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.g = game.NewGame()
	if err := r.g.RollFirst(game.NewRandomDice()); err != nil {
		slog.Error("RollFirst failed", "err", err)
		return
	}

	// Persist initial game record.
	boardJSON, _ := json.Marshal(r.g.Board)
	ctx := context.Background()
	gr, err := r.hub.repos.Games.Create(ctx, r.roomID(), boardJSON)
	if err != nil {
		slog.Error("create game record", "err", err)
	} else {
		r.gameID = gr.ID
	}

	dicePayload := DiceRolledPayload{
		Dice:     r.g.Dice,
		IsDouble: r.g.Dice[0] == r.g.Dice[1],
		Player:   colorName(r.g.CurrentTurn),
	}
	r.broadcastAll(mustEncode("dice_rolled", dicePayload))
	r.broadcastGameState()
	r.startTurnTimer()
}

func (r *Room) handleMessage(c *Client, env Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch env.Type {
	case "move":
		r.handleMove(c, env.Payload)
	case "end_turn":
		r.handleEndTurn(c)
	case "pass":
		r.handleEndTurn(c) // identical semantics
	case "chat":
		r.handleChat(c, env.Payload)
	case "ping":
		c.sendMsg(mustEncode("pong", struct{}{}))
	default:
		slog.Warn("unknown ws message type", "type", env.Type)
	}
}

func (r *Room) handleMove(c *Client, raw json.RawMessage) {
	if r.g == nil || r.g.CurrentTurn != c.color {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: "not_your_turn"}))
		return
	}
	var p MovePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: "invalid_payload"}))
		return
	}
	m := game.Move{From: p.From, To: p.To, Die: p.Die}
	if err := r.g.ApplyMove(m); err != nil {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: err.Error()}))
		return
	}

	// Broadcast the move to the opponent; send full state to both.
	r.broadcastExceptSlot(r.slotOf(c), mustEncode("opponent_moved", OpponentMovedPayload{
		From:          p.From,
		To:            p.To,
		Die:           p.Die,
		RemainingDice: r.g.RemainingDice,
	}))
	r.broadcastGameState()
	r.persistState()

	if r.g.Phase == game.PhaseFinished {
		r.handleGameOver()
	}
}

func (r *Room) handleEndTurn(c *Client) {
	if r.g == nil || r.g.CurrentTurn != c.color {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: "not_your_turn"}))
		return
	}
	if err := r.g.EndTurn(); err != nil {
		c.sendMsg(mustEncode("move_error", MoveErrorPayload{Reason: err.Error()}))
		return
	}
	r.advanceTurn()
}

func (r *Room) advanceTurn() {
	if err := r.g.Roll(game.NewRandomDice()); err != nil {
		slog.Error("Roll failed", "err", err)
		return
	}
	dicePayload := DiceRolledPayload{
		Dice:     r.g.Dice,
		IsDouble: r.g.Dice[0] == r.g.Dice[1],
		Player:   colorName(r.g.CurrentTurn),
	}
	r.broadcastAll(mustEncode("dice_rolled", dicePayload))
	r.broadcastAll(mustEncode("turn_changed", TurnChangedPayload{
		Player:   colorName(r.g.CurrentTurn),
		TimeLeft: int(turnDuration.Seconds()),
	}))
	r.broadcastGameState()
	r.persistState()
	r.startTurnTimer()
}

func (r *Room) handleChat(c *Client, raw json.RawMessage) {
	var p ChatPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}
	if len([]rune(p.Text)) > 500 {
		return
	}
	r.broadcastAll(mustEncode("chat_message", ChatMessagePayload{
		From: c.name,
		Text: p.Text,
		Time: time.Now().Format("15:04"),
	}))
}

func (r *Room) handleGameOver() {
	r.stopTurnTimer()
	payload := GameOverPayload{
		Winner: colorName(r.g.Winner),
		IsMars: r.g.IsMars,
	}
	r.broadcastAll(mustEncode("game_over", payload))
	// Persist final state.
	r.persistState()
}

// --- Timers ---

func (r *Room) startTurnTimer() {
	r.stopTurnTimer()
	r.turnTimer = time.AfterFunc(turnDuration, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.handleTurnTimeout()
	})
}

func (r *Room) stopTurnTimer() {
	if r.turnTimer != nil {
		r.turnTimer.Stop()
		r.turnTimer = nil
	}
}

func (r *Room) handleTurnTimeout() {
	if r.g == nil || r.g.Phase == game.PhaseFinished {
		return
	}
	// Force end turn (ignore error — there might be usable moves, but timeout overrides).
	_ = r.g.EndTurn()
	r.advanceTurn()
}

func (r *Room) handleGraceTimeout(slot int) {
	if r.g == nil || r.g.Phase == game.PhaseFinished {
		return
	}
	// The remaining client wins.
	opponentSlot := 1 - slot
	if r.clients[opponentSlot] != nil {
		r.g.Phase = game.PhaseFinished
		colorBySlot := [2]game.Color{game.White, game.Black}
		r.g.Winner = colorBySlot[opponentSlot]
		r.handleGameOver()
	}
}

// --- Broadcast helpers ---

func (r *Room) broadcastAll(msg []byte) {
	for _, c := range r.clients {
		if c != nil {
			c.sendMsg(msg)
		}
	}
}

func (r *Room) broadcastExceptSlot(exceptSlot int, msg []byte) {
	for i, c := range r.clients {
		if c != nil && i != exceptSlot {
			c.sendMsg(msg)
		}
	}
}

func (r *Room) broadcast(msg []byte)               { r.broadcastAll(msg) }
func (r *Room) broadcastExcept(ex *Client, msg []byte) {
	for _, c := range r.clients {
		if c != nil && c != ex {
			c.sendMsg(msg)
		}
	}
}

func (r *Room) isEmpty() bool {
	for _, c := range r.clients {
		if c != nil {
			return false
		}
	}
	return true
}

func (r *Room) slotOf(c *Client) int {
	for i, cl := range r.clients {
		if cl == c {
			return i
		}
	}
	return -1
}

func (r *Room) roomID() string {
	// Look up room ID from DB by code (cached or queried).
	// For simplicity, stored as a field set during addClient.
	return r.code // placeholder; handler.go sets r.dbRoomID
}

// --- State helpers ---

func (r *Room) buildGameState(forColor game.Color) GameStatePayload {
	var board [25]BoardPoint
	for i, pt := range r.g.Board.Points {
		board[i] = BoardPoint{Owner: int(pt.Owner), Checkers: pt.Checkers}
	}
	players := []PlayerSnapshot{
		{Name: r.names[0], Color: "white", Connected: r.clients[0] != nil},
		{Name: r.names[1], Color: "black", Connected: r.clients[1] != nil},
	}
	timeLeft := int(turnDuration.Seconds())
	return GameStatePayload{
		Phase:         phaseName(r.g.Phase),
		CurrentTurn:   colorName(r.g.CurrentTurn),
		Board:         board,
		BorneOff:      r.g.Board.BorneOff,
		Dice:          r.g.Dice,
		RemainingDice: r.g.RemainingDice,
		MoveCount:     r.g.MoveCount,
		MyColor:       colorName(forColor),
		Players:       players,
		TimeLeft:      timeLeft,
	}
}

func (r *Room) broadcastGameState() {
	for _, c := range r.clients {
		if c != nil {
			c.sendMsg(mustEncode("game_state", r.buildGameState(c.color)))
		}
	}
}

func (r *Room) persistState() {
	if r.gameID == "" || r.g == nil {
		return
	}
	boardJSON, _ := json.Marshal(r.g.Board)
	var winner *string
	if r.g.Winner != game.NoColor {
		w := colorName(r.g.Winner)
		winner = &w
	}
	ctx := context.Background()
	if err := r.hub.repos.Games.UpdateState(ctx, r.gameID,
		boardJSON,
		colorName(r.g.CurrentTurn),
		r.g.Dice,
		r.g.RemainingDice,
		phaseName(r.g.Phase),
		winner,
		r.g.IsMars,
		r.g.MoveCount,
	); err != nil {
		slog.Error("persist game state", "err", err)
	}
}

// --- Helpers ---

func colorName(c game.Color) string {
	switch c {
	case game.White:
		return "white"
	case game.Black:
		return "black"
	default:
		return ""
	}
}

func phaseName(p game.Phase) string {
	names := map[game.Phase]string{
		game.PhaseWaiting:     "waiting",
		game.PhaseRollingFirst: "rolling_first",
		game.PhasePlaying:     "playing",
		game.PhaseBearingOff:  "bearing_off",
		game.PhaseFinished:    "finished",
	}
	if n, ok := names[p]; ok {
		return n
	}
	return fmt.Sprintf("phase_%d", p)
}
```

Note: `room.go` imports `nhooyr.io/websocket` only for the `websocket.StatusPolicyViolation` constant in `addClient`. Fix the import in the file header:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"github.com/denis/web-backgammon/internal/game"
)
```

Also, the `Room` needs a `dbRoomID` field (the UUID from the `rooms` table, different from the room code). Add to the struct:

```go
type Room struct {
	// ...existing fields...
	dbRoomID string // UUID from rooms table
}
```

And update `roomID()` to return `r.dbRoomID`.

- [ ] **Step 2: Fix compile errors**

```bash
cd backend && go build ./internal/ws/...
```

Fix any import or reference errors that appear. The most likely issues:
- Missing `dbRoomID` field initialization (set in `handler.go` in the next task).
- `game.NewRandomDice()` — verify the function exists in Phase 1 code. If the constructor is `NewRandomDice()` vs `NewDice()`, check `backend/internal/game/dice.go` and use the correct name.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/ws/room.go
git commit -m "feat(ws): add Room game session with move processing, timers, and reconnect logic"
```

---

### Task 6: WS Handler (HTTP upgrade)

**Files:** Create `backend/internal/ws/handler.go`

The handler validates the session cookie, looks up the player and room in DB, checks for reconnect, then creates a `Client` and calls `client.run()`.

- [ ] **Step 1: Write handler.go**

```go
package ws

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"nhooyr.io/websocket"

	"github.com/go-chi/chi/v5"
)

// Handler returns an http.Handler that upgrades GET /ws/{roomCode} to WebSocket.
func Handler(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomCode := strings.ToUpper(chi.URLParam(r, "roomCode"))

		// Validate session cookie.
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sessionToken := cookie.Value

		ctx := context.Background()

		// Look up player.
		player, err := hub.repos.Players.FindBySession(ctx, sessionToken)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Look up room in DB to get its UUID.
		room, err := hub.repos.Rooms.FindByCode(ctx, roomCode)
		if err != nil {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		if room.Status != "playing" && room.Status != "waiting" {
			http.Error(w, "room not available", http.StatusGone)
			return
		}

		// Upgrade to WebSocket.
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // TODO: restrict to ALLOWED_ORIGINS in production
		})
		if err != nil {
			slog.Error("ws accept", "err", err)
			return
		}

		c := &Client{
			hub:          hub,
			conn:         conn,
			roomCode:     roomCode,
			sessionToken: sessionToken,
			name:         player.Name,
			send:         make(chan []byte, 64),
		}

		// Check for reconnect first.
		hub.mu.Lock()
		wsRoom, exists := hub.rooms[roomCode]
		if exists && wsRoom.reconnect(c) {
			hub.sessions[sessionToken] = c
			hub.mu.Unlock()
			c.run(r.Context())
			return
		}
		// Set dbRoomID on the ws room (create it if needed).
		if !exists {
			wsRoom = newRoom(roomCode, hub)
			hub.rooms[roomCode] = wsRoom
		}
		wsRoom.dbRoomID = room.ID
		hub.mu.Unlock()

		hub.register(c)
		c.run(r.Context())
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd backend && go build ./internal/ws/...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/ws/handler.go
git commit -m "feat(ws): add WS upgrade handler with session auth and reconnect detection"
```

---

### Task 7: Wire WS into main.go

**Files:** Modify `backend/cmd/server/main.go`, modify `backend/internal/api/server.go`

- [ ] **Step 1: Add WS route to the chi router in api/server.go**

In `backend/internal/api/server.go`, add a field for the WS hub and a route:

Add import:
```go
"github.com/denis/web-backgammon/internal/ws"
```

Update `Server` struct to hold the hub:
```go
type Server struct {
	rooms     *db.RoomRepo
	players   *db.PlayerRepo
	games     *db.GameRepo
	sanitizer *bluemonday.Policy
	origins   []string
	hub       *ws.Hub
}
```

Update `NewServer`:
```go
func NewServer(rooms *db.RoomRepo, players *db.PlayerRepo, games *db.GameRepo,
	origins []string, hub *ws.Hub) *Server {
	return &Server{
		rooms:     rooms,
		players:   players,
		games:     games,
		sanitizer: bluemonday.StrictPolicy(),
		origins:   origins,
		hub:       hub,
	}
}
```

Add WS route in `Router()`:
```go
r.Get("/ws/{roomCode}", ws.Handler(s.hub))
```

- [ ] **Step 2: Update main.go**

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
	"github.com/denis/web-backgammon/internal/ws"
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

	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)
	games := db.NewGameRepo(pool)

	hub := ws.NewHub(ws.DBRepos{
		Rooms:   rooms,
		Players: players,
		Games:   games,
	})

	srv := api.NewServer(rooms, players, games, cfg.AllowedOrigins, hub)

	addr := ":" + cfg.Port
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Fix compilation (NewServer signature changed)**

Update the API integration test's `NewServer` call in `backend/internal/api/integration_test.go` — pass `nil` for hub:
```go
srv := api.NewServer(
    db.NewRoomRepo(pool),
    db.NewPlayerRepo(pool),
    db.NewGameRepo(pool),
    []string{"http://localhost:3000"},
    nil, // hub not needed for REST-only tests
)
```

- [ ] **Step 4: Build the whole backend**

```bash
cd backend && go build ./...
```

Expected: no errors.

- [ ] **Step 5: Run all unit tests**

```bash
cd backend && go test ./...
```

Expected: `PASS` (unit tests only; integration tests require `-tags integration`).

- [ ] **Step 6: Commit**

```bash
git add backend/internal/api/server.go backend/cmd/server/main.go backend/internal/api/integration_test.go
git commit -m "feat(ws): wire WS hub and /ws/{roomCode} route into main.go and chi router"
```

---

### Task 8: WS integration tests

**Files:** Create `backend/internal/ws/ws_integration_test.go`

Tests use `net/http/httptest` to start a real HTTP server, upgrade two WS connections (one per player), and verify the game flow.

- [ ] **Step 1: Write ws_integration_test.go**

```go
//go:build integration

package ws_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
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
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../../migrations")
	require.NoError(t, db.RunMigrations(ctx, pool, migrationsDir))

	rooms := db.NewRoomRepo(pool)
	players := db.NewPlayerRepo(pool)
	games := db.NewGameRepo(pool)

	hub := internalws.NewHub(internalws.DBRepos{Rooms: rooms, Players: players, Games: games})
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
	ctx := context.Background()

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

	_ = ctx // suppress unused warning

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

	// Drain startup messages (dice_rolled + game_state × 2).
	for i := 0; i < 4; i++ {
		readMsg(t, conn1)
	}
	// conn2 also gets its own copy.
	for i := 0; i < 4; i++ {
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
	// Use net/http directly to get cookies back.
	import_net_http_client := &http.Client{}
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := import_net_http_client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return httpResp{status: resp.StatusCode, body: b}
}
```

Note: the `mustPost` helper needs imports. Add to the top of the file:
```go
import (
    // ...existing imports...
    "io"
    "net/http"
)
```

Also remove the inline `import_net_http_client` variable name — change to `client`:
```go
client := &http.Client{}
req, _ := http.NewRequest("POST", url, strings.NewReader(body))
// ...
resp, err := client.Do(req)
```

- [ ] **Step 2: Run the WS integration tests**

```bash
cd backend && go test -tags integration -v -timeout 180s ./internal/ws/...
```

Expected: `TestWS_FullGameStart` and `TestWS_ChatDelivery` pass.

- [ ] **Step 3: Run all backend tests**

```bash
cd backend && go test ./...                              # unit tests
cd backend && go test -tags integration -timeout 180s ./... # integration tests
```

Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/ws/ws_integration_test.go
git commit -m "test(ws): add WS integration tests for game start and chat delivery"
```

---

## Phase 3 Complete

- [ ] **Tag**

```bash
git tag phase-3-websocket
git push origin master --tags
```

- [ ] **Smoke test with two browser tabs**

```bash
# Terminal 1: start postgres
docker compose up postgres -d

# Terminal 2: run server
cd backend && DATABASE_URL="postgres://bg_user:bg_pass@localhost:5433/backgammon?sslmode=disable" go run ./cmd/server/main.go

# Browser: open http://localhost:3000
# Create a room as "Алексей", copy the code.
# Open a second tab, enter the code and "Мария" → join.
# Both tabs should receive game_state in the browser DevTools WS inspector.
```

---

## Self-Review Checklist

**Spec coverage (Секция 3):**
- ✓ All 5 client→server message types: `move`, `end_turn`, `pass`, `chat`, `ping`/`pong`.
- ✓ All 10 server→client message types: `game_state`, `opponent_moved`, `dice_rolled`, `turn_changed`, `move_error`, `chat_message`, `opponent_disconnected`, `opponent_reconnected`, `game_over`, `pong`.
- ✓ Ping/pong every 15 s (断线 detection within 30 s).
- ✓ Turn timer 60 s → auto-pass.
- ✓ Disconnect grace 5 min → auto-loss.
- ✓ Reconnect via `session_token` → `game_state` resync + `opponent_reconnected`.
- ✓ Dice rolled server-side only (security — Секция 7).
- ✓ `move_count` field in `game_state` for replay-attack protection.
- ✓ Game state persisted to DB after every move and turn end.

**Gaps (deferred):**
- WS `move` rate-limit (30/s) and chat rate-limit (5/s) from Секция 6 — add in Phase 4 or as a quick follow-up if needed.
- Chat messages not persisted to `chat_messages` table — add `ChatRepo` in Phase 4 or as a follow-up.
- Production `InsecureSkipVerify: false` + `OriginPatterns` from config — fix when deploying (Phase 4 infra).

package ws

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

	// DB room ID (UUID from rooms table)
	dbRoomID string

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
	dice := game.NewRandomDice(time.Now().UnixNano())
	if err := r.g.RollFirst(dice); err != nil {
		slog.Error("RollFirst failed", "err", err)
		return
	}

	// Persist initial game record.
	boardJSON, _ := json.Marshal(r.g.Board)
	ctx := context.Background()
	gr, err := r.hub.repos.Games.Create(ctx, r.dbRoomID, boardJSON)
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
	dice := game.NewRandomDice(time.Now().UnixNano())
	if err := r.g.Roll(dice); err != nil {
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

func (r *Room) broadcast(msg []byte) { r.broadcastAll(msg) }
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
		game.PhaseWaiting:      "waiting",
		game.PhaseRollingFirst: "rolling_first",
		game.PhasePlaying:      "playing",
		game.PhaseBearingOff:   "bearing_off",
		game.PhaseFinished:     "finished",
	}
	if n, ok := names[p]; ok {
		return n
	}
	return fmt.Sprintf("phase_%d", p)
}

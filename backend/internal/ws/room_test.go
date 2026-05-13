package ws

import (
	"encoding/json"
	"testing"

	"github.com/denis/web-backgammon/internal/db"
	"github.com/denis/web-backgammon/internal/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomHandleTurnTimeoutKeepsDiceWhenMovesRemain(t *testing.T) {
	g := game.NewGame()
	require.NoError(t, g.RollFirst(game.NewFixedDice([][2]int{{4, 1}})))

	originalTurn := g.CurrentTurn
	originalDice := append([]int(nil), g.Dice...)
	originalRemaining := append([]int(nil), g.RemainingDice...)

	r := newRoom("TESTROOM", nil)
	r.g = g

	r.handleTurnTimeout()

	assert.Equal(t, originalTurn, g.CurrentTurn)
	assert.Equal(t, originalDice, g.Dice)
	assert.Equal(t, originalRemaining, g.RemainingDice)
}

func TestRoomStartGameIsIdempotent(t *testing.T) {
	g := game.NewGame()
	require.NoError(t, g.RollFirst(game.NewFixedDice([][2]int{{4, 1}})))

	originalTurn := g.CurrentTurn
	originalDice := append([]int(nil), g.Dice...)
	originalRemaining := append([]int(nil), g.RemainingDice...)

	r := newRoom("TESTROOM", nil)
	r.g = g

	r.startGame()

	assert.Same(t, g, r.g)
	assert.Equal(t, originalTurn, r.g.CurrentTurn)
	assert.Equal(t, originalDice, r.g.Dice)
	assert.Equal(t, originalRemaining, r.g.RemainingDice)
}

func TestRoomBuildGameStateIncludesOnlyLegalFirstMovesFromMaximalSequences(t *testing.T) {
	b := &game.Board{}
	b.Points[14] = game.Point{Owner: game.White, Checkers: 1}
	b.Points[6] = game.Point{Owner: game.Black, Checkers: 1}

	r := newRoom("TESTROOM", nil)
	r.g = &game.Game{
		Board:         b,
		CurrentTurn:   game.White,
		Dice:          []int{3, 5},
		RemainingDice: []int{3, 5},
		Phase:         game.PhasePlaying,
	}

	state := r.buildGameState(game.White)

	assert.Equal(t, []MovePayload{{From: 14, To: 9, Die: 5}}, state.LegalMoves)
}

func TestGameFromRecordRestoresPersistedState(t *testing.T) {
	b := &game.Board{}
	b.Points[22] = game.Point{Owner: game.White, Checkers: 1}
	boardJSON, err := json.Marshal(b)
	require.NoError(t, err)
	currentTurn := "white"

	restored, err := gameFromRecord(&db.GameRecord{
		ID:            "game-id",
		BoardState:    boardJSON,
		CurrentTurn:   &currentTurn,
		Dice:          []int{2, 3},
		RemainingDice: []int{3},
		Phase:         "playing",
		MoveCount:     1,
	})

	require.NoError(t, err)
	assert.Equal(t, game.White, restored.CurrentTurn)
	assert.Equal(t, []int{2, 3}, restored.Dice)
	assert.Equal(t, []int{3}, restored.RemainingDice)
	assert.Equal(t, game.PhasePlaying, restored.Phase)
	assert.Equal(t, 1, restored.MoveCount)
	assert.Equal(t, game.White, restored.Board.Points[22].Owner)
	assert.Equal(t, 1, restored.Board.Points[22].Checkers)
}

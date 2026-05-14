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

func TestRoomBuildGameStateIncludesLargerDieWhenOnlyOneMoveUsable(t *testing.T) {
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

func TestRoomBuildGameStateIncludesCompoundMoveTargets(t *testing.T) {
	b := &game.Board{}
	b.Points[24] = game.Point{Owner: game.White, Checkers: 1}

	r := newRoom("TESTROOM", nil)
	r.g = &game.Game{
		Board:         b,
		CurrentTurn:   game.White,
		Dice:          []int{1, 5},
		RemainingDice: []int{1, 5},
		Phase:         game.PhasePlaying,
	}

	state := r.buildGameState(game.White)

	assert.Contains(t, state.LegalMoves, MovePayload{From: 24, To: 23, Die: 1})
	assert.Contains(t, state.LegalMoves, MovePayload{From: 24, To: 19, Die: 5})
	assert.Contains(t, state.LegalMoves, MovePayload{
		From: 24,
		To:   18,
		Die:  6,
		Steps: []MovePayload{
			{From: 24, To: 23, Die: 1},
			{From: 23, To: 18, Die: 5},
		},
	})
}

func TestRoomHandleMoveAdvancesTurnAfterFullMove(t *testing.T) {
	r := newRoom("TESTROOM", nil)
	r.g = game.NewGame()
	r.g.Phase = game.PhasePlaying
	r.g.CurrentTurn = game.White
	r.g.Dice = []int{6}
	r.g.RemainingDice = []int{6}

	c := &Client{color: game.White}
	raw, err := json.Marshal(MovePayload{From: 24, To: 18, Die: 6})
	require.NoError(t, err)

	r.handleMove(c, raw)

	assert.Equal(t, game.Black, r.g.CurrentTurn)
	assert.Len(t, r.g.Dice, 2)
	assert.NotEmpty(t, r.g.RemainingDice)
}

func TestGameFromRecordRestoresPersistedState(t *testing.T) {
	b := &game.Board{}
	b.Points[22] = game.Point{Owner: game.White, Checkers: 1}
	boardJSON, err := json.Marshal(b)
	require.NoError(t, err)
	currentTurn := "white"

	restored, err := gameFromRecord(&db.GameRecord{
		ID:              "game-id",
		BoardState:      boardJSON,
		CurrentTurn:     &currentTurn,
		Dice:            []int{2, 3},
		RemainingDice:   []int{3},
		Phase:           "playing",
		MoveCount:       1,
		HeadMovesWhite:  1,
		HeadMovesBlack:  0,
		TurnsWhite:      2,
		TurnsBlack:      3,
	})

	require.NoError(t, err)
	assert.Equal(t, game.White, restored.CurrentTurn)
	assert.Equal(t, []int{2, 3}, restored.Dice)
	assert.Equal(t, []int{3}, restored.RemainingDice)
	assert.Equal(t, game.PhasePlaying, restored.Phase)
	assert.Equal(t, 1, restored.MoveCount)
	assert.Equal(t, game.White, restored.Board.Points[22].Owner)
	assert.Equal(t, 1, restored.Board.Points[22].Checkers)
	assert.Equal(t, 1, restored.HeadMovesThisTurn[game.White])
	assert.Equal(t, 0, restored.HeadMovesThisTurn[game.Black])
	assert.Equal(t, 2, restored.TurnsCompleted[game.White])
	assert.Equal(t, 3, restored.TurnsCompleted[game.Black])
}

func TestGameFromRecordRestoresResult(t *testing.T) {
	b := &game.Board{}
	boardJSON, err := json.Marshal(b)
	require.NoError(t, err)
	currentTurn := "white"
	resultType := "koks"
	resultPoints := 3

	restored, err := gameFromRecord(&db.GameRecord{
		ID:            "game-id",
		BoardState:    boardJSON,
		CurrentTurn:   &currentTurn,
		Dice:          []int{},
		RemainingDice: []int{},
		Phase:         "finished",
		MoveCount:     42,
		ResultType:    &resultType,
		ResultPoints:  &resultPoints,
	})

	require.NoError(t, err)
	assert.Equal(t, game.ResultKoks, restored.Result)
	assert.Equal(t, 3, restored.ResultPoints)
}

func TestGameFromRecordRestoresBearingOffFlags(t *testing.T) {
	// White has all checkers in home (1-6), black does not.
	b := &game.Board{}
	b.Points[1] = game.Point{Owner: game.White, Checkers: 15}
	b.Points[10] = game.Point{Owner: game.Black, Checkers: 15}
	boardJSON, err := json.Marshal(b)
	require.NoError(t, err)
	currentTurn := "white"

	restored, err := gameFromRecord(&db.GameRecord{
		ID:            "game-id",
		BoardState:    boardJSON,
		CurrentTurn:   &currentTurn,
		Dice:          []int{1, 2},
		RemainingDice: []int{1, 2},
		Phase:         "bearing_off",
		MoveCount:     10,
	})

	require.NoError(t, err)
	assert.True(t, restored.BearingOff[game.White], "white should be in bearing-off")
	assert.False(t, restored.BearingOff[game.Black], "black should not be in bearing-off")
}

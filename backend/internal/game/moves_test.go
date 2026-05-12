package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSingleMoves_InitialBoardWhite(t *testing.T) {
	b := NewBoard()
	moves := GenerateSingleMoves(b, White, 3)
	assert.Contains(t, moves, Move{From: 24, To: 21, Die: 3})
}

func TestGenerateSingleMoves_NoMoveForOccupiedByOpponent(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	b.Points[4] = Point{Owner: Black, Checkers: 1}
	moves := GenerateSingleMoves(b, White, 6)
	assert.Empty(t, moves, "destination 4 is occupied by black")
}

func TestGenerateSingleMoves_MultipleSources(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 1}
	moves := GenerateSingleMoves(b, White, 4)
	assert.Contains(t, moves, Move{From: 24, To: 20, Die: 4})
	assert.Contains(t, moves, Move{From: 20, To: 16, Die: 4})
}

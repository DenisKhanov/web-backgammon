package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidMove_WhiteCorrectDirection(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_WhiteWrongDirection(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	ok, err := IsValidMove(b, White, Move{From: 10, To: 16, Die: 6})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestIsValidMove_BlackCorrectDirection(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, Black, Move{From: 1, To: 7, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_NoCheckerAtSource(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 13, To: 7, Die: 6})
	assert.False(t, ok)
}

func TestIsValidMove_OpponentAtSource(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 1, To: 0, Die: 1})
	assert.False(t, ok)
}

func TestIsValidMove_DestinationOccupiedByOpponent(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 15}
	b.Points[18] = Point{Owner: Black, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.False(t, ok, "cannot land on opponent's point (no hitting in long backgammon)")
}

func TestIsValidMove_DestinationOwnColorAllowed(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 14}
	b.Points[18] = Point{Owner: White, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 6})
	assert.True(t, ok)
}

func TestIsValidMove_DieMismatch(t *testing.T) {
	b := NewBoard()
	ok, _ := IsValidMove(b, White, Move{From: 24, To: 18, Die: 5})
	assert.False(t, ok)
}

func TestIsValidMove_OutOfBoundsForRegularMove(t *testing.T) {
	b := &Board{}
	b.Points[3] = Point{Owner: White, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 3, To: -2, Die: 5})
	assert.False(t, ok)
}

func TestIsValidMove_GlukhoiZabor_ClosingSixthBlocked(t *testing.T) {
	b := &Board{}
	b.Points[17] = Point{Owner: White, Checkers: 2}
	b.Points[18] = Point{Owner: White, Checkers: 2}
	b.Points[19] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 2}
	b.Points[21] = Point{Owner: White, Checkers: 2}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	b.Points[23] = Point{Owner: Black, Checkers: 1}

	ok, _ := IsValidMove(b, White, Move{From: 24, To: 22, Die: 2})
	assert.False(t, ok, "closing 6th consecutive point ahead of opponent must be blocked")
}

func TestIsValidMove_GlukhoiZabor_AllowedNoOpponentAhead(t *testing.T) {
	b := &Board{}
	b.Points[17] = Point{Owner: White, Checkers: 2}
	b.Points[18] = Point{Owner: White, Checkers: 2}
	b.Points[19] = Point{Owner: White, Checkers: 2}
	b.Points[20] = Point{Owner: White, Checkers: 2}
	b.Points[21] = Point{Owner: White, Checkers: 2}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	b.Points[15] = Point{Owner: Black, Checkers: 1}

	ok, _ := IsValidMove(b, White, Move{From: 24, To: 22, Die: 2})
	assert.True(t, ok, "6 in a row is allowed if no opponent checker is ahead of the wall")
}

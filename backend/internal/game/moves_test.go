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

func TestGenerateSequences_TwoDistinctDice(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 2}
	dice := []int{3, 5}

	sequences := GenerateSequences(b, White, dice)

	assert.NotEmpty(t, sequences)
	hasOrder35 := false
	hasOrder53 := false
	for _, seq := range sequences {
		if len(seq) == 2 && seq[0].Die == 3 && seq[1].Die == 5 {
			hasOrder35 = true
		}
		if len(seq) == 2 && seq[0].Die == 5 && seq[1].Die == 3 {
			hasOrder53 = true
		}
	}
	assert.True(t, hasOrder35, "expect a sequence using die 3 then 5")
	assert.True(t, hasOrder53, "expect a sequence using die 5 then 3")
}

func TestGenerateSequences_Double(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 4}
	dice := ExpandDice(2, 2)

	sequences := GenerateSequences(b, White, dice)

	found4 := false
	for _, seq := range sequences {
		if len(seq) == 4 {
			found4 = true
		}
	}
	assert.True(t, found4, "double 2 must allow a 4-move sequence")
}

func TestGenerateSequences_NoMovesAvailable(t *testing.T) {
	b := &Board{}
	b.Points[10] = Point{Owner: White, Checkers: 1}
	b.Points[10-3] = Point{Owner: Black, Checkers: 1}
	b.Points[10-5] = Point{Owner: Black, Checkers: 1}

	sequences := GenerateSequences(b, White, []int{3, 5})

	for _, seq := range sequences {
		assert.Empty(t, seq, "no moves should be possible")
	}
}

func TestGenerateSequences_MustUseLargerDie(t *testing.T) {
	// Single checker at 8. Die=3→5, die=5→3 (both 1-move sequences since
	// bear-off is not yet enabled). Larger-die rule must keep only die=5.
	b := &Board{}
	b.Points[8] = Point{Owner: White, Checkers: 1}

	sequences := GenerateSequences(b, White, []int{3, 5})

	assert.NotEmpty(t, sequences)
	for _, seq := range sequences {
		assert.Equal(t, 1, len(seq))
		assert.Equal(t, 5, seq[0].Die, "must use the larger die when only one is usable")
	}
}

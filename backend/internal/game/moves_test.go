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

func TestGenerateSingleMoves_BearOff_Exact(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 15}
	moves := GenerateSingleMoves(b, White, 6)
	hasBearOff := false
	for _, m := range moves {
		if m.From == 6 && m.To == 0 {
			hasBearOff = true
		}
	}
	assert.True(t, hasBearOff)
}

func TestGenerateSingleMoves_BearOff_FallbackHigher(t *testing.T) {
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 15}
	moves := GenerateSingleMoves(b, White, 6)
	hasBearOff := false
	for _, m := range moves {
		if m.From == 4 && m.To == 0 {
			hasBearOff = true
		}
	}
	assert.True(t, hasBearOff, "die=6 > top=4, fallback bear-off from 4 must be available")
}

func TestGenerateSingleMoves_BearOff_FallbackBlocked(t *testing.T) {
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 1}
	b.Points[5] = Point{Owner: White, Checkers: 14}
	moves := GenerateSingleMoves(b, White, 6)
	for _, m := range moves {
		assert.False(t, m.From == 4 && m.To == 0, "must not bear off from 4 when 5 has checkers")
	}
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
	// Checker at 14, opponent at 6 blocks both second moves:
	//   die=3: 14→11, then die=5: 11→6 blocked → length 1
	//   die=5: 14→9,  then die=3:  9→6 blocked → length 1
	// Both sequences have length 1; larger-die rule must keep only die=5.
	b := &Board{}
	b.Points[14] = Point{Owner: White, Checkers: 1}
	b.Points[6] = Point{Owner: Black, Checkers: 1}

	sequences := GenerateSequences(b, White, []int{3, 5})

	assert.NotEmpty(t, sequences)
	for _, seq := range sequences {
		assert.Equal(t, 1, len(seq))
		assert.Equal(t, 5, seq[0].Die, "must use the larger die when only one is usable")
	}
}

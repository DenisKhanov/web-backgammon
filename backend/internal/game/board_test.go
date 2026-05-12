package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBoard_InitialSetup(t *testing.T) {
	b := NewBoard()

	assert.Equal(t, 15, b.Points[24].Checkers, "white starts with 15 checkers on point 24")
	assert.Equal(t, White, b.Points[24].Owner)

	assert.Equal(t, 15, b.Points[1].Checkers, "black starts with 15 checkers on point 1")
	assert.Equal(t, Black, b.Points[1].Owner)

	for i := 2; i <= 23; i++ {
		assert.Equal(t, 0, b.Points[i].Checkers, "point %d must be empty", i)
		assert.Equal(t, NoColor, b.Points[i].Owner, "point %d must have no owner", i)
	}

	assert.Equal(t, 0, b.BorneOff[White])
	assert.Equal(t, 0, b.BorneOff[Black])
}

func TestBoard_CountCheckers(t *testing.T) {
	b := NewBoard()
	assert.Equal(t, 15, b.CountCheckers(White))
	assert.Equal(t, 15, b.CountCheckers(Black))
}

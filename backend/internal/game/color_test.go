package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColor_Opponent(t *testing.T) {
	assert.Equal(t, Black, White.Opponent())
	assert.Equal(t, White, Black.Opponent())
}

func TestColor_Direction(t *testing.T) {
	assert.Equal(t, -1, White.Direction(), "white moves from 24 toward 1")
	assert.Equal(t, +1, Black.Direction(), "black moves from 1 toward 24")
}

func TestColor_HomeRange(t *testing.T) {
	lo, hi := White.HomeRange()
	assert.Equal(t, 1, lo)
	assert.Equal(t, 6, hi)
	lo, hi = Black.HomeRange()
	assert.Equal(t, 19, lo)
	assert.Equal(t, 24, hi)
}

func TestColor_StartPoint(t *testing.T) {
	assert.Equal(t, 24, White.StartPoint())
	assert.Equal(t, 1, Black.StartPoint())
}

func TestColor_BearOffTarget(t *testing.T) {
	assert.Equal(t, 0, White.BearOffTarget())
	assert.Equal(t, 25, Black.BearOffTarget())
}

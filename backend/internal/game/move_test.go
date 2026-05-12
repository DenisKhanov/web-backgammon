package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove_Apply_WhiteRegular(t *testing.T) {
	b := NewBoard()
	m := Move{From: 24, To: 18, Die: 6}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 14, b.Points[24].Checkers)
	assert.Equal(t, White, b.Points[24].Owner)
	assert.Equal(t, 1, b.Points[18].Checkers)
	assert.Equal(t, White, b.Points[18].Owner)
}

func TestMove_Apply_BlackRegular(t *testing.T) {
	b := NewBoard()
	m := Move{From: 1, To: 5, Die: 4}

	err := m.Apply(b, Black)

	assert.NoError(t, err)
	assert.Equal(t, 14, b.Points[1].Checkers)
	assert.Equal(t, 1, b.Points[5].Checkers)
	assert.Equal(t, Black, b.Points[5].Owner)
}

func TestMove_Apply_EmptiesSourcePointOwner(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: White, Checkers: 1}
	m := Move{From: 24, To: 23, Die: 1}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[24].Checkers)
	assert.Equal(t, NoColor, b.Points[24].Owner, "empty point must reset its owner")
}

func TestMove_Apply_BearOffWhite(t *testing.T) {
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 1}
	m := Move{From: 5, To: 0, Die: 5}

	err := m.Apply(b, White)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[5].Checkers)
	assert.Equal(t, 1, b.BorneOff[White])
}

func TestMove_Apply_BearOffBlack(t *testing.T) {
	b := &Board{}
	b.Points[22] = Point{Owner: Black, Checkers: 1}
	m := Move{From: 22, To: 25, Die: 3}

	err := m.Apply(b, Black)

	assert.NoError(t, err)
	assert.Equal(t, 0, b.Points[22].Checkers)
	assert.Equal(t, 1, b.BorneOff[Black])
}

func TestMove_Apply_NoCheckerAtSource(t *testing.T) {
	b := NewBoard()
	m := Move{From: 10, To: 4, Die: 6}

	err := m.Apply(b, White)

	assert.Error(t, err)
}

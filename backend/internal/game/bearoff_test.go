package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllInHome_White_True(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 15}
	assert.True(t, b.AllInHome(White))
}

func TestAllInHome_White_FalseOneOutside(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	b.Points[7] = Point{Owner: White, Checkers: 1}
	assert.False(t, b.AllInHome(White))
}

func TestAllInHome_White_TrueIncludingBorneOff(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 10}
	b.BorneOff[White] = 5
	assert.True(t, b.AllInHome(White))
}

func TestAllInHome_Black(t *testing.T) {
	b := &Board{}
	b.Points[24] = Point{Owner: Black, Checkers: 10}
	b.Points[19] = Point{Owner: Black, Checkers: 5}
	assert.True(t, b.AllInHome(Black))
}

func TestBearOff_ExactDie_White(t *testing.T) {
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 1}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	ok, _ := IsValidMove(b, White, Move{From: 5, To: 0, Die: 5})
	assert.True(t, ok)
}

func TestBearOff_HigherPointFallback_White(t *testing.T) {
	b := &Board{}
	b.Points[5] = Point{Owner: White, Checkers: 15}
	ok, _ := IsValidMove(b, White, Move{From: 5, To: 0, Die: 6})
	assert.True(t, ok, "if no checker at exact-die point, may bear off from the highest occupied point")
}

func TestBearOff_NotAllInHome_Rejected(t *testing.T) {
	b := &Board{}
	b.Points[6] = Point{Owner: White, Checkers: 14}
	b.Points[7] = Point{Owner: White, Checkers: 1}
	ok, _ := IsValidMove(b, White, Move{From: 6, To: 0, Die: 6})
	assert.False(t, ok)
}

func TestBearOff_HigherStillExists_FromExactRequired(t *testing.T) {
	b := &Board{}
	b.Points[4] = Point{Owner: White, Checkers: 1}
	b.Points[5] = Point{Owner: White, Checkers: 14}

	ok, _ := IsValidMove(b, White, Move{From: 3, To: 0, Die: 4})
	assert.False(t, ok, "no checker on point 3")

	ok, _ = IsValidMove(b, White, Move{From: 5, To: 0, Die: 4})
	assert.False(t, ok, "exact die=4 is occupied, cannot bear off from higher point")

	ok, _ = IsValidMove(b, White, Move{From: 4, To: 0, Die: 4})
	assert.True(t, ok)
}

func TestBearOff_Black_ExactDie(t *testing.T) {
	b := &Board{}
	b.Points[22] = Point{Owner: Black, Checkers: 1}
	b.Points[19] = Point{Owner: Black, Checkers: 14}
	ok, _ := IsValidMove(b, Black, Move{From: 22, To: 25, Die: 3})
	assert.True(t, ok)
}

package game

type Color int

const (
	NoColor Color = iota
	White
	Black
)

func (c Color) Opponent() Color {
	switch c {
	case White:
		return Black
	case Black:
		return White
	}
	return NoColor
}

func (c Color) Direction() int {
	switch c {
	case White:
		return -1
	case Black:
		return +1
	}
	return 0
}

func (c Color) HomeRange() (lo, hi int) {
	switch c {
	case White:
		return 1, 6
	case Black:
		return 19, 24
	}
	return 0, 0
}

func (c Color) StartPoint() int {
	switch c {
	case White:
		return 24
	case Black:
		return 1
	}
	return 0
}

func (c Color) BearOffTarget() int {
	switch c {
	case White:
		return 0
	case Black:
		return 25
	}
	return -1
}

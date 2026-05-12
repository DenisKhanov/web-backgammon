package game

import "fmt"

type Move struct {
	From int
	To   int
	Die  int
}

func (m Move) Apply(b *Board, c Color) error {
	src := &b.Points[m.From]
	if src.Owner != c || src.Checkers == 0 {
		return fmt.Errorf("no %v checker at point %d", c, m.From)
	}

	src.Checkers--
	if src.Checkers == 0 {
		src.Owner = NoColor
	}

	if m.To == c.BearOffTarget() {
		b.BorneOff[c]++
		return nil
	}

	dst := &b.Points[m.To]
	dst.Owner = c
	dst.Checkers++
	return nil
}

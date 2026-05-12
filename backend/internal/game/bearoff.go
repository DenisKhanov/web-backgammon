package game

func (b *Board) AllInHome(c Color) bool {
	lo, hi := c.HomeRange()
	for p := 1; p <= 24; p++ {
		if p >= lo && p <= hi {
			continue
		}
		if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
			return false
		}
	}
	return true
}

// pointForBearOffDie maps a die value to the board point for bear-off purposes.
// White: die=6 → point 6. Black: die=6 → point 19.
func pointForBearOffDie(c Color, die int) int {
	switch c {
	case White:
		return die
	case Black:
		return 25 - die
	}
	return 0
}

// highestOccupiedInHome returns the point farthest from bear-off still occupied
// by color c in its home. Returns 0 if none found.
func highestOccupiedInHome(b *Board, c Color) int {
	lo, hi := c.HomeRange()
	switch c {
	case White:
		for p := hi; p >= lo; p-- {
			if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
				return p
			}
		}
	case Black:
		for p := lo; p <= hi; p++ {
			if b.Points[p].Owner == c && b.Points[p].Checkers > 0 {
				return p
			}
		}
	}
	return 0
}

// distanceToBearOff returns the die value that would exactly bear off from p.
// White: point 1 → distance 1, point 6 → distance 6.
// Black: point 24 → distance 1, point 19 → distance 6.
func distanceToBearOff(c Color, p int) int {
	switch c {
	case White:
		return p
	case Black:
		return 25 - p
	}
	return 0
}

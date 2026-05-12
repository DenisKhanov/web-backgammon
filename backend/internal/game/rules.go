package game

import "fmt"

func IsValidMove(b *Board, c Color, m Move) (bool, error) {
	if m.Die < 1 || m.Die > 6 {
		return false, fmt.Errorf("die out of range: %d", m.Die)
	}
	if m.From < 1 || m.From > 24 {
		return false, fmt.Errorf("from out of range: %d", m.From)
	}

	src := b.Points[m.From]
	if src.Owner != c || src.Checkers == 0 {
		return false, fmt.Errorf("no %v checker at %d", c, m.From)
	}

	// Bear-off must be checked before direction/expectedTo since fallback
	// bear-offs have from-die*dir != BearOffTarget.
	if m.To == c.BearOffTarget() {
		return isValidBearOff(b, c, m)
	}

	expectedTo := m.From + c.Direction()*m.Die
	if expectedTo != m.To {
		return false, fmt.Errorf("die %d from %d for %v leads to %d, not %d",
			m.Die, m.From, c, expectedTo, m.To)
	}

	if m.To < 1 || m.To > 24 {
		return false, fmt.Errorf("to out of range for regular move: %d", m.To)
	}

	dst := b.Points[m.To]
	if dst.Owner != NoColor && dst.Owner != c {
		return false, fmt.Errorf("point %d is occupied by opponent", m.To)
	}

	if wouldCreateGlukhoiZabor(b, c, m) {
		return false, fmt.Errorf("move would create a 6+ block ahead of opponent (glukhoi zabor)")
	}

	return true, nil
}

func isValidBearOff(b *Board, c Color, m Move) (bool, error) {
	if !b.AllInHome(c) {
		return false, fmt.Errorf("cannot bear off: not all checkers are in home")
	}

	srcDist := distanceToBearOff(c, m.From)

	if srcDist == m.Die {
		return true, nil
	}

	if m.Die > srcDist {
		highest := highestOccupiedInHome(b, c)
		if distanceToBearOff(c, highest) == srcDist {
			return true, nil
		}
		return false, fmt.Errorf("bear-off from %d with die %d not allowed: higher checker on %d", m.From, m.Die, highest)
	}

	return false, fmt.Errorf("invalid bear-off attempt from %d with die %d", m.From, m.Die)
}

// wouldCreateGlukhoiZabor моделирует ход и проверяет, образуется ли непрерывный
// ряд из 6+ пунктов цвета c с шашкой соперника впереди этого ряда.
func wouldCreateGlukhoiZabor(b *Board, c Color, m Move) bool {
	sim := *b
	sim.Points[m.From] = Point{Owner: c, Checkers: b.Points[m.From].Checkers - 1}
	if sim.Points[m.From].Checkers == 0 {
		sim.Points[m.From].Owner = NoColor
	}
	if m.To >= 1 && m.To <= 24 {
		sim.Points[m.To] = Point{Owner: c, Checkers: b.Points[m.To].Checkers + 1}
	}

	opponentAhead := func(start, end int) bool {
		opp := c.Opponent()
		if c == White {
			for p := end + 1; p <= 24; p++ {
				if sim.Points[p].Owner == opp {
					return true
				}
			}
		} else {
			for p := start - 1; p >= 1; p-- {
				if sim.Points[p].Owner == opp {
					return true
				}
			}
		}
		return false
	}

	run := 0
	runStart := 0
	for p := 1; p <= 24; p++ {
		if sim.Points[p].Owner == c {
			if run == 0 {
				runStart = p
			}
			run++
			if run >= 6 && opponentAhead(runStart, p) {
				return true
			}
		} else {
			run = 0
			runStart = 0
		}
	}
	return false
}

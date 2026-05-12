package game

// GenerateSingleMoves returns all valid moves for color c using the given die.
func GenerateSingleMoves(b *Board, c Color, die int) []Move {
	var out []Move
	for from := 1; from <= 24; from++ {
		if b.Points[from].Owner != c || b.Points[from].Checkers == 0 {
			continue
		}
		to := from + c.Direction()*die
		if to >= 1 && to <= 24 {
			m := Move{From: from, To: to, Die: die}
			if ok, _ := IsValidMove(b, c, m); ok {
				out = append(out, m)
			}
		}
	}

	// Bear-off moves when all checkers are in home.
	if b.AllInHome(c) {
		target := c.BearOffTarget()
		// Exact bear-off.
		exact := pointForBearOffDie(c, die)
		if exact >= 1 && exact <= 24 && b.Points[exact].Owner == c && b.Points[exact].Checkers > 0 {
			out = append(out, Move{From: exact, To: target, Die: die})
		}
		// Fallback: highest occupied point when die > its distance.
		highest := highestOccupiedInHome(b, c)
		if highest > 0 && distanceToBearOff(c, highest) < die {
			out = append(out, Move{From: highest, To: target, Die: die})
		}
	}
	return out
}

// GenerateSequences returns all maximal-length move sequences for the given dice.
// When only one die can be used and the dice are distinct, enforces the
// "must use the larger die" rule.
func GenerateSequences(b *Board, c Color, dice []int) [][]Move {
	var all [][]Move
	collect(b, c, dice, []Move{}, &all)

	if len(all) == 0 {
		return nil
	}

	maxLen := 0
	for _, seq := range all {
		if len(seq) > maxLen {
			maxLen = len(seq)
		}
	}

	var filtered [][]Move
	for _, seq := range all {
		if len(seq) == maxLen {
			filtered = append(filtered, seq)
		}
	}

	// If maxLen == 1 and dice had two distinct values, keep only sequences
	// that use the larger die.
	if maxLen == 1 && len(dice) == 2 && dice[0] != dice[1] {
		larger := dice[0]
		if dice[1] > larger {
			larger = dice[1]
		}
		var onlyLarger [][]Move
		for _, seq := range filtered {
			if seq[0].Die == larger {
				onlyLarger = append(onlyLarger, seq)
			}
		}
		if len(onlyLarger) > 0 {
			filtered = onlyLarger
		}
	}
	return filtered
}

func collect(b *Board, c Color, dice []int, prefix []Move, out *[][]Move) {
	if len(dice) == 0 {
		cp := append([]Move(nil), prefix...)
		*out = append(*out, cp)
		return
	}

	anyMove := false
	used := make(map[int]bool)
	for i, die := range dice {
		if used[die] {
			continue
		}
		used[die] = true

		moves := GenerateSingleMoves(b, c, die)
		for _, m := range moves {
			anyMove = true
			sim := *b
			_ = m.Apply(&sim, c)
			rest := append([]int{}, dice[:i]...)
			rest = append(rest, dice[i+1:]...)
			collect(&sim, c, rest, append(prefix, m), out)
		}
	}

	if !anyMove {
		cp := append([]Move(nil), prefix...)
		*out = append(*out, cp)
	}
}

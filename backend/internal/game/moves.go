package game

// GenerateSingleMoves returns all valid moves for color c using the given die.
func GenerateSingleMoves(b *Board, c Color, die int) []Move {
	var out []Move
	for from := 1; from <= 24; from++ {
		if b.Points[from].Owner != c || b.Points[from].Checkers == 0 {
			continue
		}
		to := from + c.Direction()*die
		m := Move{From: from, To: to, Die: die}
		if ok, _ := IsValidMove(b, c, m); ok {
			out = append(out, m)
		}
	}
	return out
}

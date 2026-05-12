package game

type Point struct {
	Owner    Color
	Checkers int
}

type Board struct {
	Points   [25]Point
	BorneOff [3]int
}

func NewBoard() *Board {
	b := &Board{}
	b.Points[White.StartPoint()] = Point{Owner: White, Checkers: 15}
	b.Points[Black.StartPoint()] = Point{Owner: Black, Checkers: 15}
	return b
}

func (b *Board) CountCheckers(c Color) int {
	total := b.BorneOff[c]
	for i := 1; i <= 24; i++ {
		if b.Points[i].Owner == c {
			total += b.Points[i].Checkers
		}
	}
	return total
}

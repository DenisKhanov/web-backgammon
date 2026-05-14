package game

import "fmt"

type Phase int

const (
	PhaseWaiting Phase = iota
	PhaseRollingFirst
	PhasePlaying
	PhaseBearingOff
	PhaseFinished
)

type Game struct {
	Board             *Board
	CurrentTurn       Color
	Dice              []int
	RemainingDice     []int
	Phase             Phase
	Winner            Color
	IsMars            bool
	MoveCount         int
	HeadMovesThisTurn [3]int
	TurnsCompleted    [3]int
}

func NewGame() *Game {
	return &Game{
		Board:       NewBoard(),
		CurrentTurn: NoColor,
		Phase:       PhaseWaiting,
		Winner:      NoColor,
	}
}

func (g *Game) Roll(d Dice) error {
	if g.Phase != PhasePlaying && g.Phase != PhaseRollingFirst && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot roll in phase %d", g.Phase)
	}
	a, b := d.Roll()
	g.Dice = []int{a, b}
	g.RemainingDice = ExpandDice(a, b)
	g.HeadMovesThisTurn[g.CurrentTurn] = 0
	if g.Phase == PhaseRollingFirst {
		g.Phase = PhasePlaying
	}
	return nil
}

func (g *Game) ApplyMove(m Move) error {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot move in phase %d", g.Phase)
	}

	if !g.isLegalNextMove(m) {
		return fmt.Errorf("move is not legal for the current turn")
	}

	dieIdx := -1
	for i, d := range g.RemainingDice {
		if d == m.Die {
			dieIdx = i
			break
		}
	}
	if dieIdx < 0 {
		return fmt.Errorf("die %d not available in remaining dice %v", m.Die, g.RemainingDice)
	}

	if g.Board.Points[m.From].Owner != g.CurrentTurn {
		return fmt.Errorf("not your checker at %d", m.From)
	}
	if m.From == g.CurrentTurn.StartPoint() &&
		g.HeadMovesThisTurn[g.CurrentTurn] >= g.maxHeadMovesThisTurn() {
		return fmt.Errorf("cannot move more checkers from start point this turn")
	}

	if ok, err := IsValidMove(g.Board, g.CurrentTurn, m); !ok {
		return err
	}

	if err := m.Apply(g.Board, g.CurrentTurn); err != nil {
		return err
	}

	g.RemainingDice = append(g.RemainingDice[:dieIdx], g.RemainingDice[dieIdx+1:]...)
	if m.From == g.CurrentTurn.StartPoint() {
		g.HeadMovesThisTurn[g.CurrentTurn]++
	}
	g.MoveCount++

	if g.Board.AllInHome(g.CurrentTurn) && g.Phase == PhasePlaying {
		g.Phase = PhaseBearingOff
	}

	if g.Board.BorneOff[g.CurrentTurn] == 15 {
		g.Phase = PhaseFinished
		g.Winner = g.CurrentTurn
		g.IsMars = g.Board.BorneOff[g.CurrentTurn.Opponent()] == 0
	}
	return nil
}

func (g *Game) EndTurn() error {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return fmt.Errorf("cannot end turn in phase %d", g.Phase)
	}

	if len(g.RemainingDice) > 0 {
		sequences := GenerateSequencesForTurn(
			g.Board,
			g.CurrentTurn,
			g.RemainingDice,
			g.HeadMovesThisTurn[g.CurrentTurn],
			g.maxHeadMovesThisTurn(),
		)
		for _, seq := range sequences {
			if len(seq) > 0 {
				return fmt.Errorf("cannot end turn: %d usable moves remain", len(seq))
			}
		}
	}

	g.TurnsCompleted[g.CurrentTurn]++
	g.HeadMovesThisTurn[g.CurrentTurn] = 0
	g.CurrentTurn = g.CurrentTurn.Opponent()
	g.Dice = nil
	g.RemainingDice = nil
	return nil
}

func (g *Game) AvailableMoves() [][]Move {
	if g.Phase != PhasePlaying && g.Phase != PhaseBearingOff {
		return nil
	}
	if len(g.RemainingDice) == 0 {
		return nil
	}
	return GenerateSequencesForTurn(
		g.Board,
		g.CurrentTurn,
		g.RemainingDice,
		g.HeadMovesThisTurn[g.CurrentTurn],
		g.maxHeadMovesThisTurn(),
	)
}

func (g *Game) isLegalNextMove(m Move) bool {
	sequences := g.AvailableMoves()
	for _, seq := range sequences {
		if len(seq) == 0 {
			continue
		}
		first := seq[0]
		if first == m {
			return true
		}
	}
	return false
}

func (g *Game) RollFirst(d Dice) error {
	if g.Phase != PhaseWaiting {
		return fmt.Errorf("RollFirst allowed only in PhaseWaiting, got %d", g.Phase)
	}
	for {
		a, b := d.Roll()
		if a == b {
			continue
		}
		if a > b {
			g.CurrentTurn = White
		} else {
			g.CurrentTurn = Black
		}
		g.Dice = []int{a, b}
		g.RemainingDice = ExpandDice(a, b)
		g.HeadMovesThisTurn[g.CurrentTurn] = 0
		g.Phase = PhasePlaying
		return nil
	}
}

func (g *Game) maxHeadMovesThisTurn() int {
	if g.CurrentTurn != White && g.CurrentTurn != Black {
		return 0
	}
	if g.TurnsCompleted[g.CurrentTurn] == 0 &&
		len(g.Dice) == 2 &&
		g.Dice[0] == g.Dice[1] &&
		(g.Dice[0] == 3 || g.Dice[0] == 4 || g.Dice[0] == 6) {
		return 2
	}
	return 1
}

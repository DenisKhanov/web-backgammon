package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGame_InitialState(t *testing.T) {
	g := NewGame()
	assert.Equal(t, PhaseWaiting, g.Phase)
	assert.NotNil(t, g.Board)
	assert.Equal(t, 15, g.Board.Points[24].Checkers)
	assert.Equal(t, NoColor, g.Winner)
}

func TestGame_Roll_TransitionsFromRollingFirst(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseRollingFirst
	g.CurrentTurn = White

	err := g.Roll(NewFixedDice([][2]int{{3, 5}}))

	assert.NoError(t, err)
	assert.Equal(t, []int{3, 5}, g.Dice)
	assert.Equal(t, []int{3, 5}, g.RemainingDice)
	assert.Equal(t, PhasePlaying, g.Phase)
}

func TestGame_Roll_DoubleExpansion(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White

	err := g.Roll(NewFixedDice([][2]int{{4, 4}}))

	assert.NoError(t, err)
	assert.Equal(t, []int{4, 4}, g.Dice)
	assert.Equal(t, []int{4, 4, 4, 4}, g.RemainingDice)
}

func TestGame_Roll_RejectedInWrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseFinished

	err := g.Roll(NewFixedDice([][2]int{{1, 2}}))

	assert.Error(t, err)
}

func TestGame_ApplyMove_HappyPath(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{3, 5}
	g.RemainingDice = []int{3, 5}

	err := g.ApplyMove(Move{From: 24, To: 19, Die: 5})

	assert.NoError(t, err)
	assert.Equal(t, []int{3}, g.RemainingDice)
	assert.Equal(t, 14, g.Board.Points[24].Checkers)
	assert.Equal(t, 1, g.Board.Points[19].Checkers)
	assert.Equal(t, 1, g.MoveCount)
}

func TestGame_ApplyMove_RejectsSecondCheckerFromStartOnDistinctDice(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{2, 3}
	g.RemainingDice = []int{2, 3}

	require.NoError(t, g.ApplyMove(Move{From: 24, To: 22, Die: 2}))

	err := g.ApplyMove(Move{From: 24, To: 21, Die: 3})

	assert.Error(t, err)
	assert.Equal(t, []int{3}, g.RemainingDice)
	assert.Equal(t, 14, g.Board.Points[24].Checkers)
}

func TestGame_ApplyMove_AllowsSecondCheckerFromStartOnFirstTurnDouble(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{6, 6}
	g.RemainingDice = []int{6, 6, 6, 6}

	require.NoError(t, g.ApplyMove(Move{From: 24, To: 18, Die: 6}))
	err := g.ApplyMove(Move{From: 24, To: 18, Die: 6})

	assert.NoError(t, err)
	assert.Equal(t, 13, g.Board.Points[24].Checkers)
	assert.Equal(t, 2, g.Board.Points[18].Checkers)
}

func TestGame_ApplyMove_RejectsSecondCheckerFromStartOnFirstTurnDoubleFive(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{5, 5}
	g.RemainingDice = []int{5, 5, 5, 5}

	require.NoError(t, g.ApplyMove(Move{From: 24, To: 19, Die: 5}))
	err := g.ApplyMove(Move{From: 24, To: 19, Die: 5})

	assert.Error(t, err)
	assert.Equal(t, 14, g.Board.Points[24].Checkers)
	assert.Equal(t, 1, g.Board.Points[19].Checkers)
}

func TestGame_ApplyMove_RejectsMoveThatIsNotLegalFirstMove(t *testing.T) {
	b := &Board{}
	b.Points[14] = Point{Owner: White, Checkers: 1}
	b.Points[6] = Point{Owner: Black, Checkers: 1}
	g := &Game{
		Board:         b,
		CurrentTurn:   White,
		Dice:          []int{3, 5},
		RemainingDice: []int{3, 5},
		Phase:         PhasePlaying,
	}

	err := g.ApplyMove(Move{From: 14, To: 11, Die: 3})

	assert.Error(t, err)
	assert.Equal(t, []int{3, 5}, g.RemainingDice)
	assert.Equal(t, White, g.Board.Points[14].Owner)
}

func TestGame_AvailableMoves_AfterStartMoveExcludesSecondCheckerFromStart(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Dice = []int{2, 3}
	g.RemainingDice = []int{2, 3}

	require.NoError(t, g.ApplyMove(Move{From: 24, To: 22, Die: 2}))

	sequences := g.AvailableMoves()

	for _, seq := range sequences {
		for _, move := range seq {
			assert.NotEqual(t, 24, move.From)
		}
	}
}

func TestGame_ApplyMove_InvalidMoveRejected(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3}

	err := g.ApplyMove(Move{From: 10, To: 7, Die: 3})

	assert.Error(t, err)
	assert.Equal(t, []int{3}, g.RemainingDice, "remaining dice must not change on rejection")
}

func TestGame_ApplyMove_DieNotInRemaining(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{5}

	err := g.ApplyMove(Move{From: 24, To: 21, Die: 3})

	assert.Error(t, err, "die 3 not in remaining dice")
}

func TestGame_ApplyMove_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.ApplyMove(Move{From: 24, To: 21, Die: 3})

	assert.Error(t, err)
}

func TestGame_ApplyMove_NotPlayerTurn(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = Black
	g.RemainingDice = []int{6}

	err := g.ApplyMove(Move{From: 24, To: 18, Die: 6})

	assert.Error(t, err)
}

func TestGame_EndTurn_AfterAllDiceUsed(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{}

	err := g.EndTurn()

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
	assert.Empty(t, g.Dice)
	assert.Empty(t, g.RemainingDice)
}

func TestGame_EndTurn_RefusedWhenDiceUsable(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3}

	err := g.EndTurn()

	assert.Error(t, err)
	assert.Equal(t, White, g.CurrentTurn, "turn must stay with white")
}

func TestGame_EndTurn_AllowedWhenNoUsableMoves(t *testing.T) {
	// Checker at 10, die=6 lands at 4 which is occupied by opponent.
	// No other checkers, no valid moves → EndTurn must succeed.
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[10] = Point{Owner: White, Checkers: 1}
	g.Board.Points[4] = Point{Owner: Black, Checkers: 1}
	g.RemainingDice = []int{6}

	err := g.EndTurn()

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
}

func TestGame_Victory_WhiteWins(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 5
	g.RemainingDice = []int{1}

	err := g.ApplyMove(Move{From: 1, To: 0, Die: 1})

	assert.NoError(t, err)
	assert.Equal(t, PhaseFinished, g.Phase)
	assert.Equal(t, White, g.Winner)
	assert.False(t, g.IsMars)
}

func TestGame_Victory_Mars(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 0
	g.Board.Points[20] = Point{Owner: Black, Checkers: 15}
	g.RemainingDice = []int{1}

	err := g.ApplyMove(Move{From: 1, To: 0, Die: 1})

	assert.NoError(t, err)
	assert.Equal(t, PhaseFinished, g.Phase)
	assert.Equal(t, White, g.Winner)
	assert.True(t, g.IsMars)
}

func TestGame_AvailableMoves_ReturnsSequences(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.RemainingDice = []int{3, 5}

	sequences := g.AvailableMoves()

	assert.NotEmpty(t, sequences)
	for _, seq := range sequences {
		assert.NotEmpty(t, seq)
		for _, m := range seq {
			assert.Contains(t, []int{3, 5}, m.Die)
		}
	}
}

func TestGame_AvailableMoves_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	sequences := g.AvailableMoves()

	assert.Nil(t, sequences)
}

func TestGame_RollFirst_WhiteWinsAndStarts(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.RollFirst(NewFixedDice([][2]int{{5, 3}}))

	assert.NoError(t, err)
	assert.Equal(t, PhasePlaying, g.Phase)
	assert.Equal(t, White, g.CurrentTurn, "white rolled 5, black rolled 3 → white starts")
	assert.Equal(t, []int{5, 3}, g.Dice)
	assert.Equal(t, []int{5, 3}, g.RemainingDice)
}

func TestGame_RollFirst_BlackWins(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting

	err := g.RollFirst(NewFixedDice([][2]int{{2, 6}}))

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
	assert.Equal(t, []int{2, 6}, g.RemainingDice)
}

func TestGame_RollFirst_Tie_Rerolls(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseWaiting
	d := NewFixedDice([][2]int{{4, 4}, {2, 5}})

	err := g.RollFirst(d)

	assert.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn, "tie 4-4, then 2-5 → black wins")
	assert.Equal(t, []int{2, 5}, g.Dice)
}

func TestGame_RollFirst_WrongPhase(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying

	err := g.RollFirst(NewFixedDice([][2]int{{1, 2}}))

	assert.Error(t, err)
}

func TestGame_RollFirst_AllNonTieRollsHaveAFirstMove(t *testing.T) {
	for a := 1; a <= 6; a++ {
		for b := 1; b <= 6; b++ {
			if a == b {
				continue
			}

			g := NewGame()
			require.NoError(t, g.RollFirst(NewFixedDice([][2]int{{a, b}})))

			sequences := g.AvailableMoves()
			require.NotEmpty(t, sequences, "roll %d:%d should produce sequences", a, b)
			assert.NotEmpty(t, sequences[0], "roll %d:%d should produce a first move", a, b)
		}
	}
}

// --- Result classification tests ---

func TestGame_Result_Oin(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 5
	g.RemainingDice = []int{1}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	assert.Equal(t, ResultOin, g.Result)
	assert.Equal(t, 1, g.ResultPoints)
	assert.False(t, g.IsMars)
}

func TestGame_Result_Mars(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 0
	// Black has checkers outside white's home → mars, not koks.
	g.Board.Points[10] = Point{Owner: Black, Checkers: 15}
	g.RemainingDice = []int{1}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	assert.Equal(t, ResultMars, g.Result)
	assert.Equal(t, 2, g.ResultPoints)
	assert.True(t, g.IsMars)
}

func TestGame_Result_Koks(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 0
	// Black has a checker in white's home (points 1-6) → koks.
	g.Board.Points[3] = Point{Owner: Black, Checkers: 15}
	g.RemainingDice = []int{1}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	assert.Equal(t, ResultKoks, g.Result)
	assert.Equal(t, 3, g.ResultPoints)
	assert.True(t, g.IsMars)
}

func TestGame_Result_HomeMars(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.BorneOff[White] = 14
	g.Board.BorneOff[Black] = 0
	// All black checkers are in black's home (19-24), none borne off → home mars.
	for p := 19; p <= 24; p++ {
		g.Board.Points[p] = Point{Owner: Black, Checkers: 2}
	}
	g.Board.Points[19] = Point{Owner: Black, Checkers: 5} // 5+2*5=15
	g.RemainingDice = []int{1}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	assert.Equal(t, ResultHomeMars, g.Result)
	assert.Equal(t, 4, g.ResultPoints)
	assert.True(t, g.IsMars)
}

// --- Per-player bearing-off phase tests ---

func TestGame_BearingOff_PerPlayer(t *testing.T) {
	g := NewGame()
	g.Phase = PhasePlaying
	g.CurrentTurn = White
	g.Board = &Board{}
	// White: all in home
	for p := 1; p <= 6; p++ {
		g.Board.Points[p] = Point{Owner: White, Checkers: 2}
	}
	g.Board.Points[1] = Point{Owner: White, Checkers: 3} // total 15
	g.Board.BorneOff[White] = 0
	// Black: not in home
	g.Board.Points[10] = Point{Owner: Black, Checkers: 15}
	g.Dice = []int{1, 2}
	g.RemainingDice = []int{1, 2}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	assert.True(t, g.BearingOff[White], "white should be in bearing-off")
	assert.False(t, g.BearingOff[Black], "black should not be in bearing-off")
	assert.Equal(t, PhaseBearingOff, g.Phase)
}

func TestGame_EndTurn_RevertsPhaseWhenOpponentNotBearingOff(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.BearingOff[White] = true
	g.BearingOff[Black] = false
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 1}
	g.Board.Points[10] = Point{Owner: Black, Checkers: 15}
	g.Board.BorneOff = [3]int{White: 14, Black: 0}
	g.Dice = []int{1, 2}
	g.RemainingDice = []int{1}

	require.NoError(t, g.ApplyMove(Move{From: 1, To: 0, Die: 1}))

	// White wins, game over — no EndTurn needed.
	assert.Equal(t, PhaseFinished, g.Phase)
}

func TestGame_EndTurn_PhaseRevertsToPlayingForNonBearingOffPlayer(t *testing.T) {
	g := NewGame()
	g.Phase = PhaseBearingOff
	g.CurrentTurn = White
	g.BearingOff[White] = true
	g.BearingOff[Black] = false
	g.Board = &Board{}
	g.Board.Points[1] = Point{Owner: White, Checkers: 2}
	g.Board.Points[10] = Point{Owner: Black, Checkers: 15}
	g.Board.BorneOff = [3]int{White: 13, Black: 0}
	g.RemainingDice = []int{}
	g.Dice = []int{1, 2}

	err := g.EndTurn()

	require.NoError(t, err)
	assert.Equal(t, Black, g.CurrentTurn)
	assert.Equal(t, PhasePlaying, g.Phase, "phase should revert to playing for non-bearing-off player")
}

func TestGame_IsBearingOff(t *testing.T) {
	g := NewGame()
	g.BearingOff[White] = true
	g.BearingOff[Black] = false

	assert.True(t, g.IsBearingOff(White))
	assert.False(t, g.IsBearingOff(Black))
}

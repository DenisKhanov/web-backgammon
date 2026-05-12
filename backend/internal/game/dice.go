package game

import "math/rand"

type Dice interface {
	Roll() (int, int)
}

type RandomDice struct {
	rng *rand.Rand
}

func NewRandomDice(seed int64) *RandomDice {
	return &RandomDice{rng: rand.New(rand.NewSource(seed))}
}

func (d *RandomDice) Roll() (int, int) {
	return d.rng.Intn(6) + 1, d.rng.Intn(6) + 1
}

type FixedDice struct {
	rolls []([2]int)
	idx   int
}

func NewFixedDice(rolls [][2]int) *FixedDice {
	return &FixedDice{rolls: rolls}
}

func (d *FixedDice) Roll() (int, int) {
	r := d.rolls[d.idx]
	d.idx++
	return r[0], r[1]
}

func ExpandDice(a, b int) []int {
	if a == b {
		return []int{a, a, a, a}
	}
	return []int{a, b}
}

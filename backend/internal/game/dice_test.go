package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandDice_Regular(t *testing.T) {
	result := ExpandDice(3, 5)
	assert.Equal(t, []int{3, 5}, result)
}

func TestExpandDice_Double(t *testing.T) {
	result := ExpandDice(4, 4)
	assert.Equal(t, []int{4, 4, 4, 4}, result, "double gives 4 uses")
}

func TestFixedDice_Roll(t *testing.T) {
	d := NewFixedDice([][2]int{{3, 5}, {6, 6}})

	a, b := d.Roll()
	assert.Equal(t, 3, a)
	assert.Equal(t, 5, b)

	a, b = d.Roll()
	assert.Equal(t, 6, a)
	assert.Equal(t, 6, b)
}

func TestRandomDice_RollInRange(t *testing.T) {
	d := NewRandomDice(42)
	for i := 0; i < 100; i++ {
		a, b := d.Roll()
		assert.GreaterOrEqual(t, a, 1)
		assert.LessOrEqual(t, a, 6)
		assert.GreaterOrEqual(t, b, 1)
		assert.LessOrEqual(t, b, 6)
	}
}

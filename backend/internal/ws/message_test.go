package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode_MoveError(t *testing.T) {
	b := mustEncode("move_error", MoveErrorPayload{Reason: "glukhoi_zabor"})

	var env Message
	require.NoError(t, json.Unmarshal(b, &env))
	assert.Equal(t, "move_error", env.Type)

	var p MoveErrorPayload
	require.NoError(t, json.Unmarshal(env.Payload, &p))
	assert.Equal(t, "glukhoi_zabor", p.Reason)
}

func TestEncode_GameState(t *testing.T) {
	payload := GameStatePayload{
		Phase:       "playing",
		CurrentTurn: "white",
		MyColor:     "white",
		Dice:        []int{3, 5},
	}
	b := mustEncode("game_state", payload)
	assert.Contains(t, string(b), `"type":"game_state"`)
	assert.Contains(t, string(b), `"phase":"playing"`)
}

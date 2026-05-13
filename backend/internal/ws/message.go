package ws

import "encoding/json"

// --- Envelope ---

// Message is the top-level JSON envelope for every WS frame.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func encode(msgType string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Type: msgType, Payload: p})
}

func mustEncode(msgType string, payload any) []byte {
	b, err := encode(msgType, payload)
	if err != nil {
		panic(err)
	}
	return b
}

// --- Client → Server payload types ---

type MovePayload struct {
	From  int           `json:"from"`
	To    int           `json:"to"`
	Die   int           `json:"die"`
	Steps []MovePayload `json:"steps,omitempty"`
}

type ChatPayload struct {
	Text string `json:"text"`
}

// end_turn, pass, ping have no payload fields.

// --- Server → Client payload types ---

type BoardPoint struct {
	Owner    int `json:"owner"` // 0=none 1=white 2=black
	Checkers int `json:"checkers"`
}

type PlayerSnapshot struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	Connected bool   `json:"connected"`
}

type GameStatePayload struct {
	Phase         string           `json:"phase"`
	CurrentTurn   string           `json:"currentTurn"`
	Board         [25]BoardPoint   `json:"board"`
	BorneOff      [3]int           `json:"borneOff"`
	Dice          []int            `json:"dice"`
	RemainingDice []int            `json:"remainingDice"`
	LegalMoves    []MovePayload    `json:"legalMoves"`
	MoveCount     int              `json:"moveCount"`
	MyColor       string           `json:"myColor"`
	Players       []PlayerSnapshot `json:"players"`
	TimeLeft      int              `json:"timeLeft"` // seconds remaining in current turn
}

type DiceRolledPayload struct {
	Dice     []int  `json:"dice"`
	IsDouble bool   `json:"isDouble"`
	Player   string `json:"player"` // "white" | "black"
}

type OpponentMovedPayload struct {
	From          int   `json:"from"`
	To            int   `json:"to"`
	Die           int   `json:"die"`
	RemainingDice []int `json:"remainingDice"`
}

type TurnChangedPayload struct {
	Player   string `json:"player"`
	TimeLeft int    `json:"timeLeft"`
}

type MoveErrorPayload struct {
	Reason string `json:"reason"`
}

type ChatMessagePayload struct {
	From string `json:"from"`
	Text string `json:"text"`
	Time string `json:"time"` // "15:04"
}

type GameOverPayload struct {
	Winner string `json:"winner"`
	IsMars bool   `json:"isMars"`
}

type OpponentDisconnectedPayload struct {
	GracePeriod int `json:"gracePeriod"` // seconds
}

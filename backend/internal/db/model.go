package db

import "time"

type Room struct {
	ID        string
	Code      string
	Status    string // waiting | playing | finished | abandoned
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Player struct {
	ID           string
	RoomID       string
	Name         string
	Color        *string // nil until assigned; "white" | "black"
	SessionToken string
	JoinedAt     time.Time
	LastSeenAt   time.Time
}

type GameRecord struct {
	ID            string
	RoomID        string
	BoardState    []byte // JSONB raw bytes
	CurrentTurn   *string
	Dice          []int
	RemainingDice []int
	Phase         string
	Winner        *string
	IsMars        bool
	TurnStartedAt time.Time
	MoveCount     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID         uuid.UUID  `json:"id"`
	Mode       string     `json:"mode"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

type GamePlayer struct {
	ID           uuid.UUID `json:"id"`
	GameID       uuid.UUID `json:"game_id"`
	UserID       uuid.UUID `json:"user_id"`
	Score        int       `json:"score"`
	LinesCleared int       `json:"lines_cleared"`
	LevelReached int       `json:"level_reached"`
	Placement    *int      `json:"placement"`
	IsWinner     bool      `json:"is_winner"`
}

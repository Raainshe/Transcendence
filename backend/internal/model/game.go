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

type CreateGameRequest struct {
	Mode         string    `json:"mode"`
	Score        int       `json:"score"`
	LinesCleared int       `json:"lines_cleared"`
	LevelReached int       `json:"level_reached"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	IsWinner     bool      `json:"is_winner"`
}

type LeaderboardEntry struct {
	Rank         int64      `json:"rank"`
	UserID       uuid.UUID  `json:"user_id"`
	Username     string     `json:"username"`
	AvatarURL    *string    `json:"avatar_url"`
	Score        int        `json:"score"`
	LinesCleared int        `json:"lines_cleared"`
	LevelReached int        `json:"level_reached"`
	Mode         string     `json:"mode"`
	FinishedAt   *time.Time `json:"finished_at"`
}

type UserStats struct {
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
	BestScore   int `json:"best_score"`
	TotalLines  int `json:"total_lines"`
	AvgScore    int `json:"avg_score"`
}

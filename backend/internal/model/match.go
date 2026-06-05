package model

import "github.com/google/uuid"

type MatchPlayerView struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

type MatchDetail struct {
	GameID     uuid.UUID          `json:"game_id"`
	Status     string             `json:"status"`
	Mode       string             `json:"mode"`
	SharedSeed int64              `json:"shared_seed"`
	Players    []MatchPlayerView  `json:"players"`
	Results    *MatchEndedPayload `json:"results,omitempty"`
}

// PlayerConnectionBroadcast is sent for player.disconnected and player.reconnected.
type PlayerConnectionBroadcast struct {
	UserID string `json:"user_id"`
}

// PlayerEliminatedUpload is the client-sent payload for player.eliminated.
type PlayerEliminatedUpload struct {
	Reason string `json:"reason"`
	Score  int    `json:"score"`
	Lines  int    `json:"lines"`
	Level  int    `json:"level"`
}

// PlayerEliminatedBroadcast is the server-sent payload for player.eliminated.
type PlayerEliminatedBroadcast struct {
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"`
	Placement int    `json:"placement"`
}

// MatchEndedPlayer is one row in the match.ended payload.
type MatchEndedPlayer struct {
	UserID            uuid.UUID `json:"user_id"`
	Username          string    `json:"username"`
	Score             int       `json:"score"`
	Lines             int       `json:"lines"`
	Level             int       `json:"level"`
	Placement         int       `json:"placement"`
	IsWinner          bool      `json:"is_winner"`
	EliminationReason *string   `json:"elimination_reason,omitempty"`
}

// MatchEndedPayload is broadcast when a multiplayer match finishes.
type MatchEndedPayload struct {
	WinnerUserID *uuid.UUID         `json:"winner_user_id,omitempty"`
	Players      []MatchEndedPlayer `json:"players"`
}

// PlayerMatchStats holds final scoreboard values for one player.
type PlayerMatchStats struct {
	UserID uuid.UUID
	Score  int
	Lines  int
	Level  int
}

// PlayerElimination records elimination order metadata from the lifecycle tracker.
type PlayerElimination struct {
	UserID    uuid.UUID
	Reason    string
	Placement int
}

// EndMatchInput is built by the WebSocket handler when a match should finish.
type EndMatchInput struct {
	GameID        uuid.UUID
	SurvivorID    *uuid.UUID
	AllEliminated bool
	Stats         []PlayerMatchStats
	Eliminations  []PlayerElimination
}

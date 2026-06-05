package model

import "github.com/google/uuid"

type MatchPlayerView struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

type MatchDetail struct {
	GameID     uuid.UUID         `json:"game_id"`
	Status     string            `json:"status"`
	Mode       string            `json:"mode"`
	SharedSeed int64             `json:"shared_seed"`
	Players    []MatchPlayerView `json:"players"`
}

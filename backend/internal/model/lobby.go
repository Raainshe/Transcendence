package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	LobbyStatusWaiting = "waiting"
	LobbyStatusClosed  = "closed"
)

type Lobby struct {
	ID          uuid.UUID  `json:"id"`
	HostUserID  uuid.UUID  `json:"host_user_id"`
	InviteCode  string     `json:"invite_code"`
	MaxPlayers  int        `json:"max_players"`
	Status      string     `json:"status"`
	GameID      *uuid.UUID `json:"game_id,omitempty"`
	SharedSeed  *int64     `json:"shared_seed,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type LobbyMemberView struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	IsReady   bool      `json:"is_ready"`
	JoinedAt  time.Time `json:"joined_at"`
}

type LobbyDetail struct {
	Lobby
	Members []LobbyMemberView `json:"members"`
}

type CreateLobbyRequest struct {
	MaxPlayers int `json:"max_players"`
}

type SetReadyRequest struct {
	Ready bool `json:"ready"`
}

type JoinLobbyByCodeRequest struct {
	InviteCode string `json:"invite_code"`
}

type MatchStartPlayer struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

type StartLobbyResult struct {
	GameID     uuid.UUID          `json:"game_id"`
	SharedSeed int64              `json:"shared_seed"`
	Players    []MatchStartPlayer `json:"players"`
}

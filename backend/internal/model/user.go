package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OnlineWindow is how long after the last authenticated request a user is
// considered "online". HTTP-based presence; will tighten once WebSocket lands.
const OnlineWindow = 5 * time.Minute

type UpdateUserRequest struct {
	Username  *string `json:"username"`
	AvatarURL *string `json:"avatar_url"`
}

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash *string    `json:"-"`
	AvatarURL    *string    `json:"avatar_url"`
	AchievementList Achievements `json:"achievement_list"`
	Role         string     `json:"role"`
	Is2FAEnabled bool       `json:"is_2fa_enabled"`
	TOTPSecret   *string    `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
}

type Achievements struct {
	AvatarChange 	bool `json:"avatar_change"`
	HighestScore2K 	bool `json:"highest_score_2_k"`
	HighestScore5K 	bool `json:"highest_score_5_k"`
	HighestScore10K bool `json:"highest_score_10_k"`
	TotalPoints50K 	bool `json:"total_points_50_k"`
	TotalPoints100K bool `json:"total_points_100_k"`
	Level2        	bool `json:"level_2"`
	Level5        	bool `json:"level_5"`
	Level10        	bool `json:"level_10"`
	Streak2			bool `json:"streak_2"`
	Streak5			bool `json:"streak_5"`
	FirstMpGame		bool `json:"first_mp_game"`
	Played10		bool `json:"played_10"`
	Played50		bool `json:"played_50"`
	Played100		bool `json:"played_100"`
	FirstFriend		bool `json:"first_friend"`
	FirstYear		bool `json:"first_year"`
	FirstClear		bool `json:"first_clear"`
	FirstTetris		bool `json:"first_tetris"`
}

// MarshalJSON adds a computed is_online field. Value receiver so it fires
// for both User and *User; the wrapper type avoids infinite recursion.
func (u User) MarshalJSON() ([]byte, error) {
	type alias User
	return json.Marshal(&struct {
		alias
		IsOnline bool `json:"is_online"`
	}{
		alias:    alias(u),
		IsOnline: u.LastSeenAt != nil && time.Since(*u.LastSeenAt) < OnlineWindow,
	})
}

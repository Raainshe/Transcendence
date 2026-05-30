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
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	PasswordHash  *string    `json:"-"`
	AvatarURL     *string    `json:"avatar_url"`
	Role          string     `json:"role"`
	Is2FAEnabled  bool       `json:"is_2fa_enabled"`
	TOTPSecret    *string    `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastSeenAt    *time.Time `json:"last_seen_at"`
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

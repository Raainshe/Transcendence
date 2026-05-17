package model

import (
	"time"

	"github.com/google/uuid"
)

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

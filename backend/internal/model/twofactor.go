package model

import (
	"time"
	"github.com/google/uuid"
)

const (
	TwoFAPurposeLogin = "login"
	TwoFAPurposeEnable = "enable"
)

type TwoFactorCode struct {
	ID uuid.UUID
	UserID uuid.UUID
	CodeHash string
	Purpose string
	ExpiresAt time.Time
	ConsumedAt *time.Time
	CreatedAt time.Time
}

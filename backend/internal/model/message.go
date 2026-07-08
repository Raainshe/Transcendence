package model

import (
	"time"
	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `json:"id"`
	SenderID uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	ReadAt *time.Time `json:"read_at"`
}

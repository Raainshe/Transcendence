package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// ParseUserID validates a JWT and returns the embedded user id.
func ParseUserID(secret, tokenStr string) (uuid.UUID, error) {
	var c Claims
	token, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, jwt.ErrSignatureInvalid
	}
	return c.UserID, nil
}

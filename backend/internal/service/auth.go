package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/model"
	"backend/internal/repository"
)

var (
	ErrEmailTaken   = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
	ErrInvalidCreds  = errors.New("invalid credentials")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthService struct {
	users     repository.UserRepository
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, secret string) *AuthService {
	return &AuthService{users: users, jwtSecret: secret}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*model.User, string, error) {
	if _, err := s.users.FindByEmail(ctx, req.Email); err == nil {
		return nil, "", ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", err
	}

	if _, err := s.users.FindByUsername(ctx, req.Username); err == nil {
		return nil, "", ErrUsernameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, "", err
	}
	hashStr := string(hash)

	now := time.Now().UTC()
	user := &model.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: &hashStr,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.issueToken(user.ID)
	return user, token, err
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*model.User, string, error) {
	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrInvalidCreds
		}
		return nil, "", err
	}

	if user.PasswordHash == nil {
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCreds
	}

	token, err := s.issueToken(user.ID)
	return user, token, err
}

func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	var c Claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidCreds
	}
	return s.issueToken(c.UserID)
}

func (s *AuthService) issueToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

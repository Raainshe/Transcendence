package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/mailer"
	
	"fmt"
	"crypto/rand"
	"math/big"
)

const MinPasswordLength = 8

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
	ErrInvalidCreds  = errors.New("invalid credentials")
	ErrPasswordWeak  = errors.New("password must be at least 8 characters")
	ErrTwoFARequired  = errors.New("two-factor authentication required")
	ErrInvalidCode  = errors.New("invalid or expired verification code")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Scope string `json:"scope,omitempty"`
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
	codes	  repository.TwoFactorRepository
	mailer    mailer.Mailer
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, codes repository.TwoFactorRepository, m mailer.Mailer, secret string) *AuthService {
	return &AuthService{users: users, codes: codes, mailer: m, jwtSecret: secret}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*model.User, string, error) {
	if len(req.Password) < MinPasswordLength {
		return nil, "", ErrPasswordWeak
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	if _, err := s.users.FindByEmail(ctx, email); err == nil {
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
		Email:        email,
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
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.users.FindByEmail(ctx, email)
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

	if user.Is2FAEnabled {
		if err := s.sendCode(ctx, user, model.TwoFAPurposeLogin); err != nil {
			return nil, "", err
		}
		pending, err := s.issuePendingToken(user.ID)
		if err != nil {
			return nil, "", err
		}
		return user, pending, ErrTwoFARequired
	}

	token, err := s.issueToken(user.ID)
	return user, token, err
}

const scopePending = "2fa_pending"

func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	var c Claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid || c.Scope == scopePending {
		return "", ErrInvalidCreds
	}
	return s.issueToken(c.UserID)
}

func (s *AuthService) issueToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		Scope: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) VerifyLogin(ctx context.Context, pendingToken, code string) (string, error) {
	userID, err := s.parsePending(pendingToken)
	if err != nil {
		return "", ErrInvalidCreds
	}
	if err := s.checkCode(ctx, userID, model.TwoFAPurposeLogin, code); err != nil {
		return "", err
	}
	return s.issueToken(userID)
}

func (s *AuthService) RequestEnable2FA(ctx context.Context, userID uuid.UUID) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	return s.sendCode(ctx, user, model.TwoFAPurposeEnable)
}

func (s *AuthService) ConfirmEnable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	if err := s.checkCode(ctx, userID, model.TwoFAPurposeEnable, code); err != nil {
		return err
	}
	return s.users.Set2FAEnabled(ctx, userID, true)
}

func (s *AuthService) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	return s.users.Set2FAEnabled(ctx, userID, false)
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *AuthService) sendCode(ctx context.Context, user *model.User, purpose string) error {
	code, err := generateCode()
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_ = s.codes.DeleteForUser(ctx, user.ID, purpose)
	now := time.Now().UTC()
	rec := &model.TwoFactorCode{
		ID: uuid.New(),
		UserID: user.ID,
		CodeHash: string(hash),
		Purpose: purpose,
		ExpiresAt: now.Add(10 * time.Minute),
		CreatedAt: now,
	}
	if err := s.codes.Create(ctx, rec); err != nil {
		return err
	}
	body := fmt.Sprintf("Your verification code is %s. It expires in 10 minutes.", code)
	return s.mailer.Send(user.Email, "Your verification code", body)
}

func (s *AuthService) checkCode(ctx context.Context, userID uuid.UUID, purpose, code string) error {
	rec, err := s.codes.GetActive(ctx, userID, purpose)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidCode
		}
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(rec.CodeHash), []byte(code)) != nil {
		return ErrInvalidCode
	}
	return s.codes.MarkConsumed(ctx, rec.ID)
}

func (s *AuthService) issuePendingToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID,
		Scope: scopePending,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) parsePending(tokenString string) (uuid.UUID, error) {
	var c Claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func (t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid || c.Scope != scopePending {
		return uuid.Nil, ErrInvalidCreds
	}
	return c.UserID, nil
}

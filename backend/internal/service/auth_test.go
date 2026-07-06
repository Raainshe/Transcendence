package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"backend/internal/mailer"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestAuthService_Register(t *testing.T) {
	existingUser := testutil.NewTestUser()

	tests := []struct {
		name             string
		req              service.RegisterRequest
		findByEmailFn    func(context.Context, string) (*model.User, error)
		findByUsernameFn func(context.Context, string) (*model.User, error)
		createFn         func(context.Context, *model.User) error
		wantErr          error
		wantToken        bool
	}{
		{
			name:      "success",
			req:       service.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "secret12"},
			wantToken: true,
		},
		{
			name: "email already taken",
			req:  service.RegisterRequest{Username: "alice", Email: "taken@example.com", Password: "secret12"},
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return existingUser, nil
			},
			wantErr: service.ErrEmailTaken,
		},
		{
			name: "username already taken",
			req:  service.RegisterRequest{Username: "taken", Email: "new@example.com", Password: "secret12"},
			findByUsernameFn: func(_ context.Context, _ string) (*model.User, error) {
				return existingUser, nil
			},
			wantErr: service.ErrUsernameTaken,
		},
		{
			name: "db error on create",
			req:  service.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "secret12"},
			createFn: func(_ context.Context, _ *model.User) error {
				return errors.New("db down")
			},
			wantErr: errors.New("db down"),
		},
		{
			name:    "password too short",
			req:     service.RegisterRequest{Username: "alice", Email: "alice@example.com", Password: "short"},
			wantErr: service.ErrPasswordWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{
				FindByEmailFn:    tt.findByEmailFn,
				FindByUsernameFn: tt.findByUsernameFn,
				CreateFn:         tt.createFn,
			}
			svc := service.NewAuthService(repo, &testutil.MockTwoFactorRepo{}, mailer.LogMailer{}, "test-secret")

			user, token, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if tt.wantErr == service.ErrEmailTaken || tt.wantErr == service.ErrUsernameTaken || tt.wantErr == service.ErrPasswordWeak {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("error = %v, want %v", err, tt.wantErr)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Error("expected non-nil user")
			}
			if tt.wantToken && token == "" {
				t.Error("expected non-empty token")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	password := "correct-password"
	hash := testutil.HashPassword(password)

	userWithPassword := testutil.NewTestUser()
	userWithPassword.PasswordHash = &hash

	userWithoutPassword := testutil.NewTestUser()

	tests := []struct {
		name          string
		req           service.LoginRequest
		findByEmailFn func(context.Context, string) (*model.User, error)
		wantErr       error
		wantToken     bool
	}{
		{
			name: "success",
			req:  service.LoginRequest{Email: "test@example.com", Password: password},
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return userWithPassword, nil
			},
			wantToken: true,
		},
		{
			name:    "user not found",
			req:     service.LoginRequest{Email: "nobody@example.com", Password: password},
			wantErr: service.ErrInvalidCreds,
		},
		{
			name: "oauth user has no password hash",
			req:  service.LoginRequest{Email: "oauth@example.com", Password: password},
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return userWithoutPassword, nil
			},
			wantErr: service.ErrInvalidCreds,
		},
		{
			name: "wrong password",
			req:  service.LoginRequest{Email: "test@example.com", Password: "wrong-password"},
			findByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
				return userWithPassword, nil
			},
			wantErr: service.ErrInvalidCreds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{FindByEmailFn: tt.findByEmailFn}
			svc := service.NewAuthService(repo, &testutil.MockTwoFactorRepo{}, mailer.LogMailer{}, "test-secret")

			user, token, err := svc.Login(context.Background(), tt.req)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Error("expected non-nil user")
			}
			if tt.wantToken && token == "" {
				t.Error("expected non-empty token")
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	userID := uuid.New()
	const secret = "test-secret"

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantToken bool
	}{
		{
			name:      "valid token returns new token",
			token:     testutil.MakeTestToken(userID, secret),
			wantToken: true,
		},
		{
			name:    "garbage string",
			token:   "not.a.real.token",
			wantErr: true,
		},
		{
			name:    "signed with wrong secret",
			token:   testutil.MakeTestToken(userID, "other-secret"),
			wantErr: true,
		},
		{
			name:    "expired token",
			token:   testutil.MakeExpiredToken(userID, secret),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewAuthService(&testutil.MockUserRepo{}, &testutil.MockTwoFactorRepo{}, mailer.LogMailer{}, secret)

			newToken, err := svc.RefreshToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantToken && newToken == "" {
				t.Error("expected non-empty token")
			}
		})
	}
}

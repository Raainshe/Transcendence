package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestGameService_RecordMatch(t *testing.T) {
	userID := uuid.New()
	req := model.CreateGameRequest{
		Mode:         "marathon",
		Score:        1000,
		LinesCleared: 40,
		LevelReached: 5,
		StartedAt:    time.Now().Add(-5 * time.Minute),
		FinishedAt:   time.Now(),
		IsWinner:     true,
	}

	tests := []struct {
		name          string
		recordMatchFn func(context.Context, *model.Game, *model.GamePlayer) error
		wantErr       bool
	}{
		{
			name: "success",
		},
		{
			name: "db error",
			recordMatchFn: func(_ context.Context, _ *model.Game, _ *model.GamePlayer) error {
				return errors.New("insert failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{RecordMatchFn: tt.recordMatchFn}
			svc := service.NewGameService(repo)

			game, err := svc.RecordMatch(context.Background(), userID, req)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if game == nil {
				t.Fatal("expected non-nil game")
			}
			if game.Mode != req.Mode {
				t.Errorf("game.Mode = %q, want %q", game.Mode, req.Mode)
			}
			if game.Status != "finished" {
				t.Errorf("game.Status = %q, want %q", game.Status, "finished")
			}
		})
	}
}

func TestGameService_GetLeaderboard(t *testing.T) {
	var capturedLimit int

	tests := []struct {
		name       string
		inputLimit int
		wantLimit  int
	}{
		{name: "zero defaults to 20", inputLimit: 0, wantLimit: 20},
		{name: "over 100 clamped", inputLimit: 500, wantLimit: 100},
		{name: "within range used as-is", inputLimit: 25, wantLimit: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{
				ListLeaderboardFn: func(_ context.Context, limit int) ([]model.LeaderboardEntry, error) {
					capturedLimit = limit
					return []model.LeaderboardEntry{}, nil
				},
			}
			svc := service.NewGameService(repo)
			_, err := svc.GetLeaderboard(context.Background(), tt.inputLimit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedLimit != tt.wantLimit {
				t.Errorf("ListLeaderboard called with limit=%d, want %d", capturedLimit, tt.wantLimit)
			}
		})
	}
}

func TestGameService_GetUserStats(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		getUserStatsFn func(context.Context, uuid.UUID) (*model.UserStats, error)
		wantErr        bool
	}{
		{
			name: "success",
			getUserStatsFn: func(_ context.Context, _ uuid.UUID) (*model.UserStats, error) {
				return &model.UserStats{GamesPlayed: 5, Wins: 3, BestScore: 9000}, nil
			},
		},
		{
			name: "db error propagated",
			getUserStatsFn: func(_ context.Context, _ uuid.UUID) (*model.UserStats, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{GetUserStatsFn: tt.getUserStatsFn}
			svc := service.NewGameService(repo)

			stats, err := svc.GetUserStats(context.Background(), userID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stats == nil {
				t.Error("expected non-nil stats")
			}
		})
	}
}

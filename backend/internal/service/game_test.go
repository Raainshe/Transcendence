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

func TestGameService_ListGames(t *testing.T) {
	tests := []struct {
		name        string
		userID      *uuid.UUID
		inputLimit  int
		wantLimit   int
		listFn      func(context.Context, *uuid.UUID, int, int) ([]model.Game, error)
		countFn     func(context.Context, *uuid.UUID) (int, error)
		wantTotal   int
		wantErr     bool
	}{
		{name: "default limit (0 -> 20)", inputLimit: 0, wantLimit: 20, wantTotal: 0},
		{name: "limit > 100 clamped", inputLimit: 999, wantLimit: 100, wantTotal: 0},
		{name: "within range used as-is", inputLimit: 30, wantLimit: 30, wantTotal: 0},
		{
			name:    "repo error",
			listFn:  func(_ context.Context, _ *uuid.UUID, _, _ int) ([]model.Game, error) { return nil, errors.New("db down") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLimit int
			repo := &testutil.MockGameRepo{
				ListGamesFn: tt.listFn,
				CountGamesFn: tt.countFn,
			}
			if repo.ListGamesFn == nil {
				repo.ListGamesFn = func(_ context.Context, _ *uuid.UUID, limit, _ int) ([]model.Game, error) {
					capturedLimit = limit
					return []model.Game{}, nil
				}
			}
			svc := service.NewGameService(repo)
			_, total, err := svc.ListGames(context.Background(), tt.userID, tt.inputLimit, 0)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantLimit != 0 && capturedLimit != tt.wantLimit {
				t.Errorf("ListGames called with limit=%d, want %d", capturedLimit, tt.wantLimit)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestGameService_GetGame(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name             string
		findGameDetailFn func(context.Context, uuid.UUID) (*model.GameDetail, error)
		wantErr          bool
	}{
		{
			name: "success",
			findGameDetailFn: func(_ context.Context, _ uuid.UUID) (*model.GameDetail, error) {
				return &model.GameDetail{Game: model.Game{ID: id, Mode: "marathon"}}, nil
			},
		},
		{
			name:    "not found defaults",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockGameRepo{FindGameDetailFn: tt.findGameDetailFn}
			svc := service.NewGameService(repo)
			detail, err := svc.GetGame(context.Background(), id)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if detail == nil || detail.ID != id {
				t.Errorf("detail = %+v, want id=%v", detail, id)
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

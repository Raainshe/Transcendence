package service

import (
	"context"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

type GameService struct {
	games repository.GameRepository
}

func NewGameService(games repository.GameRepository) *GameService {
	return &GameService{games: games}
}

func (s *GameService) RecordMatch(ctx context.Context, userID uuid.UUID, req model.CreateGameRequest) (*model.Game, error) {
	placement := 1
	game := &model.Game{
		ID:         uuid.New(),
		Mode:       req.Mode,
		Status:     "finished",
		CreatedAt:  req.StartedAt,
		FinishedAt: &req.FinishedAt,
	}
	player := &model.GamePlayer{
		ID:           uuid.New(),
		GameID:       game.ID,
		UserID:       userID,
		Score:        req.Score,
		LinesCleared: req.LinesCleared,
		LevelReached: req.LevelReached,
		Placement:    &placement,
		IsWinner:     req.IsWinner,
	}
	if err := s.games.RecordMatch(ctx, game, player); err != nil {
		return nil, err
	}
	return game, nil
}

func (s *GameService) GetLeaderboard(ctx context.Context, limit int) ([]model.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.games.ListLeaderboard(ctx, limit)
}

func (s *GameService) GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error) {
	return s.games.GetUserStats(ctx, userID)
}

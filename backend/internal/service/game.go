package service

import (
	"context"
	"log"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

type GameService struct {
	games        repository.GameRepository
	achievements *AchievementService
}

func NewGameService(games repository.GameRepository, achievements *AchievementService) *GameService {
	return &GameService{games: games, achievements: achievements}
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

	isWinner := req.IsWinner
	if req.Mode != "multiplayer" {
		isWinner = false
	} //doIneed?

	player := &model.GamePlayer{
		ID:           uuid.New(),
		GameID:       game.ID,
		UserID:       userID,
		Score:        req.Score,
		LinesCleared: req.LinesCleared,
		LevelReached: req.LevelReached,
		Placement:    &placement,
		IsWinner:     isWinner,
	}
	if err := s.games.RecordMatch(ctx, game, player); err != nil {
		return nil, err
	}
	err := s.achievements.OnGameEnd(ctx, userID, *player, *game)
	if err != nil {
		log.Printf("achievement check failed for user %s: %v", userID, err)
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

func (s *GameService) ListGames(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Game, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	games, err := s.games.ListGames(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.games.CountGames(ctx, userID)
	return games, total, err
}

func (s *GameService) GetGame(ctx context.Context, id uuid.UUID) (*model.GameDetail, error) {
	return s.games.FindGameDetail(ctx, id)
}

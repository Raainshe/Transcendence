package service

import (
	"context"
	"log"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

type AchievementService struct {
	achievements repository.AchievementsRepository
	games        repository.GameRepository
}

func NewAchievementService(achievements repository.AchievementsRepository, games repository.GameRepository) *AchievementService {
	return &AchievementService{achievements: achievements, games: games}
}

func (s *AchievementService) GetAchievementsByID(ctx context.Context, id uuid.UUID) (*model.Achievements, error) {
	return s.achievements.FindAchievementsByID(ctx, id)
}
func (s *AchievementService) OnGameEnd(ctx context.Context, userID uuid.UUID, player model.GamePlayer, game model.Game) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}

	//log.Printf("OnGameEnd: userID=%s score=%d lines=%d level=%d isWinner=%v mode=%s",
	//	userID, player.Score, player.LinesCleared, player.LevelReached, player.IsWinner, game.Mode)

	changed := false

	if !a.FirstClear && player.LinesCleared >= 1 {
		a.FirstClear = true
		changed = true
	}
	/* 	if !a.FirstTetris && player.LinesCleared >= 4 {
		a.FirstTetris = true
		changed = true
	} */ //dont have data for this
	if !a.HighestScore2K && player.Score >= 2000 {
		a.HighestScore2K = true
		changed = true
	}
	if !a.HighestScore10K && player.Score >= 10000 {
		a.HighestScore10K = true
		changed = true
		log.Printf("unlocked HighestScore10K for user %s", userID)
	}
	if !a.HighestScore50K && player.Score >= 50000 {
		a.HighestScore50K = true
		changed = true
	}
	if !a.Level2 && player.LevelReached >= 2 {
		a.Level2 = true
		changed = true
	}
	if !a.Level10 && player.LevelReached >= 10 {
		a.Level10 = true
		changed = true
	}
	if !a.Level50 && player.LevelReached >= 50 {
		a.Level50 = true
		changed = true
	}
	if !a.FirstWin && player.IsWinner {
		a.FirstWin = true
		changed = true
	}

	if !a.Played10 || !a.Played50 || !a.Played100 || !a.TotalPoints30K || !a.TotalPoints100K {
		stats, err := s.games.GetUserStats(ctx, userID)
		if err != nil {
			return err
		}
		if !a.Played10 && stats.GamesPlayed >= 10 {
			a.Played10 = true
			changed = true
		}
		if !a.Played50 && stats.GamesPlayed >= 50 {
			a.Played50 = true
			changed = true
		}
		if !a.Played100 && stats.GamesPlayed >= 100 {
			a.Played100 = true
			changed = true
		}
		if !a.TotalPoints30K && stats.TotalScore >= 30000 {
			a.TotalPoints30K = true
			changed = true
		}
		if !a.TotalPoints100K && stats.TotalScore >= 100000 {
			a.TotalPoints100K = true
			changed = true
		}
		if !a.TotalPoints250K && stats.TotalScore >= 250000 {
			a.TotalPoints250K = true
			changed = true
		}
		if !a.HundrethWin && stats.Wins == 100 {
			a.HundrethWin = true
			changed = true
		}
	}

	if changed {
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

// LobbyService.StartLobby
func (s *AchievementService) OnMPGame(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.FirstMpGame {
		a.FirstMpGame = true
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

func (s *AchievementService) OnFriendAdded(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.FirstFriend {
		a.FirstFriend = true
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

func (s *AchievementService) OnAvatarUploaded(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.AvatarChange {
		a.AvatarChange = true
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

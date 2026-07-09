package service

import (
	"context"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

const (
	XPGame        = 20
	XPWin         = 50
	XPAchievement = 100

// XPImproveScore = 20 would have to call stats every time....
)

type GamificationService struct {
	achievements repository.AchievementsRepository
	games        repository.GameRepository
	user         repository.UserRepository
}

func NewGamificationService(achievements repository.AchievementsRepository, games repository.GameRepository, user repository.UserRepository) *GamificationService {
	return &GamificationService{achievements: achievements, games: games, user: user}
}

func (s *GamificationService) GetAchievementsByID(ctx context.Context, id uuid.UUID) (*model.Achievements, error) {
	return s.achievements.FindAchievementsByID(ctx, id)
}
func (s *GamificationService) OnGameEnd(ctx context.Context, userID uuid.UUID, player model.GamePlayer, game model.Game) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}

	//log.Printf("OnGameEnd: userID=%s score=%d lines=%d level=%d isWinner=%v mode=%s",
	//	userID, player.Score, player.LinesCleared, player.LevelReached, player.IsWinner, game.Mode)

	changed := 0

	if !a.FirstClear && player.LinesCleared >= 1 {
		a.FirstClear = true
		changed++
	}
	/* 	if !a.FirstTetris && player.LinesCleared >= 4 {
		a.FirstTetris = true
		changed = true
	} */ //dont have data for this
	if !a.HighestScore2K && player.Score >= 2000 {
		a.HighestScore2K = true
		changed++
	}
	if !a.HighestScore10K && player.Score >= 10000 {
		a.HighestScore10K = true
		changed++
	}
	if !a.HighestScore50K && player.Score >= 50000 {
		a.HighestScore50K = true
		changed++
	}
	if !a.Level2 && player.LevelReached >= 2 {
		a.Level2 = true
		changed++
	}
	if !a.Level10 && player.LevelReached >= 10 {
		a.Level10 = true
		changed++
	}
	if !a.Level50 && player.LevelReached >= 50 {
		a.Level50 = true
		changed++
	}
	if !a.FirstWin && player.IsWinner {
		a.FirstWin = true
		changed++
	}

	if !a.Played10 || !a.Played50 || !a.Played100 || !a.TotalPoints30K || !a.TotalPoints100K {
		stats, err := s.games.GetUserStats(ctx, userID)
		if err != nil {
			return err
		}
		if !a.Played10 && stats.GamesPlayed >= 10 {
			a.Played10 = true
			changed++
		}
		if !a.Played50 && stats.GamesPlayed >= 50 {
			a.Played50 = true
			changed++
		}
		if !a.Played100 && stats.GamesPlayed >= 100 {
			a.Played100 = true
			changed++
		}
		if !a.TotalPoints30K && stats.TotalScore >= 30000 {
			a.TotalPoints30K = true
			changed++
		}
		if !a.TotalPoints100K && stats.TotalScore >= 100000 {
			a.TotalPoints100K = true
			changed++
		}
		if !a.TotalPoints250K && stats.TotalScore >= 250000 {
			a.TotalPoints250K = true
			changed++
		}
		if !a.HundrethWin && stats.Wins == 100 {
			a.HundrethWin = true
			changed++
		}
	}

	s.user.AddXP(ctx, userID, changed*XPAchievement+XPGame)
	if player.IsWinner {
		s.user.AddXP(ctx, userID, XPWin)
	}

	if changed != 0 {
		return s.achievements.Update(ctx, userID, *a)
	}

	return nil
}

// LobbyService.StartLobby
func (s *GamificationService) OnMPGame(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.FirstMpGame {
		a.FirstMpGame = true
		s.user.AddXP(ctx, userID, XPAchievement)
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

func (s *GamificationService) OnFriendAdded(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.FirstFriend {
		a.FirstFriend = true
		s.user.AddXP(ctx, userID, XPAchievement)
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

func (s *GamificationService) OnAvatarUploaded(ctx context.Context, userID uuid.UUID) error {
	a, err := s.achievements.FindAchievementsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !a.AvatarChange {
		a.AvatarChange = true
		s.user.AddXP(ctx, userID, XPAchievement)
		return s.achievements.Update(ctx, userID, *a)
	}
	return nil
}

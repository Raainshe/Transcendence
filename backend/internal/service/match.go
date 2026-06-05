package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

var (
	ErrNotMatchMember = errors.New("user is not a match participant")
	ErrNotMultiplayer = errors.New("game is not a multiplayer match")
)

type MatchService struct {
	games   repository.GameRepository
	lobbies repository.LobbyRepository
}

func NewMatchService(games repository.GameRepository, lobbies repository.LobbyRepository) *MatchService {
	return &MatchService{games: games, lobbies: lobbies}
}

func (s *MatchService) GetMatch(ctx context.Context, callerID, gameID uuid.UUID) (*model.MatchDetail, error) {
	game, err := s.games.FindByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Mode != "multiplayer" {
		return nil, ErrNotMultiplayer
	}

	ok, err := s.games.IsGamePlayer(ctx, gameID, callerID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotMatchMember
	}

	lobby, err := s.lobbies.FindByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if lobby.SharedSeed == nil || *lobby.SharedSeed == 0 {
		return nil, repository.ErrNotFound
	}

	players, err := s.games.ListMatchPlayers(ctx, gameID)
	if err != nil {
		return nil, err
	}

	detail := &model.MatchDetail{
		GameID:     gameID,
		Status:     game.Status,
		Mode:       game.Mode,
		SharedSeed: *lobby.SharedSeed,
		Players:    players,
	}

	if game.Status == "finished" {
		results, err := s.games.ListMatchResults(ctx, gameID)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
		if results != nil {
			detail.Results = results
		}
	}

	return detail, nil
}

package service

import "backend/internal/repository"

type GameService struct {
	games repository.GameRepository
}

func NewGameService(games repository.GameRepository) *GameService {
	return &GameService{games: games}
}

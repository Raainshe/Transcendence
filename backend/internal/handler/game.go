package handler

import (
	"net/http"

	"backend/internal/service"
)

type GameHandler struct {
	games *service.GameService
}

func NewGameHandler(games *service.GameService) *GameHandler {
	return &GameHandler{games: games}
}

func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

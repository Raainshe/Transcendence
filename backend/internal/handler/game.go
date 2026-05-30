package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
)

type GameHandler struct {
	games *service.GameService
}

func NewGameHandler(games *service.GameService) *GameHandler {
	return &GameHandler{games: games}
}

var validModes = map[string]bool{
	"marathon":    true,
	"sprint":      true,
	"ultra":       true,
	"multiplayer": true,
}

func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var req model.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if !validModes[req.Mode] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "mode must be one of: marathon, sprint, ultra, multiplayer"})
		return
	}
	game, err := h.games.RecordMatch(r.Context(), userID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to record match"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"game": game})
}

func (h *GameHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	stats, err := h.games.GetUserStats(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user stats"})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *GameHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	entries, err := h.games.GetLeaderboard(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch leaderboard"})
		return
	}
	if entries == nil {
		entries = []model.LeaderboardEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
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
	var userIDFilter *uuid.UUID
	if raw := r.URL.Query().Get("user_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id must be a valid UUID"})
			return
		}
		userIDFilter = &id
	}
	limit, ok := parseQueryInt(r, "limit", 0)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be an integer"})
		return
	}
	offset, ok := parseQueryInt(r, "offset", 0)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "offset must be an integer"})
		return
	}

	games, total, err := h.games.ListGames(r.Context(), userIDFilter, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list games"})
		return
	}
	if games == nil {
		games = []model.Game{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"games": games, "total": total})
}

func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid game id"})
		return
	}
	detail, err := h.games.GetGame(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "game not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch game"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"game": detail})
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
	limit, ok := parseQueryInt(r, "limit", 0)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be an integer"})
		return
	}
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

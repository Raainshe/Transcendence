package handler

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PublicHandler struct {
	games *service.GameService
	lobby *service.LobbyService
	user  *service.UserService
}

type publicLeaderboardEntry struct {
	Rank         int64     `json:"rank"`
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	AvatarURL    *string   `json:"avatar_url"`
	Score        int       `json:"score"`
	LevelReached int       `json:"level_reached"`
	Mode         string    `json:"mode"`
}

type publicUserStats struct {
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
	BestScore   int `json:"best_score"`
	TotalLines  int `json:"total_lines"`
	AvgScore    int `json:"avg_score"`
}

func NewPublicHandler(games *service.GameService, lobby *service.LobbyService, user *service.UserService) *PublicHandler {
	return &PublicHandler{games: games, lobby: lobby, user: user}
}

func (h *PublicHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, 200, map[string]any{"leaderboard": toPublicLeaderboard(entries)})
}

func (h *PublicHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	_, err = h.user.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}
	stats, err := h.games.GetUserStats(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user stats"})
		return
	}
	writeJSON(w, 200, map[string]any{"stats": toPublicUserStats(stats)})
}

func toPublicLeaderboard(entries []model.LeaderboardEntry) []publicLeaderboardEntry {
	pe := make([]publicLeaderboardEntry, 0, len(entries))
	for _, e := range entries {
		pe = append(pe, publicLeaderboardEntry{
			Rank:         e.Rank,
			UserID:       e.UserID,
			Username:     e.Username,
			AvatarURL:    e.AvatarURL,
			Score:        e.Score,
			LevelReached: e.LevelReached,
			Mode:         e.Mode,
		})
	}
	return pe
}

func toPublicUserStats(stats *model.UserStats) publicUserStats {
	ps := publicUserStats{
		GamesPlayed: stats.GamesPlayed,
		Wins:        stats.Wins,
		BestScore:   stats.BestScore,
		TotalLines:  stats.TotalLines,
		AvgScore:    stats.AvgScore,
	}
	return ps
}

package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PublicHandler struct {
	games *service.GameService
	lobby *service.LobbyService
	user  *service.UserService
}

// The public types below are a narrowed view of the internal
// models (model.Lobby, model.LeaderboardEntry, model.UserStats, ...)
// to keep the public API contract stable as internal models evolve.
// When adding a field to one of those models, decide explicitly whether it
// belongs here too
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

type publicLobbyMember struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	IsReady   bool      `json:"is_ready"`
}

type publicLobby struct {
	ID         uuid.UUID           `json:"id"`
	HostUserID uuid.UUID           `json:"host_user_id"`
	InviteCode string              `json:"invite_code"`
	MaxPlayers int                 `json:"max_players"`
	Status     string              `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	Name       string              `json:"name"`
	Members    []publicLobbyMember `json:"members"`
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
	writeJSON(w, http.StatusOK, map[string]any{"leaderboard": toPublicLeaderboard(entries)})
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
	writeJSON(w, http.StatusOK, map[string]any{"stats": toPublicUserStats(stats)})
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

func (h *PublicHandler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		MaxPlayers int `json:"max_players"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	lobby, err := h.lobby.CreateLobby(r.Context(), userID, req.MaxPlayers)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"lobby": toPublicLobby(lobby)})
}

func toPublicLobby(d *model.LobbyDetail) publicLobby {
	members := make([]publicLobbyMember, 0, len(d.Members))
	for _, m := range d.Members {
		members = append(members, publicLobbyMember{
			UserID:    m.UserID,
			Username:  m.Username,
			AvatarURL: m.AvatarURL,
			IsReady:   m.IsReady,
		})
	}
	return publicLobby{
		ID:         d.ID,
		HostUserID: d.HostUserID,
		InviteCode: d.InviteCode,
		MaxPlayers: d.MaxPlayers,
		Status:     d.Status,
		CreatedAt:  d.CreatedAt,
		Name:       d.Name,
		Members:    members,
	}
}

func (h *PublicHandler) DeleteLobby(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}
	err = h.lobby.CloseLobby(r.Context(), userID, id)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PublicHandler) UpdateLobbyName(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	lobby, err := h.lobby.UpdateLobbyName(r.Context(), userID, id, req.Name)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lobby": toPublicLobby(lobby)})
}

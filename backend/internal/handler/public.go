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
	Rank         int64     `json:"rank" example:"1"`
	UserID       uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username     string    `json:"username" example:"fearcon"`
	AvatarURL    *string   `json:"avatar_url" example:"/uploads/avatars/550e8400-e29b-41d4-a716-446655440000/abc123.jpg"`
	Score        int       `json:"score" example:"98500"`
	LevelReached int       `json:"level_reached" example:"12"`
	Mode         string    `json:"mode" example:"marathon"`
}

type publicUserStats struct {
	GamesPlayed int `json:"games_played" example:"42"`
	Wins        int `json:"wins" example:"18"`
	BestScore   int `json:"best_score" example:"98500"`
	TotalLines  int `json:"total_lines" example:"1240"`
	AvgScore    int `json:"avg_score" example:"61200"`
}

type publicLobbyMember struct {
	UserID    uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username  string    `json:"username" example:"fearcon"`
	AvatarURL *string   `json:"avatar_url" example:"/uploads/avatars/550e8400-e29b-41d4-a716-446655440000/abc123.jpg"`
	IsReady   bool      `json:"is_ready" example:"true"`
}

type publicLobby struct {
	ID         uuid.UUID           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	HostUserID uuid.UUID           `json:"host_user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	InviteCode string              `json:"invite_code" example:"ABC123"`
	MaxPlayers int                 `json:"max_players" example:"4"`
	Status     string              `json:"status" example:"waiting"`
	CreatedAt  time.Time           `json:"created_at"`
	Name       string              `json:"name" example:"Friday Night Tetris"`
	Members    []publicLobbyMember `json:"members"`
}

type leaderboardResponse struct {
	Leaderboard []publicLeaderboardEntry `json:"leaderboard"`
}
type userStatsResponse struct {
	Stats publicUserStats `json:"stats"`
}
type lobbyResponse struct {
	Lobby publicLobby `json:"lobby"`
}
type errorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

type updateLobbyNameRequest struct {
	Name string `json:"name" example:"Friday Night Tetris"`
}

func NewPublicHandler(games *service.GameService, lobby *service.LobbyService, user *service.UserService) *PublicHandler {
	return &PublicHandler{games: games, lobby: lobby, user: user}
}

// @Summary		Shows current Leaderboard
// @Description	Returns the top players ranked by score across all modes. Limit defaults to 20 and is capped at 100.
// @Security 	ApiKeyAuth
// @Tags 		leaderboard
// @Produce 	json
// @Param 		limit query int false "Max entries to return"
// @Success 	200 {object} leaderboardResponse
// @Failure 	400 {object} errorResponse
// @Failure		500 {object} errorResponse
// @Router 		/leaderboard [get]
func (h *PublicHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit, ok := parseQueryInt(r, "limit", 0)
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "limit must be an integer"})
		return
	}
	entries, err := h.games.GetLeaderboard(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch leaderboard"})
		return
	}
	writeJSON(w, http.StatusOK, leaderboardResponse{Leaderboard: toPublicLeaderboard(entries)})
}

// @Summary		Shows User Statistics
// @Description Returns aggregate game stats (games played, wins, best score, total lines, average score) for the given user.
// @Security 	ApiKeyAuth
// @Tags		userstats
// @Produce 	json
// @Param 		id		path		string	true	"User ID"
// @Success 	200 	{object} 	userStatsResponse
// @Failure 	400 {object} errorResponse
// @Failure		404 {object} errorResponse
// @Failure		500 {object} errorResponse
// @Router		/users/{id}/stats [get]
func (h *PublicHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}
	_, err = h.user.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch user"})
		return
	}
	stats, err := h.games.GetUserStats(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to fetch user stats"})
		return
	}
	writeJSON(w, http.StatusOK, userStatsResponse{Stats: toPublicUserStats(stats)})
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

// @Summary		Creates a new lobby
// @Description Creates a new lobby with the caller as host. max_players must be between 2 and 4. A user can only be in one waiting lobby at a time — creating a second one fails with 409.
// @Security 	ApiKeyAuth
// @Tags		lobby
// @Accept 		json
// @Produce 	json
// @Param 		request body model.CreateLobbyRequest true "Lobby creation parameters"
// @Success 	201 {object} lobbyResponse
// @Failure 	400 {object} errorResponse
// @Failure 	409 {object} errorResponse
// @Failure 	500 {object} errorResponse
// @Router		/lobbies [post]
func (h *PublicHandler) CreateLobby(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req model.CreateLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	lobby, err := h.lobby.CreateLobby(r.Context(), userID, req.MaxPlayers)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, lobbyResponse{Lobby: toPublicLobby(lobby)})
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

// @Summary 	Deletes Lobby
// @Description Permanently deletes a lobby. Only the host may delete it.
// @Security 	ApiKeyAuth
// @Tags		lobby
// @Param		id path string true "Lobby ID"
// @Success 	204
// @Failure 	400 {object} errorResponse
// @Failure 	403 {object} errorResponse
// @Failure 	404 {object} errorResponse
// @Failure 	500 {object} errorResponse
// @Router		/lobbies/{id} [delete]
func (h *PublicHandler) DeleteLobby(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid lobby id"})
		return
	}
	err = h.lobby.CloseLobby(r.Context(), userID, id)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary		Sets or Changes current LobbyName
// @Description	Renames a lobby. Only the host may rename it, renaming after it has started or closed fails with 409. Name must be non-empty and 64 characters or fewer.
// @Security 	ApiKeyAuth
// @Tags		lobby
// @Accept		json
// @Produce		json
// @Param 		id path	string true "Lobby ID"
// @Param 		request body updateLobbyNameRequest true "New lobby name"
// @Success 	200 {object} lobbyResponse
// @Failure 	400 {object} errorResponse
// @Failure 	403 {object} errorResponse
// @Failure 	404 {object} errorResponse
// @Failure 	409 {object} errorResponse
// @Failure 	500 {object} errorResponse
// @Router		/lobbies/{id} [put]
func (h *PublicHandler) UpdateLobbyName(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid lobby id"})
		return
	}

	var req updateLobbyNameRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	lobby, err := h.lobby.UpdateLobbyName(r.Context(), userID, id, req.Name)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lobbyResponse{Lobby: toPublicLobby(lobby)})
}

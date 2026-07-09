package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	//"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
)

type AchievementsHandler struct {
	gamification *service.GamificationService
}

func NewAchievementsHandler(gamification *service.GamificationService) *AchievementsHandler {
	return &AchievementsHandler{gamification: gamification}
}

//only if we want it to be protected
/* func (h *AchievementsHandler) GetMyachievements(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	h.writeachievements(w, r, userID)
} */

func (h *AchievementsHandler) GetUserAchievements(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	h.writeAchievements(w, r, id)
}

func (h *AchievementsHandler) writeAchievements(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	a, err := h.gamification.GetAchievementsByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch achievements"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"achievements": a})
}

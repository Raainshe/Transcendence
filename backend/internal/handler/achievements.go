package handler

import (
	"errors"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
)

type AchievementHandler struct {
	achievements *service.AchievementService
}

func NewAchievementHandler(achievements *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievements: achievements}
}

func (h *AchievementHandler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	a, err := h.achievements.GetAchievementsByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"achievements": a})
}

func (h *AchievementHandler) GetBadges(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	a, err := h.achievements.GetAchievementsByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch badges"})
		return
	}
	badges := model.BadgesFromAchievements(*a)
	writeJSON(w, http.StatusOK, map[string]any{"badges": badges})
}

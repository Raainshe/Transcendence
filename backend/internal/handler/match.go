package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
)

type MatchHandler struct {
	matches *service.MatchService
}

func NewMatchHandler(matches *service.MatchService) *MatchHandler {
	return &MatchHandler{matches: matches}
}

func (h *MatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid match id"})
		return
	}

	detail, err := h.matches.GetMatch(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) ||
			errors.Is(err, service.ErrNotMatchMember) ||
			errors.Is(err, service.ErrNotMultiplayer) {
			writeJSON(w, http.StatusOK, map[string]any{"match": nil})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch match"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"match": detail})
}

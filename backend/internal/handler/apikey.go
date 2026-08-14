package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type APIKeyHandler struct {
	keys *service.APIKeyService
}

func NewAPIKeyHandler(keys *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{keys: keys}
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	key, err := h.keys.GenerateAPIKey(r.Context(), userID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrKeyNameRequired):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key name required"})
		case errors.Is(err, service.ErrKeyNameTooLong):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "API key name too long"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"api_key": key, "message": "Store this key, it will not be shown again."})
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	list, err := h.keys.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list API keys"})
		return
	}
	if list == nil {
		list = []model.APIKeyList{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"api_keys": list})
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid API key"})
		return
	}
	err = h.keys.Revoke(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "API key not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke API key"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

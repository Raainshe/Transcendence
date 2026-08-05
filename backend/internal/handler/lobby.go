package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
)

type LobbyHandler struct {
	lobbies *service.LobbyService
}

func NewLobbyHandler(lobbies *service.LobbyService) *LobbyHandler {
	return &LobbyHandler{lobbies: lobbies}
}

func (h *LobbyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req model.CreateLobbyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	detail, err := h.lobbies.CreateLobby(r.Context(), userID, req.MaxPlayers)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"lobby":       detail,
		"invite_code": detail.InviteCode,
	})
}

func (h *LobbyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	detail, err := h.lobbies.GetLobby(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "lobby not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch lobby"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": detail})
}

func (h *LobbyHandler) GetByCode(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimSpace(chi.URLParam(r, "code"))
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invite code is required"})
		return
	}

	detail, err := h.lobbies.GetLobbyByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "lobby not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch lobby"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": detail})
}

func (h *LobbyHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	detail, err := h.lobbies.JoinLobby(r.Context(), userID, id)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": detail})
}

func (h *LobbyHandler) JoinByCode(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req model.JoinLobbyByCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	code := strings.TrimSpace(req.InviteCode)
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invite_code is required"})
		return
	}

	detail, err := h.lobbies.JoinLobbyByCode(r.Context(), userID, code)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": detail})
}

func (h *LobbyHandler) Leave(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	if err := h.lobbies.LeaveLobby(r.Context(), userID, id); err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

func (h *LobbyHandler) SetReady(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	var req model.SetReadyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	detail, err := h.lobbies.SetReady(r.Context(), userID, id, req.Ready)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": detail})
}

func (h *LobbyHandler) Start(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid lobby id"})
		return
	}

	result, err := h.lobbies.StartLobby(r.Context(), userID, id)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeLobbyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "lobby not found"})
	case errors.Is(err, service.ErrInvalidMaxPlayers):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotLobbyHost):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrLobbyNotJoinable):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrLobbyFull):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrAlreadyInLobby):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrInAnotherLobby):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotEnoughPlayers):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotAllReady):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrNotLobbyMember):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "lobby operation failed"})
	}
}

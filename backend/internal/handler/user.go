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

const maxAvatarSize = 5 << 20 // 5 MB

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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
	users, total, err := h.users.ListUsers(r.Context(), limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list users"})
		return
	}
	if users == nil {
		users = []model.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users, "total": total})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch user"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	user, err := h.users.UpdateMe(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "username already in use"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update profile"})
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	// Override the global 1MB API cap with the larger avatar budget.
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid form data"})
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "avatar field is required"})
		return
	}
	defer file.Close()

	user, err := h.users.UploadAvatar(r.Context(), userID, file, header)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidFileType):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upload avatar"})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	if err := h.users.DeleteMe(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if err := h.users.DeleteAvatar(r.Context(), userID); err != nil {
		switch {
		case errors.Is(err, service.ErrNoAvatar):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "no avatar set"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete avatar"})
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	friends, err := h.users.GetFriends(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get friends"})
		return
	}
	if friends == nil {
		friends = []model.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"friends": friends})
}

func (h *UserHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	requests, err := h.users.GetPendingRequests(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get friend requests"})
		return
	}
	if requests == nil {
		requests = []model.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": requests})
}

func (h *UserHandler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	blocked, err := h.users.GetBlockedUsers(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get blocked users"})
		return
	}
	if blocked == nil {
		blocked = []model.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"blocked": blocked})
}

func (h *UserHandler) AddFriend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	err = h.users.SendFriendRequest(r.Context(), userID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSelfRelationship):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot add yourself as a friend"})
		case errors.Is(err, service.ErrRelationshipExists):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "relationship already exists"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to send friend request"})
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *UserHandler) AcceptFriend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	requesterID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	err = h.users.AcceptFriendRequest(r.Context(), userID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "friend request not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept friend request"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	otherID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	err = h.users.RemoveFriend(r.Context(), userID, otherID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "friendship not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove friend"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	err = h.users.BlockUser(r.Context(), userID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSelfRelationship):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot block yourself"})
		case errors.Is(err, service.ErrRelationshipExists):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "user is already blocked"})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to block user"})
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	targetID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	err = h.users.UnblockUser(r.Context(), userID, targetID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "user is not blocked"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to unblock user"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

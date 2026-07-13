package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChatHandler struct {
	chat *service.ChatService
}

func NewChatHandler(chat *service.ChatService) *ChatHandler {
	return &ChatHandler{chat: chat}
}

func (h *ChatHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	friendID, err := uuid.Parse(chi.URLParam(r, "friendID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	limit, _ := parseQueryInt(r, "limit", 50)
	messages, err := h.chat.GetConversation(r.Context(), userID, friendID, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotFriends) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load conversation"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (h *ChatHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	friendID, err := uuid.Parse(chi.URLParam(r, "friendID"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	if err := h.chat.MarkRead(r.Context(), userID, friendID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not mark read"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *ChatHandler) GetUnread(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	counts, err := h.chat.UnreadCounts(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not load unread counts"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"unread": counts})
}

package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"backend/internal/service"
	"backend/internal/middleware"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username, email and password are required"})
		return
	}

	user, token, err := h.auth.Register(r.Context(), service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already in use"})
		case errors.Is(err, service.ErrUsernameTaken):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "username already in use"})
		case errors.Is(err, service.ErrPasswordWeak):
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "registration failed"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"user": user, "token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, token, err := h.auth.Login(r.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrTwoFARequired) {
			writeJSON(w, http.StatusAccepted, map[string]any{
				"two_factor_required": true,
				"pending_token": token,
			})
			return
		}
		if errors.Is(err, service.ErrInvalidCreds) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user, "token": token})
}

func (h *AuthHandler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PendingToken string `json:"pending_token"`
		Code         string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	token, err := h.auth.VerifyLogin(r.Context(), req.PendingToken, req.Code)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid code or session"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid authorization header"})
		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := h.auth.RefreshToken(tokenStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) OAuthRedirect(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (h *AuthHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if err := h.auth.RequestEnable2FA(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not start 2FA setup"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "verification code sent"})
}

func (h *AuthHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if err := h.auth.ConfirmEnable2FA(r.Context(), userID, req.Code); err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid or expired code"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not enable 2FA"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "two-factor authentication enabled"})
}

func (h *AuthHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if err := h.auth.Disable2FA(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not disable 2FA"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "two-factor authentication disabled"})
}

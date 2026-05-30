//go:build integration

package itest

import (
	"net/http"
	"strings"
	"testing"
)

func TestAuth_RegisterLoginRefresh(t *testing.T) {
	truncate(t)

	// Register
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	// Use token on /users/me
	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/me", token, "")
	mustStatus(t, resp, raw, http.StatusOK)

	// Login
	resp, raw = doJSON(t, http.MethodPost, "/api/v1/auth/login", "",
		`{"email":"alice@example.com","password":"secret12"}`)
	mustStatus(t, resp, raw, http.StatusOK)
	var login struct {
		Token string `json:"token"`
	}
	decodeJSON(t, raw, &login)
	if login.Token == "" {
		t.Fatal("login returned empty token")
	}

	// Refresh
	resp, raw = doJSON(t, http.MethodPost, "/api/v1/auth/refresh", login.Token, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var refreshed struct {
		Token string `json:"token"`
	}
	decodeJSON(t, raw, &refreshed)
	if refreshed.Token == "" {
		t.Fatal("refresh returned empty token")
	}

	// New token works
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me", refreshed.Token, "")
	mustStatus(t, resp, raw, http.StatusOK)
}

func TestAuth_DuplicateEmail(t *testing.T) {
	truncate(t)
	registerUser(t, "alice", "dup@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/auth/register", "",
		`{"username":"alice2","email":"dup@example.com","password":"secret12"}`)
	mustStatus(t, resp, raw, http.StatusConflict)
}

func TestAuth_LoginCaseInsensitiveEmail(t *testing.T) {
	truncate(t)
	registerUser(t, "alice", "Alice@Example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/auth/login", "",
		`{"email":"alice@example.com","password":"secret12"}`)
	mustStatus(t, resp, raw, http.StatusOK)
}

func TestAuth_PasswordTooShort(t *testing.T) {
	truncate(t)
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/auth/register", "",
		`{"username":"bob","email":"bob@example.com","password":"short"}`)
	mustStatus(t, resp, raw, http.StatusBadRequest)
}

func TestAuth_MissingBearer(t *testing.T) {
	truncate(t)
	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/me", "", "")
	mustStatus(t, resp, raw, http.StatusUnauthorized)

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	decodeJSON(t, raw, &body)
	if body["error"] == "" {
		t.Errorf("expected error key in response, got %s", raw)
	}
}

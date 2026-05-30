//go:build integration

package itest

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestUsers_ListPagination(t *testing.T) {
	truncate(t)
	for i := 0; i < 25; i++ {
		registerUser(t, fmt.Sprintf("user%02d", i), fmt.Sprintf("u%02d@example.com", i), "secret12")
	}

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users?limit=10&offset=5", "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	var out struct {
		Users []map[string]any `json:"users"`
		Total int              `json:"total"`
	}
	decodeJSON(t, raw, &out)
	if len(out.Users) != 10 {
		t.Errorf("len(users) = %d, want 10", len(out.Users))
	}
	if out.Total != 25 {
		t.Errorf("total = %d, want 25", out.Total)
	}
}

func TestUsers_GetByID_NotFound(t *testing.T) {
	truncate(t)
	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/"+uuid.New().String(), "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

func TestUsers_UpdateMe_UsernameConflict(t *testing.T) {
	truncate(t)
	registerUser(t, "alice", "alice@example.com", "secret12")
	_, bobToken := registerUser(t, "bob", "bob@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPatch, "/api/v1/users/me", bobToken, `{"username":"alice"}`)
	mustStatus(t, resp, raw, http.StatusConflict)
}

func TestUsers_DeleteMe(t *testing.T) {
	truncate(t)
	aliceID, aliceToken := registerUser(t, "alice", "alice@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodDelete, "/api/v1/users/me", aliceToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/"+aliceID.String(), "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

func TestUsers_NoPasswordHashLeak(t *testing.T) {
	truncate(t)
	registerUser(t, "alice", "alice@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users", "", "")
	mustStatus(t, resp, raw, http.StatusOK)

	// bcrypt hashes always start with $2a$ / $2b$ / $2y$ — none of those bytes should appear.
	for _, marker := range [][]byte{[]byte("$2a$"), []byte("$2b$"), []byte("$2y$"), []byte("password_hash")} {
		if bytes.Contains(raw, marker) {
			t.Errorf("response body contains %q — sensitive data leak: %s", marker, raw)
		}
	}
}

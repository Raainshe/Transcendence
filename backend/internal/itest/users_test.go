//go:build integration

package itest

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUsers_LookupByUsername(t *testing.T) {
	truncate(t)
	registerUser(t, "alice", "alice@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodGet, "/api/v1/users?username=alice", "", "")
	mustStatus(t, resp, raw, http.StatusOK)
	var found struct {
		Users []struct {
			Username string `json:"username"`
		} `json:"users"`
		Total int `json:"total"`
	}
	decodeJSON(t, raw, &found)
	if found.Total != 1 || len(found.Users) != 1 || found.Users[0].Username != "alice" {
		t.Fatalf("lookup alice = %+v, body %s", found, raw)
	}

	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users?username=nobody", "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

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

func TestUsers_OnlineStatus(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	// last_seen_at is written by an async goroutine in the JWT middleware, so the
	// first authenticated GET may still see is_online=false. Poll until it lands.
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/me", token, "")
		mustStatus(t, resp, raw, http.StatusOK)
		var out struct {
			User struct {
				IsOnline bool `json:"is_online"`
			} `json:"user"`
		}
		decodeJSON(t, raw, &out)
		if out.User.IsOnline {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("is_online still false after 2s of authenticated requests (body: %s)", raw)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestUsers_AvatarDelete(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	// Upload avatar
	resp, raw := uploadAvatar(t, token, "avatar.png", "image/png", tinyPNG(t))
	mustStatus(t, resp, raw, http.StatusOK)
	var upload struct {
		User struct {
			AvatarURL *string `json:"avatar_url"`
		} `json:"user"`
	}
	decodeJSON(t, raw, &upload)
	if upload.User.AvatarURL == nil {
		t.Fatalf("upload: avatar_url is nil; body: %s", raw)
	}
	avatarURL := *upload.User.AvatarURL

	// Delete avatar
	resp, raw = doJSON(t, http.MethodDelete, "/api/v1/users/me/avatar", token, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	// /users/me now has null avatar_url
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me", token, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var me struct {
		User struct {
			AvatarURL *string `json:"avatar_url"`
		} `json:"user"`
	}
	decodeJSON(t, raw, &me)
	if me.User.AvatarURL != nil {
		t.Errorf("after delete, avatar_url = %v; want nil", *me.User.AvatarURL)
	}

	// The file URL itself 404s
	resp, raw = doJSON(t, http.MethodGet, avatarURL, "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)

	// Deleting again → 404
	resp, raw = doJSON(t, http.MethodDelete, "/api/v1/users/me/avatar", token, "")
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

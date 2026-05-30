//go:build integration

package itest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

var (
	db          *sql.DB
	srv         *httptest.Server
	uploadRoot  string
	avatarsRoot string
)

// truncate wipes all test data between tests. Order doesn't matter — CASCADE handles FKs.
func truncate(t *testing.T) {
	t.Helper()
	const q = `TRUNCATE files, game_players, games, relationships, oauth_providers, users RESTART IDENTITY CASCADE`
	if _, err := db.Exec(q); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// doJSON sends a JSON request and returns the response + body bytes.
// body may be "" for GET/DELETE. token may be "" to skip auth header.
func doJSON(t *testing.T, method, path, token, body string) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp, raw
}

func decodeJSON(t *testing.T, raw []byte, v any) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode JSON: %v (body: %s)", err, raw)
	}
}

// registerUser creates a user and returns its id + JWT.
func registerUser(t *testing.T, username, email, password string) (uuid.UUID, string) {
	t.Helper()
	body := fmt.Sprintf(`{"username":%q,"email":%q,"password":%q}`, username, email, password)
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/auth/register", "", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status %d body %s", username, resp.StatusCode, raw)
	}
	var out struct {
		User  struct {
			ID uuid.UUID `json:"id"`
		} `json:"user"`
		Token string `json:"token"`
	}
	decodeJSON(t, raw, &out)
	return out.User.ID, out.Token
}

// mustStatus asserts the response had the expected status, dumping the body if not.
func mustStatus(t *testing.T, resp *http.Response, raw []byte, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("status = %d, want %d (body: %s)", resp.StatusCode, want, raw)
	}
}

//go:build integration

package itest

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

// 1x1 transparent PNG, base64-encoded.
const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	out, err := base64.StdEncoding.DecodeString(tinyPNGBase64)
	if err != nil {
		t.Fatalf("decode tiny png: %v", err)
	}
	return out
}

// uploadAvatar posts a multipart form with the given file part.
func uploadAvatar(t *testing.T, token, filename, contentType string, body []byte) (*http.Response, []byte) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	hdr := make(map[string][]string)
	hdr["Content-Disposition"] = []string{`form-data; name="avatar"; filename="` + filename + `"`}
	hdr["Content-Type"] = []string{contentType}
	part, err := mw.CreatePart(hdr)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("write part: %v", err)
	}
	mw.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/users/me/avatar", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do upload: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp, raw
}

func TestUploads_DirectoryListingBlocked(t *testing.T) {
	resp, raw := doJSON(t, http.MethodGet, "/uploads/", "", "")
	mustStatus(t, resp, raw, http.StatusNotFound)
}

func TestUploads_AvatarRoundTrip(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	resp, raw := uploadAvatar(t, token, "avatar.png", "image/png", tinyPNG(t))
	mustStatus(t, resp, raw, http.StatusOK)

	var out struct {
		User struct {
			AvatarURL *string `json:"avatar_url"`
		} `json:"user"`
	}
	decodeJSON(t, raw, &out)
	if out.User.AvatarURL == nil {
		t.Fatalf("avatar_url is nil; body: %s", raw)
	}

	// Fetch the file back.
	fetchResp, fetchRaw := doJSON(t, http.MethodGet, *out.User.AvatarURL, "", "")
	mustStatus(t, fetchResp, fetchRaw, http.StatusOK)
	if !bytes.HasPrefix(fetchRaw, []byte("\x89PNG")) {
		t.Errorf("served file is not a PNG; first bytes: % x", fetchRaw[:min(8, len(fetchRaw))])
	}
}

func TestUploads_TooLarge(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	// 6 MB of zeros prefixed with a PNG header so MIME sniff would pass.
	body := append(tinyPNG(t), bytes.Repeat([]byte{0}, 6*1024*1024)...)
	resp, raw := uploadAvatar(t, token, "big.png", "image/png", body)
	if resp.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 for oversized upload, got 200 (body: %s)", raw)
	}
}

func TestUploads_RejectsNonImage(t *testing.T) {
	truncate(t)
	_, token := registerUser(t, "alice", "alice@example.com", "secret12")

	resp, raw := uploadAvatar(t, token, "notes.txt", "text/plain", []byte("hello world, this is not an image at all"))
	mustStatus(t, resp, raw, http.StatusUnprocessableEntity)
}

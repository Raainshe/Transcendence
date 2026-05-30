//go:build integration

package itest

import (
	"bytes"
	"net/http"
	"testing"
)

func TestBodyLimit_OneMBCap(t *testing.T) {
	truncate(t)

	// Build a 2 MB JSON body. Even though the parser would reject it as bad JSON,
	// the MaxBytesReader middleware should fail first.
	big := bytes.Repeat([]byte("a"), 2*1024*1024)
	body := `{"username":"alice","email":"alice@example.com","password":"secret12","filler":"` + string(big) + `"}`

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/auth/register", "", body)

	// http.MaxBytesReader causes the JSON decode to fail with a "request body too large"
	// error, which our handler returns as 400. We just want to verify it does NOT succeed
	// and does NOT get all the way through to user creation.
	if resp.StatusCode == http.StatusCreated {
		t.Errorf("oversized body was accepted (status 201); want rejection — body: %s", raw)
	}
}

//go:build integration

package itest

import (
	"net/http"
	"testing"
)

func TestFriends_FullCycle(t *testing.T) {
	truncate(t)
	aliceID, aliceToken := registerUser(t, "alice", "alice@example.com", "secret12")
	bobID, bobToken := registerUser(t, "bob", "bob@example.com", "secret12")

	// Alice sends request to Bob
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/users/me/friends/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusCreated)

	// Bob accepts
	resp, raw = doJSON(t, http.MethodPatch, "/api/v1/users/me/friends/"+aliceID.String(), bobToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	// Both see each other in friends
	for name, tok := range map[string]string{"alice": aliceToken, "bob": bobToken} {
		resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/me/friends", tok, "")
		mustStatus(t, resp, raw, http.StatusOK)
		var out struct {
			Friends []map[string]any `json:"friends"`
		}
		decodeJSON(t, raw, &out)
		if len(out.Friends) != 1 {
			t.Errorf("%s friends count = %d, want 1 (body: %s)", name, len(out.Friends), raw)
		}
	}

	// Alice removes the friendship
	resp, raw = doJSON(t, http.MethodDelete, "/api/v1/users/me/friends/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	// Both lists empty
	for name, tok := range map[string]string{"alice": aliceToken, "bob": bobToken} {
		resp, raw := doJSON(t, http.MethodGet, "/api/v1/users/me/friends", tok, "")
		mustStatus(t, resp, raw, http.StatusOK)
		var out struct {
			Friends []map[string]any `json:"friends"`
		}
		decodeJSON(t, raw, &out)
		if len(out.Friends) != 0 {
			t.Errorf("%s friends count after remove = %d, want 0", name, len(out.Friends))
		}
	}
}

func TestFriends_PendingShownToReceiver(t *testing.T) {
	truncate(t)
	aliceID, aliceToken := registerUser(t, "alice", "alice@example.com", "secret12")
	bobID, bobToken := registerUser(t, "bob", "bob@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/users/me/friends/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusCreated)

	// Alice (sender) has no pending requests
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me/friends/requests", aliceToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var aliceOut struct {
		Requests []map[string]any `json:"requests"`
	}
	decodeJSON(t, raw, &aliceOut)
	if len(aliceOut.Requests) != 0 {
		t.Errorf("alice (sender) sees %d pending; want 0", len(aliceOut.Requests))
	}

	// Bob sees Alice
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me/friends/requests", bobToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var bobOut struct {
		Requests []struct {
			ID string `json:"id"`
		} `json:"requests"`
	}
	decodeJSON(t, raw, &bobOut)
	if len(bobOut.Requests) != 1 || bobOut.Requests[0].ID != aliceID.String() {
		t.Errorf("bob pending = %+v, want [alice]", bobOut.Requests)
	}
}

func TestBlock_PreventsDuplicate(t *testing.T) {
	truncate(t)
	_, aliceToken := registerUser(t, "alice", "alice@example.com", "secret12")
	bobID, _ := registerUser(t, "bob", "bob@example.com", "secret12")

	resp, raw := doJSON(t, http.MethodPost, "/api/v1/users/me/block/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	resp, raw = doJSON(t, http.MethodPost, "/api/v1/users/me/block/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusConflict)
}

func TestBlock_RemovesAcceptedFriendship(t *testing.T) {
	truncate(t)
	aliceID, aliceToken := registerUser(t, "alice", "alice@example.com", "secret12")
	bobID, bobToken := registerUser(t, "bob", "bob@example.com", "secret12")

	// Become friends
	resp, raw := doJSON(t, http.MethodPost, "/api/v1/users/me/friends/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusCreated)
	resp, raw = doJSON(t, http.MethodPatch, "/api/v1/users/me/friends/"+aliceID.String(), bobToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	// Alice blocks Bob
	resp, raw = doJSON(t, http.MethodPost, "/api/v1/users/me/block/"+bobID.String(), aliceToken, "")
	mustStatus(t, resp, raw, http.StatusNoContent)

	// Bob's friends list no longer contains Alice
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me/friends", bobToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var bobOut struct {
		Friends []map[string]any `json:"friends"`
	}
	decodeJSON(t, raw, &bobOut)
	if len(bobOut.Friends) != 0 {
		t.Errorf("after block, bob still sees %d friends; want 0", len(bobOut.Friends))
	}

	// Bob's blocked list is empty (only Alice sees the block on her side)
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me/blocks", bobToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var bobBlocks struct {
		Blocked []map[string]any `json:"blocked"`
	}
	decodeJSON(t, raw, &bobBlocks)
	if len(bobBlocks.Blocked) != 0 {
		t.Errorf("bob.blocked = %d, want 0", len(bobBlocks.Blocked))
	}

	// Alice's blocked list has Bob
	resp, raw = doJSON(t, http.MethodGet, "/api/v1/users/me/blocks", aliceToken, "")
	mustStatus(t, resp, raw, http.StatusOK)
	var aliceBlocks struct {
		Blocked []struct {
			ID string `json:"id"`
		} `json:"blocked"`
	}
	decodeJSON(t, raw, &aliceBlocks)
	if len(aliceBlocks.Blocked) != 1 || aliceBlocks.Blocked[0].ID != bobID.String() {
		t.Errorf("alice.blocked = %+v, want [bob]", aliceBlocks.Blocked)
	}
}

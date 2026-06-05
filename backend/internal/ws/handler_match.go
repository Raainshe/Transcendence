package ws

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"backend/internal/model"
)

func (h *Handler) handlePlayerEliminated(c *Client, room string, rawPayload json.RawMessage) {
	gameID, ok := h.parseMatchRoom(c, room)
	if !ok {
		return
	}
	if !h.verifyMatchParticipant(c, gameID) {
		return
	}
	if h.hub.Lifecycle().IsFinished(gameID) {
		return
	}

	var upload model.PlayerEliminatedUpload
	if err := json.Unmarshal(rawPayload, &upload); err != nil {
		h.sendError(c, "invalid elimination")
		return
	}
	if err := ValidatePlayerEliminatedUpload(upload); err != nil {
		h.sendError(c, "invalid elimination")
		return
	}

	h.processElimination(c, gameID, c.UserID(), upload)
}

func (h *Handler) processElimination(
	c *Client,
	gameID, userID uuid.UUID,
	upload model.PlayerEliminatedUpload,
) {
	if h.hub.Lifecycle().IsFinished(gameID) {
		return
	}

	h.ensureLifecycle(gameID)

	placement, aliveRemaining, registered := h.hub.Lifecycle().RegisterElimination(gameID, userID, upload.Reason)
	if !registered {
		return
	}

	board := ""
	if cached, ok := h.hub.matchStates.Get(gameID, userID); ok {
		board = cached.Board
	}
	h.hub.StorePlayerState(gameID, userID, CachedPlayerState{
		UserID: userID,
		Score:  upload.Score,
		Lines:  upload.Lines,
		Level:  upload.Level,
		Alive:  false,
		Board:  board,
	})

	elimEnv, err := NewEnvelope(TypePlayerEliminated, model.PlayerEliminatedBroadcast{
		UserID:    userID.String(),
		Reason:    upload.Reason,
		Placement: placement,
	})
	if err == nil {
		h.hub.BroadcastMatch(gameID, elimEnv)
	}

	stateEnv, err := NewEnvelope(TypePlayerState, PlayerStateBroadcast{
		UserID: userID.String(),
		Score:  upload.Score,
		Lines:  upload.Lines,
		Level:  upload.Level,
		Alive:  false,
		Board:  board,
	})
	if err == nil {
		h.hub.BroadcastMatch(gameID, stateEnv)
	}

	if aliveRemaining == 1 {
		if survivor, ok := h.hub.Lifecycle().Survivor(gameID); ok {
			h.finishMatch(gameID, &survivor, false)
		}
		return
	}
	if aliveRemaining == 0 {
		h.finishMatch(gameID, nil, true)
	}
}

func (h *Handler) ensureLifecycle(gameID uuid.UUID) {
	if len(h.hub.Lifecycle().PlayerIDs(gameID)) > 0 {
		return
	}
	players, err := h.games.ListMatchPlayers(context.Background(), gameID)
	if err != nil {
		return
	}
	ids := make([]uuid.UUID, 0, len(players))
	for _, p := range players {
		ids = append(ids, p.UserID)
	}
	h.hub.Lifecycle().InitMatch(gameID, ids)
}

func (h *Handler) finishMatch(gameID uuid.UUID, survivorID *uuid.UUID, allEliminated bool) {
	if h.matches == nil || h.hub.Lifecycle().IsFinished(gameID) {
		return
	}

	stats := make([]model.PlayerMatchStats, 0)
	for _, st := range h.hub.matchStates.List(gameID) {
		stats = append(stats, model.PlayerMatchStats{
			UserID: st.UserID,
			Score:  st.Score,
			Lines:  st.Lines,
			Level:  st.Level,
		})
	}

	eliminations := make([]model.PlayerElimination, 0, len(h.hub.Lifecycle().EliminatedEntries(gameID)))
	for _, e := range h.hub.Lifecycle().EliminatedEntries(gameID) {
		eliminations = append(eliminations, model.PlayerElimination{
			UserID:    e.UserID,
			Reason:    e.Reason,
			Placement: e.Placement,
		})
	}

	payload, err := h.matches.EndMatch(context.Background(), model.EndMatchInput{
		GameID:        gameID,
		SurvivorID:    survivorID,
		AllEliminated: allEliminated,
		Stats:         stats,
		Eliminations:  eliminations,
	})
	if err != nil {
		return
	}

	h.hub.Lifecycle().MarkFinished(gameID)

	endedEnv, err := NewEnvelope(TypeMatchEnded, payload)
	if err != nil {
		return
	}
	h.hub.BroadcastMatch(gameID, endedEnv)
	h.hub.EvictMatch(gameID)
}

func (h *Handler) parseMatchRoom(c *Client, room string) (uuid.UUID, bool) {
	if len(room) < 7 || room[:6] != "match:" {
		h.sendError(c, "invalid room")
		return uuid.Nil, false
	}
	gameID, err := uuid.Parse(room[6:])
	if err != nil {
		h.sendError(c, "invalid match id")
		return uuid.Nil, false
	}
	return gameID, true
}

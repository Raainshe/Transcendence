package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

var (
	ErrMatchAlreadyFinished = errors.New("match is already finished")
	ErrMatchNotInProgress   = errors.New("match is not in progress")
)

func (s *MatchService) EndMatch(ctx context.Context, in model.EndMatchInput) (*model.MatchEndedPayload, error) {
	game, err := s.games.FindByID(ctx, in.GameID)
	if err != nil {
		return nil, err
	}
	if game.Status == "finished" {
		return nil, ErrMatchAlreadyFinished
	}
	if game.Status != "in_progress" {
		return nil, ErrMatchNotInProgress
	}

	roster, err := s.games.ListMatchPlayers(ctx, in.GameID)
	if err != nil {
		return nil, err
	}

	statsByUser := make(map[uuid.UUID]model.PlayerMatchStats, len(in.Stats))
	for _, st := range in.Stats {
		statsByUser[st.UserID] = st
	}

	elimByUser := make(map[uuid.UUID]model.PlayerElimination, len(in.Eliminations))
	for _, e := range in.Eliminations {
		elimByUser[e.UserID] = e
	}

	var winnerID *uuid.UUID
	finalPlacements := make(map[uuid.UUID]int, len(roster))

	if in.SurvivorID != nil {
		winnerID = in.SurvivorID
		finalPlacements[*in.SurvivorID] = 1
		for _, e := range in.Eliminations {
			finalPlacements[e.UserID] = e.Placement
		}
	} else if in.AllEliminated {
		winnerID = pickHighestScoreWinner(statsByUser, roster)
		if winnerID != nil {
			finalPlacements[*winnerID] = 1
			next := 2
			for _, e := range in.Eliminations {
				if e.UserID == *winnerID {
					continue
				}
				finalPlacements[e.UserID] = next
				next++
			}
		} else {
			for _, e := range in.Eliminations {
				finalPlacements[e.UserID] = e.Placement
			}
		}
	}

	now := time.Now().UTC()
	dbPlayers := make([]model.GamePlayer, 0, len(roster))
	endedPlayers := make([]model.MatchEndedPlayer, 0, len(roster))

	for _, rp := range roster {
		st, ok := statsByUser[rp.UserID]
		if !ok {
			st = model.PlayerMatchStats{UserID: rp.UserID, Level: 1}
		}
		placement := finalPlacements[rp.UserID]
		if placement == 0 {
			placement = len(roster)
		}
		isWinner := winnerID != nil && rp.UserID == *winnerID

		var placementPtr *int
		if placement > 0 {
			placementPtr = &placement
		}

		dbPlayers = append(dbPlayers, model.GamePlayer{
			GameID:       in.GameID,
			UserID:       rp.UserID,
			Score:        st.Score,
			LinesCleared: st.Lines,
			LevelReached: st.Level,
			Placement:    placementPtr,
			IsWinner:     isWinner,
		})

		var reasonPtr *string
		if elim, ok := elimByUser[rp.UserID]; ok && elim.Reason != "" {
			reasonPtr = &elim.Reason
		}

		endedPlayers = append(endedPlayers, model.MatchEndedPlayer{
			UserID:            rp.UserID,
			Username:          rp.Username,
			Score:             st.Score,
			Lines:             st.Lines,
			Level:             st.Level,
			Placement:         placement,
			IsWinner:          isWinner,
			EliminationReason: reasonPtr,
		})
	}

	if err := s.games.FinishMultiplayerMatch(ctx, in.GameID, now, dbPlayers); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrMatchAlreadyFinished
		}
		return nil, err
	}

	for _, p := range dbPlayers {
		err := s.gamification.OnGameEnd(ctx, p.UserID, p, *game)
		if err != nil {
			log.Printf("achievement check failed for user %s: %v", p.UserID, err)
		}
	}

	return &model.MatchEndedPayload{
		WinnerUserID: winnerID,
		Players:      endedPlayers,
	}, nil
}

func pickHighestScoreWinner(stats map[uuid.UUID]model.PlayerMatchStats, roster []model.MatchPlayerView) *uuid.UUID {
	var (
		bestID    *uuid.UUID
		bestScore = -1
		tie       bool
	)
	for _, rp := range roster {
		st, ok := stats[rp.UserID]
		if !ok {
			continue
		}
		if st.Score > bestScore {
			id := rp.UserID
			bestID = &id
			bestScore = st.Score
			tie = false
		} else if st.Score == bestScore && bestScore >= 0 {
			tie = true
		}
	}
	if tie || bestID == nil {
		return nil
	}
	return bestID
}

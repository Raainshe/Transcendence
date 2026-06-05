package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/testutil"
)

func TestMatchService_EndMatch_SurvivorWins(t *testing.T) {
	gameID := uuid.New()
	hostID := uuid.New()
	guestID := uuid.New()

	games := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: id, Mode: "multiplayer", Status: "in_progress"}, nil
		},
		ListMatchPlayersFn: func(_ context.Context, _ uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{
				{UserID: hostID, Username: "host"},
				{UserID: guestID, Username: "guest"},
			}, nil
		},
		FinishMultiplayerMatchFn: func(_ context.Context, gid uuid.UUID, _ time.Time, players []model.GamePlayer) error {
			if gid != gameID {
				t.Fatalf("gameID = %v", gid)
			}
			if len(players) != 2 {
				t.Fatalf("players = %d", len(players))
			}
			return nil
		},
	}

	svc := service.NewMatchService(games, &testutil.MockLobbyRepo{})
	payload, err := svc.EndMatch(context.Background(), model.EndMatchInput{
		GameID:     gameID,
		SurvivorID: &guestID,
		Stats: []model.PlayerMatchStats{
			{UserID: hostID, Score: 100, Lines: 1, Level: 1},
			{UserID: guestID, Score: 500, Lines: 5, Level: 2},
		},
		Eliminations: []model.PlayerElimination{
			{UserID: hostID, Reason: "topOut", Placement: 2},
		},
	})
	if err != nil {
		t.Fatalf("EndMatch: %v", err)
	}
	if payload.WinnerUserID == nil || *payload.WinnerUserID != guestID {
		t.Fatalf("winner = %+v", payload.WinnerUserID)
	}
}

func TestMatchService_EndMatch_AllDeadHighestScore(t *testing.T) {
	gameID := uuid.New()
	a := uuid.New()
	b := uuid.New()

	var saved []model.GamePlayer
	games := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: id, Mode: "multiplayer", Status: "in_progress"}, nil
		},
		ListMatchPlayersFn: func(_ context.Context, _ uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{
				{UserID: a, Username: "a"},
				{UserID: b, Username: "b"},
			}, nil
		},
		FinishMultiplayerMatchFn: func(_ context.Context, _ uuid.UUID, _ time.Time, players []model.GamePlayer) error {
			saved = players
			return nil
		},
	}

	svc := service.NewMatchService(games, &testutil.MockLobbyRepo{})
	payload, err := svc.EndMatch(context.Background(), model.EndMatchInput{
		GameID:        gameID,
		AllEliminated: true,
		Stats: []model.PlayerMatchStats{
			{UserID: a, Score: 900, Lines: 9, Level: 3},
			{UserID: b, Score: 400, Lines: 4, Level: 2},
		},
		Eliminations: []model.PlayerElimination{
			{UserID: a, Reason: "topOut", Placement: 2},
			{UserID: b, Reason: "topOut", Placement: 1},
		},
	})
	if err != nil {
		t.Fatalf("EndMatch: %v", err)
	}
	if payload.WinnerUserID == nil || *payload.WinnerUserID != a {
		t.Fatalf("winner = %+v, want %v", payload.WinnerUserID, a)
	}
	if !saved[0].IsWinner && !saved[1].IsWinner {
		// one must be winner in saved slice
		found := false
		for _, p := range saved {
			if p.UserID == a && p.IsWinner {
				found = true
			}
		}
		if !found {
			t.Fatal("expected player a to be winner in DB payload")
		}
	}
}

func TestMatchService_EndMatch_AllDeadTieNoWinner(t *testing.T) {
	gameID := uuid.New()
	a := uuid.New()
	b := uuid.New()

	games := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: id, Mode: "multiplayer", Status: "in_progress"}, nil
		},
		ListMatchPlayersFn: func(_ context.Context, _ uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{
				{UserID: a, Username: "a"},
				{UserID: b, Username: "b"},
			}, nil
		},
		FinishMultiplayerMatchFn: func(_ context.Context, _ uuid.UUID, _ time.Time, players []model.GamePlayer) error {
			for _, p := range players {
				if p.IsWinner {
					t.Fatal("no winner expected on score tie")
				}
			}
			return nil
		},
	}

	svc := service.NewMatchService(games, &testutil.MockLobbyRepo{})
	payload, err := svc.EndMatch(context.Background(), model.EndMatchInput{
		GameID:        gameID,
		AllEliminated: true,
		Stats: []model.PlayerMatchStats{
			{UserID: a, Score: 500, Lines: 5, Level: 2},
			{UserID: b, Score: 500, Lines: 5, Level: 2},
		},
		Eliminations: []model.PlayerElimination{
			{UserID: a, Reason: "topOut", Placement: 2},
			{UserID: b, Reason: "topOut", Placement: 1},
		},
	})
	if err != nil {
		t.Fatalf("EndMatch: %v", err)
	}
	if payload.WinnerUserID != nil {
		t.Fatalf("winner = %+v, want nil", payload.WinnerUserID)
	}
}

func TestMatchService_EndMatch_AlreadyFinished(t *testing.T) {
	gameID := uuid.New()
	games := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: id, Status: "finished"}, nil
		},
	}
	svc := service.NewMatchService(games, &testutil.MockLobbyRepo{})
	_, err := svc.EndMatch(context.Background(), model.EndMatchInput{GameID: gameID})
	if err != service.ErrMatchAlreadyFinished {
		t.Fatalf("err = %v", err)
	}
}

func TestMatchService_EndMatch_DBNotFound(t *testing.T) {
	gameID := uuid.New()
	a := uuid.New()
	games := &testutil.MockGameRepo{
		FindByIDFn: func(_ context.Context, id uuid.UUID) (*model.Game, error) {
			return &model.Game{ID: id, Mode: "multiplayer", Status: "in_progress"}, nil
		},
		ListMatchPlayersFn: func(_ context.Context, _ uuid.UUID) ([]model.MatchPlayerView, error) {
			return []model.MatchPlayerView{{UserID: a, Username: "a"}}, nil
		},
		FinishMultiplayerMatchFn: func(context.Context, uuid.UUID, time.Time, []model.GamePlayer) error {
			return repository.ErrNotFound
		},
	}
	svc := service.NewMatchService(games, &testutil.MockLobbyRepo{})
	_, err := svc.EndMatch(context.Background(), model.EndMatchInput{
		GameID:     gameID,
		SurvivorID: &a,
		Stats:      []model.PlayerMatchStats{{UserID: a, Level: 1}},
	})
	if err != service.ErrMatchAlreadyFinished {
		t.Fatalf("err = %v", err)
	}
}

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/testutil"
)

func newUserSvc(t *testing.T, users *testutil.MockUserRepo, files *testutil.MockFileRepo, rels *testutil.MockRelationshipRepo) *service.UserService {
	t.Helper()
	if users == nil {
		users = &testutil.MockUserRepo{}
	}
	if files == nil {
		files = &testutil.MockFileRepo{}
	}
	if rels == nil {
		rels = &testutil.MockRelationshipRepo{}
	}
	return service.NewUserService(users, files, rels, t.TempDir())
}

func TestUserService_GetByID(t *testing.T) {
	existing := testutil.NewTestUser()

	tests := []struct {
		name       string
		findByIDFn func(context.Context, uuid.UUID) (*model.User, error)
		wantErr    error
	}{
		{
			name: "found",
			findByIDFn: func(_ context.Context, _ uuid.UUID) (*model.User, error) {
				return existing, nil
			},
		},
		{
			name:    "not found",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserSvc(t, &testutil.MockUserRepo{FindByIDFn: tt.findByIDFn}, nil, nil)
			user, err := svc.GetByID(context.Background(), existing.ID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user == nil {
				t.Error("expected non-nil user")
			}
		})
	}
}

func TestUserService_ListUsers(t *testing.T) {
	var capturedLimit int

	tests := []struct {
		name        string
		inputLimit  int
		wantLimit   int
	}{
		{name: "zero limit defaults to 20", inputLimit: 0, wantLimit: 20},
		{name: "limit over 100 clamped to 100", inputLimit: 200, wantLimit: 100},
		{name: "limit within range used as-is", inputLimit: 50, wantLimit: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{
				ListFn: func(_ context.Context, limit, _ int) ([]model.User, error) {
					capturedLimit = limit
					return []model.User{}, nil
				},
			}
			svc := newUserSvc(t, repo, nil, nil)
			_, _, err := svc.ListUsers(context.Background(), tt.inputLimit, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if capturedLimit != tt.wantLimit {
				t.Errorf("called List with limit=%d, want %d", capturedLimit, tt.wantLimit)
			}
		})
	}
}

func TestUserService_UpdateMe(t *testing.T) {
	self := testutil.NewTestUser()
	other := testutil.NewTestUser()
	takenUsername := "taken"

	tests := []struct {
		name             string
		req              model.UpdateUserRequest
		findByUsernameFn func(context.Context, string) (*model.User, error)
		wantErr          error
	}{
		{
			name: "nil username skips uniqueness check",
			req:  model.UpdateUserRequest{},
		},
		{
			name: "new username available",
			req:  model.UpdateUserRequest{Username: ptr("newname")},
		},
		{
			name: "username taken by another user",
			req:  model.UpdateUserRequest{Username: &takenUsername},
			findByUsernameFn: func(_ context.Context, _ string) (*model.User, error) {
				return other, nil
			},
			wantErr: service.ErrUsernameTaken,
		},
		{
			name: "username same as own — no conflict",
			req:  model.UpdateUserRequest{Username: &self.Username},
			findByUsernameFn: func(_ context.Context, _ string) (*model.User, error) {
				return self, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testutil.MockUserRepo{
				FindByUsernameFn: tt.findByUsernameFn,
				UpdateFn: func(_ context.Context, id uuid.UUID, _ model.UpdateUserRequest) (*model.User, error) {
					return &model.User{ID: id}, nil
				},
			}
			svc := newUserSvc(t, repo, nil, nil)
			_, err := svc.UpdateMe(context.Background(), self.ID, tt.req)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUserService_SendFriendRequest(t *testing.T) {
	fromID := uuid.New()
	toID := uuid.New()
	existing := &model.Relationship{ID: uuid.New(), Status: model.RelationshipPending}

	tests := []struct {
		name    string
		from    uuid.UUID
		to      uuid.UUID
		findFn  func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantErr error
	}{
		{
			name: "success",
			from: fromID, to: toID,
		},
		{
			name:    "self-request",
			from:    fromID, to: fromID,
			wantErr: service.ErrSelfRelationship,
		},
		{
			name: "relationship already exists",
			from: fromID, to: toID,
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return existing, nil
			},
			wantErr: service.ErrRelationshipExists,
		},
		{
			name: "db error on find",
			from: fromID, to: toID,
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return nil, errors.New("db error")
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relRepo := &testutil.MockRelationshipRepo{FindFn: tt.findFn}
			svc := newUserSvc(t, nil, nil, relRepo)
			err := svc.SendFriendRequest(context.Background(), tt.from, tt.to)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if errors.Is(tt.wantErr, service.ErrSelfRelationship) || errors.Is(tt.wantErr, service.ErrRelationshipExists) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("error = %v, want %v", err, tt.wantErr)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUserService_AcceptFriendRequest(t *testing.T) {
	accepterID := uuid.New()
	requesterID := uuid.New()
	relID := uuid.New()

	makePending := func() *model.Relationship {
		return &model.Relationship{ID: relID, Status: model.RelationshipPending}
	}

	tests := []struct {
		name              string
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantErr           error
		wantUpdate        bool
	}{
		{
			name: "success",
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return makePending(), nil
			},
			wantUpdate: true,
		},
		{
			name:    "pending request not found",
			wantErr: repository.ErrNotFound,
		},
		{
			name: "relationship is already accepted",
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipAccepted}, nil
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name: "relationship is blocked",
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updatedStatus string
			relRepo := &testutil.MockRelationshipRepo{
				FindDirectionalFn: tt.findDirectionalFn,
				UpdateStatusFn: func(_ context.Context, _ uuid.UUID, status string) error {
					updatedStatus = status
					return nil
				},
			}
			svc := newUserSvc(t, nil, nil, relRepo)
			err := svc.AcceptFriendRequest(context.Background(), accepterID, requesterID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantUpdate && updatedStatus != model.RelationshipAccepted {
				t.Errorf("UpdateStatus called with %q, want %q", updatedStatus, model.RelationshipAccepted)
			}
		})
	}
}

func TestUserService_RemoveFriend(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name    string
		findFn  func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantErr error
		wantDel bool
	}{
		{
			name: "remove accepted friendship",
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipAccepted}, nil
			},
			wantDel: true,
		},
		{
			name: "reject pending request",
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipPending}, nil
			},
			wantDel: true,
		},
		{
			name: "blocked relationship is not a friendship",
			findFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name:    "not found",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted := false
			relRepo := &testutil.MockRelationshipRepo{
				FindFn: tt.findFn,
				DeleteFn: func(_ context.Context, _ uuid.UUID) error {
					deleted = true
					return nil
				},
			}
			svc := newUserSvc(t, nil, nil, relRepo)
			err := svc.RemoveFriend(context.Background(), userID, otherID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantDel && !deleted {
				t.Error("expected Delete to be called")
			}
		})
	}
}

func TestUserService_BlockUser(t *testing.T) {
	blockerID := uuid.New()
	targetID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name              string
		blocker           uuid.UUID
		target            uuid.UUID
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantErr           error
		wantUpdateStatus  string
		wantCreate        bool
		wantDelete        bool
	}{
		{
			name:    "self-block",
			blocker: blockerID, target: blockerID,
			wantErr: service.ErrSelfRelationship,
		},
		{
			name:    "already blocked",
			blocker: blockerID, target: targetID,
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantErr: service.ErrRelationshipExists,
		},
		{
			name:    "existing accepted → upgrade to blocked",
			blocker: blockerID, target: targetID,
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipAccepted}, nil
			},
			wantUpdateStatus: model.RelationshipBlocked,
		},
		{
			name:    "existing pending same direction → upgrade to blocked",
			blocker: blockerID, target: targetID,
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipPending}, nil
			},
			wantUpdateStatus: model.RelationshipBlocked,
		},
		{
			name:    "no existing row → delete reverse if any, then create",
			blocker: blockerID, target: targetID,
			wantCreate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updatedStatus string
			created := false
			deleted := false
			relRepo := &testutil.MockRelationshipRepo{
				FindDirectionalFn: tt.findDirectionalFn,
				UpdateStatusFn: func(_ context.Context, _ uuid.UUID, status string) error {
					updatedStatus = status
					return nil
				},
				CreateFn: func(_ context.Context, _, _ uuid.UUID, _ string) (*model.Relationship, error) {
					created = true
					return &model.Relationship{}, nil
				},
				DeleteFn: func(_ context.Context, _ uuid.UUID) error {
					deleted = true
					return nil
				},
			}
			svc := newUserSvc(t, nil, nil, relRepo)
			err := svc.BlockUser(context.Background(), tt.blocker, tt.target)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantUpdateStatus != "" && updatedStatus != tt.wantUpdateStatus {
				t.Errorf("UpdateStatus called with %q, want %q", updatedStatus, tt.wantUpdateStatus)
			}
			if tt.wantCreate && !created {
				t.Error("expected Create to be called")
			}
			_ = deleted // reverse-delete is optional path
		})
	}
}

func TestUserService_UnblockUser(t *testing.T) {
	blockerID := uuid.New()
	targetID := uuid.New()
	relID := uuid.New()

	tests := []struct {
		name              string
		findDirectionalFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Relationship, error)
		wantErr           error
		wantDel           bool
	}{
		{
			name: "success",
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipBlocked}, nil
			},
			wantDel: true,
		},
		{
			name: "relationship is not a block",
			findDirectionalFn: func(_ context.Context, _, _ uuid.UUID) (*model.Relationship, error) {
				return &model.Relationship{ID: relID, Status: model.RelationshipAccepted}, nil
			},
			wantErr: repository.ErrNotFound,
		},
		{
			name:    "no relationship found",
			wantErr: repository.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted := false
			relRepo := &testutil.MockRelationshipRepo{
				FindDirectionalFn: tt.findDirectionalFn,
				DeleteFn: func(_ context.Context, _ uuid.UUID) error {
					deleted = true
					return nil
				},
			}
			svc := newUserSvc(t, nil, nil, relRepo)
			err := svc.UnblockUser(context.Background(), blockerID, targetID)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantDel && !deleted {
				t.Error("expected Delete to be called")
			}
		})
	}
}

func ptr(s string) *string { return &s }

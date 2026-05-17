package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"backend/internal/model"
	"backend/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]model.User, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	users, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.users.Count(ctx)
	return users, total, err
}

func (s *UserService) UpdateMe(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error) {
	if req.Username != nil {
		existing, err := s.users.FindByUsername(ctx, *req.Username)
		if err == nil && existing.ID != id {
			return nil, ErrUsernameTaken
		} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return nil, err
		}
	}
	return s.users.Update(ctx, id, req)
}

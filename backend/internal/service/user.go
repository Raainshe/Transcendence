package service

import (
	"context"

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

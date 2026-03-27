package users_service

import (
	"context"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func (s *UserService) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	users, err := s.usersRepository.GetUsers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get users from repository: %w", err)
	}

	return users, nil
}

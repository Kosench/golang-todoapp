package users_service

import (
	"context"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

type UserService struct {
	usersRepository UserRepository
}

type UserRepository interface {
	CreateUser(
		ctx context.Context,
		user domain.User,
	) (domain.User, error)
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
		usersRepository: repository,
	}
}

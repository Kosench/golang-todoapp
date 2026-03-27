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
	GetUsers(
		ctx context.Context,
		limit *int,
		offset *int,
	) ([]domain.User, error)
	GetUser(
		ctx context.Context,
		id int,
	) (domain.User, error)
	DeleteUser(
		ctx context.Context,
		id int,
	) error
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{
		usersRepository: repository,
	}
}

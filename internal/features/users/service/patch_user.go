package users_service

import (
	"context"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func (s *UserService) PatchUser(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error) {
	//1. получить юзера
	user, err := s.usersRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}

	//2. пропатчить юзера на уровне сервиса
	if err := user.ApplyPatch(patch); err != nil {
		return domain.User{}, fmt.Errorf("apply user patch: %w", err)
	}

	//3. сохранить пропатченного юзера в репу
	patchedUser, err := s.usersRepository.PatchUser(ctx, id, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("patch user: %w", err)
	}

	return patchedUser, nil
}

package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool"
)

func (r *UsersRepository) PatchUser(ctx context.Context, id int, user domain.User) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE todoapp.users
	SET
	    full_name=$1,
		phone_number=$2,
		version=version+1
	WHERE id=$3 AND version=$4
	RETURNING 
		id,
		version,
		full_name,
		phone_number
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		user.FullName,
		user.PhoneNumber,
		id,
		user.Version,
	)

	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Version,
		&userModel.FullName,
		&userModel.PhoneNumber,
	)
	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			exists, existsErr := r.userExists(ctx, id)
			if existsErr != nil {
				return domain.User{}, fmt.Errorf("check user existence: %w", existsErr)
			}
			if !exists {
				return domain.User{}, fmt.Errorf("user with id='%d': %w", id, core_errors.ErrNotFound)
			}

			return domain.User{}, fmt.Errorf("user with id='%d' concurrently accessed: %w", id, core_errors.ErrConflict)
		}
		if errors.Is(err, core_postgres_pool.ErrViolatesCheckConstraint) {
			return domain.User{}, fmt.Errorf("user violates database constraint: %w", core_errors.ErrInvalidArgument)
		}

		return domain.User{}, fmt.Errorf("scan error: %w", err)
	}

	userDomain := domain.NewUser(
		userModel.ID,
		userModel.Version,
		userModel.FullName,
		userModel.PhoneNumber,
	)

	return userDomain, nil
}

func (r *UsersRepository) userExists(ctx context.Context, id int) (bool, error) {
	query := `
	SELECT EXISTS(
		SELECT 1
		FROM todoapp.users
		WHERE id=$1
	);
	`

	row := r.pool.QueryRow(ctx, query, id)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("scan user exists: %w", err)
	}

	return exists, nil
}

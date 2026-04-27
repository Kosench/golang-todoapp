//go:build integration

package repository_integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
)

func TestUsersRepository_CreateUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	user, err := repo.CreateUser(ctx, domain.NewUserUninitialized("John Doe", ptr("+1234567890")))
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}

	if user.ID <= 0 {
		t.Errorf("ID = %d, want positive", user.ID)
	}
	if user.Version != 1 {
		t.Errorf("Version = %d, want 1", user.Version)
	}
	if user.FullName != "John Doe" {
		t.Errorf("FullName = %q, want %q", user.FullName, "John Doe")
	}
	if user.PhoneNumber == nil || *user.PhoneNumber != "+1234567890" {
		t.Errorf("PhoneNumber = %v, want %q", user.PhoneNumber, "+1234567890")
	}
}

func TestUsersRepository_GetUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	got, err := repo.GetUser(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUser() error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.FullName != created.FullName {
		t.Errorf("FullName = %q, want %q", got.FullName, created.FullName)
	}
}

func TestUsersRepository_GetUser_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	_, err := repo.GetUser(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestUsersRepository_GetUsers(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	first := createUser(t, ctx, "John Doe")
	second := createUser(t, ctx, "Jane Doe")

	users, err := repo.GetUsers(ctx, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetUsers() error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
	if users[0].ID != first.ID {
		t.Errorf("users[0].ID = %d, want %d", users[0].ID, first.ID)
	}
	if users[1].ID != second.ID {
		t.Errorf("users[1].ID = %d, want %d", users[1].ID, second.ID)
	}
}

func TestUsersRepository_PatchUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	patched := created
	patched.FullName = "Jane Doe"
	patched.PhoneNumber = nil

	got, err := repo.PatchUser(ctx, created.ID, patched)
	if err != nil {
		t.Fatalf("PatchUser() error: %v", err)
	}

	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want %q", got.FullName, "Jane Doe")
	}
	if got.PhoneNumber != nil {
		t.Errorf("PhoneNumber = %v, want nil", got.PhoneNumber)
	}
	if got.Version != created.Version+1 {
		t.Errorf("Version = %d, want %d", got.Version, created.Version+1)
	}
}

func TestUsersRepository_PatchUser_Conflict(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	firstPatch := created
	firstPatch.FullName = "Jane Doe"

	updated, err := repo.PatchUser(ctx, created.ID, firstPatch)
	if err != nil {
		t.Fatalf("first PatchUser() error: %v", err)
	}

	stale := created
	stale.FullName = "Stale Name"

	_, err = repo.PatchUser(ctx, updated.ID, stale)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestUsersRepository_DeleteUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)
	created := createUser(t, ctx, "John Doe")

	if err := repo.DeleteUser(ctx, created.ID); err != nil {
		t.Fatalf("DeleteUser() error: %v", err)
	}

	_, err := repo.GetUser(ctx, created.ID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetUser() after delete error = %v, want ErrNotFound", err)
	}
}

func TestUsersRepository_DeleteUser_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := users_postgres_repository.NewUsersRepository(pool)

	err := repo.DeleteUser(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

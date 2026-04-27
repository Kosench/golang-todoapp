//go:build integration

package repository_integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
)

const missingID = 999999999

func ptr[T any](v T) *T {
	return &v
}

func cleanup(t *testing.T) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE todoapp.tasks, todoapp.users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("cleanup database: %v", err)
	}
}

func createUser(t *testing.T, ctx context.Context, fullName string) domain.User {
	t.Helper()

	repo := users_postgres_repository.NewUsersRepository(pool)

	user, err := repo.CreateUser(ctx, domain.NewUserUninitialized(fullName, ptr("+1234567890")))
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func fixedTime() time.Time {
	return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

package users_service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	users_service "github.com/Kosench/golang-todoapp/internal/features/users/service"
)

func ptr[T any](v T) *T {
	return &v
}

type MockUserRepository struct {
	CreateUserFunc func(ctx context.Context, user domain.User) (domain.User, error)
	GetUsersFunc   func(ctx context.Context, limit *int, offset *int) ([]domain.User, error)
	GetUserFunc    func(ctx context.Context, id int) (domain.User, error)
	DeleteUserFunc func(ctx context.Context, id int) error
	PatchUserFunc  func(ctx context.Context, id int, user domain.User) (domain.User, error)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	return m.CreateUserFunc(ctx, user)
}

func (m *MockUserRepository) GetUsers(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
	return m.GetUsersFunc(ctx, limit, offset)
}

func (m *MockUserRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	return m.GetUserFunc(ctx, id)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	return m.DeleteUserFunc(ctx, id)
}

func (m *MockUserRepository) PatchUser(ctx context.Context, id int, user domain.User) (domain.User, error) {
	return m.PatchUserFunc(ctx, id, user)
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &MockUserRepository{
			CreateUserFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
				user.ID = 1
				user.Version = 1
				return user, nil
			},
		}

		svc := users_service.NewUserService(repo)
		input := domain.NewUserUninitialized("John Doe", ptr("+1234567890"))
		got, err := svc.CreateUser(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 1 {
			t.Errorf("ID = %d, want 1", got.ID)
		}
		if got.FullName != "John Doe" {
			t.Errorf("FullName = %q, want 'John Doe'", got.FullName)
		}
	})

	t.Run("validation error: name too short", func(t *testing.T) {
		repo := &MockUserRepository{}

		svc := users_service.NewUserService(repo)
		input := domain.NewUserUninitialized("AB", nil) // 2 символа < 3
		_, err := svc.CreateUser(ctx, input)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repoErr := errors.New("unique constraint violation")
		repo := &MockUserRepository{
			CreateUserFunc: func(ctx context.Context, user domain.User) (domain.User, error) {
				return domain.User{}, repoErr
			},
		}

		svc := users_service.NewUserService(repo)
		input := domain.NewUserUninitialized("John Doe", nil)
		_, err := svc.CreateUser(ctx, input)

		if !errors.Is(err, repoErr) {
			t.Errorf("expected repoErr, got: %v", err)
		}
	})
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := domain.NewUser(1, 1, "John", nil)
		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return expected, nil
			},
		}

		svc := users_service.NewUserService(repo)
		got, err := svc.GetUser(ctx, 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FullName != "John" {
			t.Errorf("FullName = %q, want 'John'", got.FullName)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return domain.User{}, core_errors.ErrNotFound
			},
		}

		svc := users_service.NewUserService(repo)
		_, err := svc.GetUser(ctx, 999)

		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestUserService_GetUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		users := []domain.User{
			domain.NewUser(1, 1, "Alice", nil),
			domain.NewUser(2, 1, "Bob", nil),
		}

		repo := &MockUserRepository{
			GetUsersFunc: func(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
				return users, nil
			},
		}

		svc := users_service.NewUserService(repo)
		got, err := svc.GetUsers(ctx, ptr(10), ptr(0))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len = %d, want 2", len(got))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockUserRepository{
			GetUsersFunc: func(ctx context.Context, limit *int, offset *int) ([]domain.User, error) {
				return nil, errors.New("timeout")
			},
		}

		svc := users_service.NewUserService(repo)
		_, err := svc.GetUsers(ctx, nil, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &MockUserRepository{
			DeleteUserFunc: func(ctx context.Context, id int) error {
				return nil
			},
		}

		svc := users_service.NewUserService(repo)
		err := svc.DeleteUser(ctx, 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockUserRepository{
			DeleteUserFunc: func(ctx context.Context, id int) error {
				return core_errors.ErrNotFound
			},
		}

		svc := users_service.NewUserService(repo)
		err := svc.DeleteUser(ctx, 999)

		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestUserService_PatchUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success: update FullName", func(t *testing.T) {
		existing := domain.NewUser(1, 1, "Old Name", ptr("+1234567890"))

		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			PatchUserFunc: func(ctx context.Context, id int, user domain.User) (domain.User, error) {
				if user.FullName != "New Name" {
					t.Errorf("PatchUser got FullName=%q, want 'New Name'", user.FullName)
				}
				user.Version = 2
				return user, nil
			},
		}

		svc := users_service.NewUserService(repo)
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: ptr("New Name"), Set: true},
			domain.Nullable[string]{},
		)

		got, err := svc.PatchUser(ctx, 1, patch)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FullName != "New Name" {
			t.Errorf("FullName = %q, want 'New Name'", got.FullName)
		}
	})

	t.Run("success: delete phone", func(t *testing.T) {
		existing := domain.NewUser(1, 1, "John Doe", ptr("+1234567890"))

		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			PatchUserFunc: func(ctx context.Context, id int, user domain.User) (domain.User, error) {
				if user.PhoneNumber != nil {
					t.Errorf("PhoneNumber should be nil after deleting, got %v", *user.PhoneNumber)
				}
				return user, nil
			},
		}

		svc := users_service.NewUserService(repo)
		patch := domain.NewUserPatch(
			domain.Nullable[string]{},
			domain.Nullable[string]{Value: nil, Set: true}, // удаляем телефон
		)

		got, err := svc.PatchUser(ctx, 1, patch)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.PhoneNumber != nil {
			t.Error("PhoneNumber should be nil")
		}
	})

	t.Run("error: user not found", func(t *testing.T) {
		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return domain.User{}, core_errors.ErrNotFound
			},
		}

		svc := users_service.NewUserService(repo)
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: ptr("Name"), Set: true},
			domain.Nullable[string]{},
		)

		_, err := svc.PatchUser(ctx, 999, patch)

		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("error: invalid patch result", func(t *testing.T) {
		existing := domain.NewUser(1, 1, "John Doe", nil)

		repo := &MockUserRepository{
			GetUserFunc: func(ctx context.Context, id int) (domain.User, error) {
				return existing, nil
			},
			// PatchUserFunc не задаём — ApplyPatch упадёт на валидации
		}

		svc := users_service.NewUserService(repo)
		patch := domain.NewUserPatch(
			domain.Nullable[string]{Value: ptr("AB"), Set: true}, // слишком короткое
			domain.Nullable[string]{},
		)

		_, err := svc.PatchUser(ctx, 1, patch)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})
}

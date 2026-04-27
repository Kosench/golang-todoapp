package tasks_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	tasks_service "github.com/Kosench/golang-todoapp/internal/features/tasks/service"
)

func ptr[T any](v T) *T {
	return &v
}

type MockTaskRepository struct {
	CreateTaskFunc func(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasksFunc   func(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error)
	GetTaskFunc    func(ctx context.Context, id int) (domain.Task, error)
	DeleteTaskFunc func(ctx context.Context, id int) error
	PatchTaskFunc  func(ctx context.Context, id int, task domain.Task) (domain.Task, error)
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	return m.CreateTaskFunc(ctx, task)
}

func (m *MockTaskRepository) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	return m.GetTasksFunc(ctx, userID, limit, offset)
}

func (m *MockTaskRepository) GetTask(ctx context.Context, id int) (domain.Task, error) {
	return m.GetTaskFunc(ctx, id)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, id int) error {
	return m.DeleteTaskFunc(ctx, id)
}

func (m *MockTaskRepository) PatchTask(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
	return m.PatchTaskFunc(ctx, id, task)
}

func TestTaskService_CreateTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		// Задаём поведение мока: при вызове CreateTask вернуть задачу с ID=42
		repo := &MockTaskRepository{
			CreateTaskFunc: func(ctx context.Context, task domain.Task) (domain.Task, error) {
				// Мок имитирует БД: назначает ID и Version
				task.ID = 42
				task.Version = 1
				return task, nil
			},
		}

		svc := tasks_service.NewTaskService(repo)

		input := domain.NewTaskUninitialized("Buy milk", nil, 1, now)
		got, err := svc.CreateTask(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 42 {
			t.Errorf("ID = %d, want 42", got.ID)
		}
		if got.Title != "Buy milk" {
			t.Errorf("Title = %q, want 'Buy milk'", got.Title)
		}
	})

	t.Run("validation error: empty title", func(t *testing.T) {
		// Репо не должен вызываться — ошибка на этапе валидации
		repo := &MockTaskRepository{
			// CreateTaskFunc не задаём — если вызовется, будет panic.
			// Это гарантирует, что сервис НЕ дошёл до репозитория.
		}

		svc := tasks_service.NewTaskService(repo)

		input := domain.NewTaskUninitialized("", nil, 1, now) // пустой title
		_, err := svc.CreateTask(ctx, input)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repoErr := errors.New("connection refused")
		repo := &MockTaskRepository{
			CreateTaskFunc: func(ctx context.Context, task domain.Task) (domain.Task, error) {
				return domain.Task{}, repoErr
			},
		}

		svc := tasks_service.NewTaskService(repo)

		input := domain.NewTaskUninitialized("Valid title", nil, 1, now)
		_, err := svc.CreateTask(ctx, input)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Ошибка обёрнута через fmt.Errorf("...: %w"), поэтому проверяем через Is
		if !errors.Is(err, repoErr) {
			t.Errorf("expected repoErr in chain, got: %v", err)
		}
	})
}

func TestTaskService_GetTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		expected := domain.NewTask(1, 1, "Task", nil, false, now, nil, 1)

		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				if id != 1 {
					t.Errorf("GetTask called with id=%d, want 1", id)
				}
				return expected, nil
			},
		}

		svc := tasks_service.NewTaskService(repo)
		got, err := svc.GetTask(ctx, 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != expected.ID {
			t.Errorf("ID = %d, want %d", got.ID, expected.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				return domain.Task{}, core_errors.ErrNotFound
			},
		}

		svc := tasks_service.NewTaskService(repo)
		_, err := svc.GetTask(ctx, 999)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestTaskService_GetTasks(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success with filters", func(t *testing.T) {
		tasks := []domain.Task{
			domain.NewTask(1, 1, "Task 1", nil, false, now, nil, 1),
			domain.NewTask(2, 1, "Task 2", nil, false, now, nil, 1),
		}

		repo := &MockTaskRepository{
			GetTasksFunc: func(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
				// Проверяем, что фильтры прокидываются корректно
				if userID == nil || *userID != 1 {
					t.Errorf("userID = %v, want ptr(1)", userID)
				}
				if limit == nil || *limit != 10 {
					t.Errorf("limit = %v, want ptr(10)", limit)
				}
				return tasks, nil
			},
		}

		svc := tasks_service.NewTaskService(repo)
		got, err := svc.GetTasks(ctx, ptr(1), ptr(10), nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("len = %d, want 2", len(got))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &MockTaskRepository{
			GetTasksFunc: func(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
				return nil, errors.New("db error")
			},
		}

		svc := tasks_service.NewTaskService(repo)
		_, err := svc.GetTasks(ctx, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deletedID := 0
		repo := &MockTaskRepository{
			DeleteTaskFunc: func(ctx context.Context, id int) error {
				deletedID = id // запоминаем, с каким ID вызвали
				return nil
			},
		}

		svc := tasks_service.NewTaskService(repo)
		err := svc.DeleteTask(ctx, 42)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deletedID != 42 {
			t.Errorf("DeleteTask called with id=%d, want 42", deletedID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &MockTaskRepository{
			DeleteTaskFunc: func(ctx context.Context, id int) error {
				return core_errors.ErrNotFound
			},
		}

		svc := tasks_service.NewTaskService(repo)
		err := svc.DeleteTask(ctx, 999)

		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestTaskService_PatchTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success: update title", func(t *testing.T) {
		existingTask := domain.NewTask(1, 1, "Old title", nil, false, now, nil, 1)

		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				return existingTask, nil
			},
			PatchTaskFunc: func(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
				// Проверяем, что в репо пришла задача с обновлённым title
				if task.Title != "New title" {
					t.Errorf("PatchTask received title=%q, want 'New title'", task.Title)
				}
				task.Version = 2 // имитируем инкремент версии в БД
				return task, nil
			},
		}

		svc := tasks_service.NewTaskService(repo)

		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: ptr("New title"), Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		got, err := svc.PatchTask(ctx, 1, patch)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Title != "New title" {
			t.Errorf("Title = %q, want 'New title'", got.Title)
		}
		if got.Version != 2 {
			t.Errorf("Version = %d, want 2", got.Version)
		}
	})

	t.Run("error: task not found on get", func(t *testing.T) {
		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				return domain.Task{}, core_errors.ErrNotFound
			},
			// PatchTaskFunc не задаём — не должен вызываться
		}

		svc := tasks_service.NewTaskService(repo)
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: ptr("New"), Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		_, err := svc.PatchTask(ctx, 999, patch)

		if !errors.Is(err, core_errors.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("error: invalid patch (title too long)", func(t *testing.T) {
		existingTask := domain.NewTask(1, 1, "Old title", nil, false, now, nil, 1)

		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				return existingTask, nil
			},
			// PatchTaskFunc не задаём — ApplyPatch упадёт раньше
		}

		svc := tasks_service.NewTaskService(repo)

		// Строим патч с title из 101 символа — ApplyPatch → Validate → ошибка
		longTitle := make([]byte, 101)
		for i := range longTitle {
			longTitle[i] = 'a'
		}
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: ptr(string(longTitle)), Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		_, err := svc.PatchTask(ctx, 1, patch)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})

	t.Run("error: repository PatchTask fails", func(t *testing.T) {
		existingTask := domain.NewTask(1, 1, "Title", nil, false, now, nil, 1)

		repo := &MockTaskRepository{
			GetTaskFunc: func(ctx context.Context, id int) (domain.Task, error) {
				return existingTask, nil
			},
			PatchTaskFunc: func(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
				return domain.Task{}, core_errors.ErrConflict
			},
		}

		svc := tasks_service.NewTaskService(repo)
		patch := domain.NewTaskPatch(
			domain.Nullable[string]{Value: ptr("Updated"), Set: true},
			domain.Nullable[string]{},
			domain.Nullable[bool]{},
		)

		_, err := svc.PatchTask(ctx, 1, patch)

		if !errors.Is(err, core_errors.ErrConflict) {
			t.Errorf("expected ErrConflict, got: %v", err)
		}
	})
}

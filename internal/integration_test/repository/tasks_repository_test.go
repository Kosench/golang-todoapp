//go:build integration

package repository_integration_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)

func TestTasksRepository_CreateTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	task := domain.NewTaskUninitialized("Buy milk", ptr("Two bottles"), author.ID, fixedTime())

	created, err := repo.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error: %v", err)
	}

	if created.ID <= 0 {
		t.Errorf("ID = %d, want positive", created.ID)
	}
	if created.Version != 1 {
		t.Errorf("Version = %d, want 1", created.Version)
	}
	if created.Title != "Buy milk" {
		t.Errorf("Title = %q, want %q", created.Title, "Buy milk")
	}
	if created.Description == nil || *created.Description != "Two bottles" {
		t.Errorf("Description = %v, want %q", created.Description, "Two bottles")
	}
	if created.Completed {
		t.Errorf("Completed = true, want false")
	}
	if created.AuthorUserID != author.ID {
		t.Errorf("AuthorUserID = %d, want %d", created.AuthorUserID, author.ID)
	}
}

func TestTasksRepository_CreateTask_ForeignKeyViolation(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	task := domain.NewTaskUninitialized("Bad task", nil, missingID, fixedTime())

	_, err := repo.CreateTask(ctx, task)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_GetTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Buy milk", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	got, err := repo.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Title != created.Title {
		t.Errorf("Title = %q, want %q", got.Title, created.Title)
	}
}

func TestTasksRepository_GetTask_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	_, err := repo.GetTask(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_GetTasks(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	first, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("First task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}
	second, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Second task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}

	tasks, err := repo.GetTasks(ctx, nil, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}
	if tasks[0].ID != first.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, first.ID)
	}
	if tasks[1].ID != second.ID {
		t.Errorf("tasks[1].ID = %d, want %d", tasks[1].ID, second.ID)
	}
}

func TestTasksRepository_GetTasks_FilterByUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	firstAuthor := createUser(t, ctx, "First Author")
	secondAuthor := createUser(t, ctx, "Second Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	firstTask, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("First author task", nil, firstAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first author task: %v", err)
	}
	_, err = repo.CreateTask(ctx, domain.NewTaskUninitialized("Second author task", nil, secondAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second author task: %v", err)
	}

	tasks, err := repo.GetTasks(ctx, &firstAuthor.ID, ptr(10), ptr(0))
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != firstTask.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, firstTask.ID)
	}
	if tasks[0].AuthorUserID != firstAuthor.ID {
		t.Errorf("AuthorUserID = %d, want %d", tasks[0].AuthorUserID, firstAuthor.ID)
	}
}

func TestTasksRepository_PatchTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Old title", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	completedAt := fixedTime().Add(time.Hour)
	patched := created
	patched.Title = "New title"
	patched.Completed = true
	patched.CompletedAt = &completedAt

	got, err := repo.PatchTask(ctx, created.ID, patched)
	if err != nil {
		t.Fatalf("PatchTask() error: %v", err)
	}

	if got.Title != "New title" {
		t.Errorf("Title = %q, want %q", got.Title, "New title")
	}
	if !got.Completed {
		t.Errorf("Completed = false, want true")
	}
	if got.CompletedAt == nil {
		t.Fatalf("CompletedAt = nil, want non-nil")
	}
	if got.Version != created.Version+1 {
		t.Errorf("Version = %d, want %d", got.Version, created.Version+1)
	}
}

func TestTasksRepository_PatchTask_Conflict(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Old title", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	firstPatch := created
	firstPatch.Title = "New title"

	updated, err := repo.PatchTask(ctx, created.ID, firstPatch)
	if err != nil {
		t.Fatalf("first PatchTask() error: %v", err)
	}

	stale := created
	stale.Title = "Stale title"

	_, err = repo.PatchTask(ctx, updated.ID, stale)
	if !errors.Is(err, core_errors.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
}

func TestTasksRepository_DeleteTask(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Task Author")
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	created, err := repo.CreateTask(ctx, domain.NewTaskUninitialized("Task to delete", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if err := repo.DeleteTask(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	_, err = repo.GetTask(ctx, created.ID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("GetTask() after delete error = %v, want ErrNotFound", err)
	}
}

func TestTasksRepository_DeleteTask_NotFound(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	repo := tasks_postgres_repository.NewTasksRepository(pool)

	err := repo.DeleteTask(ctx, missingID)
	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

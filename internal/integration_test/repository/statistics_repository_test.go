//go:build integration

package repository_integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	statistics_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/statistics/repository/postgres"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
)

func TestStatisticsRepository_GetTasks_All(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Stats Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	first, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("First task", nil, author.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first task: %v", err)
	}
	second, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Second task", nil, author.ID, fixedTime().Add(time.Hour)))
	if err != nil {
		t.Fatalf("create second task: %v", err)
	}

	tasks, err := statsRepo.GetTasks(ctx, nil, nil, nil)
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

func TestStatisticsRepository_GetTasks_FilterByUser(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	firstAuthor := createUser(t, ctx, "First Author")
	secondAuthor := createUser(t, ctx, "Second Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	firstTask, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("First author task", nil, firstAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create first author task: %v", err)
	}
	_, err = tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Second author task", nil, secondAuthor.ID, fixedTime()))
	if err != nil {
		t.Fatalf("create second author task: %v", err)
	}

	tasks, err := statsRepo.GetTasks(ctx, &firstAuthor.ID, nil, nil)
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != firstTask.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, firstTask.ID)
	}
}

func TestStatisticsRepository_GetTasks_FilterByDateRange(t *testing.T) {
	cleanup(t)

	ctx := context.Background()
	author := createUser(t, ctx, "Stats Author")
	tasksRepo := tasks_postgres_repository.NewTasksRepository(pool)
	statsRepo := statistics_postgres_repository.NewStatisticsRepository(pool)

	baseTime := fixedTime()
	beforeRange := baseTime.Add(-time.Hour)
	inRange := baseTime
	afterRange := baseTime.Add(time.Hour)

	_, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("Before range", nil, author.ID, beforeRange))
	if err != nil {
		t.Fatalf("create before range task: %v", err)
	}
	expected, err := tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("In range", nil, author.ID, inRange))
	if err != nil {
		t.Fatalf("create in range task: %v", err)
	}
	_, err = tasksRepo.CreateTask(ctx, domain.NewTaskUninitialized("After range", nil, author.ID, afterRange))
	if err != nil {
		t.Fatalf("create after range task: %v", err)
	}

	from := baseTime.Add(-time.Minute)
	to := baseTime.Add(time.Minute)

	tasks, err := statsRepo.GetTasks(ctx, nil, &from, &to)
	if err != nil {
		t.Fatalf("GetTasks() error: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].ID != expected.ID {
		t.Errorf("tasks[0].ID = %d, want %d", tasks[0].ID, expected.ID)
	}
}

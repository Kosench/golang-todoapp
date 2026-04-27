package statistics_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	statistics_service "github.com/Kosench/golang-todoapp/internal/features/statistics/service"
)

func ptr[T any](v T) *T {
	return &v
}

type MockStatisticsRepository struct {
	GetTasksFunc func(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error)
}

func (m *MockStatisticsRepository) GetTasks(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
	return m.GetTasksFunc(ctx, userID, from, to)
}

func TestStatisticsService_GetStatistics(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("success: no filters", func(t *testing.T) {
		tasks := []domain.Task{
			domain.NewTask(1, 1, "Task 1", nil, true, now, ptr(now.Add(time.Hour)), 1),
			domain.NewTask(2, 1, "Task 2", nil, false, now, nil, 1),
		}

		repo := &MockStatisticsRepository{
			GetTasksFunc: func(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
				return tasks, nil
			},
		}

		svc := statistics_service.NewStatisticsService(repo)
		got, err := svc.GetStatistics(ctx, nil, nil, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.TasksCreated != 2 {
			t.Errorf("TasksCreated = %d, want 2", got.TasksCreated)
		}
		if got.TasksCompleted != 1 {
			t.Errorf("TasksCompleted = %d, want 1", got.TasksCompleted)
		}
	})

	t.Run("success: empty result", func(t *testing.T) {
		repo := &MockStatisticsRepository{
			GetTasksFunc: func(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
				return []domain.Task{}, nil
			},
		}

		svc := statistics_service.NewStatisticsService(repo)
		got, err := svc.GetStatistics(ctx, nil, nil, nil)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.TasksCreated != 0 {
			t.Errorf("TasksCreated = %d, want 0", got.TasksCreated)
		}
		if got.TasksCompletedRate != nil {
			t.Error("TasksCompletedRate should be nil for empty tasks")
		}
	})

	t.Run("error: 'to' before 'from'", func(t *testing.T) {
		// Бизнес-правило: to должно быть после from.
		from := now
		to := now.Add(-time.Hour) // to < from

		repo := &MockStatisticsRepository{
			// GetTasksFunc не задаём — не должен вызываться
		}

		svc := statistics_service.NewStatisticsService(repo)
		_, err := svc.GetStatistics(ctx, nil, &from, &to)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})

	t.Run("error: 'to' equals 'from'", func(t *testing.T) {
		// to.Equal(from) тоже невалидно
		from := now
		to := from // равны

		repo := &MockStatisticsRepository{}

		svc := statistics_service.NewStatisticsService(repo)
		_, err := svc.GetStatistics(ctx, nil, &from, &to)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, core_errors.ErrInvalidArgument) {
			t.Errorf("expected ErrInvalidArgument, got: %v", err)
		}
	})

	t.Run("error: repository fails", func(t *testing.T) {
		repo := &MockStatisticsRepository{
			GetTasksFunc: func(ctx context.Context, userID *int, from *time.Time, to *time.Time) ([]domain.Task, error) {
				return nil, errors.New("connection lost")
			},
		}

		svc := statistics_service.NewStatisticsService(repo)
		_, err := svc.GetStatistics(ctx, nil, nil, nil)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

package statistics_service

import (
	"testing"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCalcStatistics(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name                   string
		tasks                  []domain.Task
		wantCreated            int
		wantCompleted          int
		wantRateNil            bool          // TasksCompletedRate == nil?
		wantRate               float64       // ожидаемое значение, если не nil
		wantAvgCompletionNil   bool          // TasksAverageCompletionTime == nil?
		wantAvgCompletionApprx time.Duration // приблизительное значение
	}{
		{
			name:                 "empty tasks",
			tasks:                []domain.Task{},
			wantCreated:          0,
			wantCompleted:        0,
			wantRateNil:          true,
			wantAvgCompletionNil: true,
		},
		{
			name: "all tasks incomplete",
			tasks: []domain.Task{
				domain.NewTask(1, 1, "Task 1", nil, false, now, nil, 1),
				domain.NewTask(2, 1, "Task 2", nil, false, now, nil, 1),
			},
			wantCreated:          2,
			wantCompleted:        0,
			wantRateNil:          false,
			wantRate:             0, // 0 из 2 = 0%
			wantAvgCompletionNil: true,
		},
		{
			name: "all tasks completed",
			tasks: []domain.Task{
				// Задача 1: создана now, завершена через 2 часа
				domain.NewTask(1, 1, "Task 1", nil, true, now, ptr(now.Add(2*time.Hour)), 1),
				// Задача 2: создана now, завершена через 4 часа
				domain.NewTask(2, 1, "Task 2", nil, true, now, ptr(now.Add(4*time.Hour)), 1),
			},
			wantCreated:            2,
			wantCompleted:          2,
			wantRateNil:            false,
			wantRate:               100, // 2 из 2 = 100%
			wantAvgCompletionNil:   false,
			wantAvgCompletionApprx: 3 * time.Hour, // (2h + 4h) / 2 = 3h
		},
		{
			name: "mixed: some completed, some not",
			tasks: []domain.Task{
				domain.NewTask(1, 1, "Done", nil, true, now, ptr(now.Add(time.Hour)), 1),
				domain.NewTask(2, 1, "Not done", nil, false, now, nil, 1),
				domain.NewTask(3, 1, "Not done 2", nil, false, now, nil, 1),
			},
			wantCreated:            3,
			wantCompleted:          1,
			wantRateNil:            false,
			wantRate:               100.0 / 3.0, // ~33.33%
			wantAvgCompletionNil:   false,
			wantAvgCompletionApprx: time.Hour, // только 1 задача за 1h
		},
		{
			name: "single completed task",
			tasks: []domain.Task{
				domain.NewTask(1, 1, "Solo", nil, true, now, ptr(now.Add(30*time.Minute)), 1),
			},
			wantCreated:            1,
			wantCompleted:          1,
			wantRateNil:            false,
			wantRate:               100,
			wantAvgCompletionNil:   false,
			wantAvgCompletionApprx: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcStatistics(tt.tasks)

			if got.TasksCreated != tt.wantCreated {
				t.Errorf("TasksCreated = %d, want %d", got.TasksCreated, tt.wantCreated)
			}

			if got.TasksCompleted != tt.wantCompleted {
				t.Errorf("TasksCompleted = %d, want %d", got.TasksCompleted, tt.wantCompleted)
			}

			// Проверяем TasksCompletedRate
			if tt.wantRateNil {
				if got.TasksCompletedRate != nil {
					t.Errorf("TasksCompletedRate = %v, want nil", *got.TasksCompletedRate)
				}
			} else {
				if got.TasksCompletedRate == nil {
					t.Fatal("TasksCompletedRate is nil, want non-nil")
				}
				// Сравниваем с допуском для float64
				if diff := *got.TasksCompletedRate - tt.wantRate; diff > 0.01 || diff < -0.01 {
					t.Errorf("TasksCompletedRate = %f, want %f", *got.TasksCompletedRate, tt.wantRate)
				}
			}

			// Проверяем TasksAverageCompletionTime
			if tt.wantAvgCompletionNil {
				if got.TasksAverageCompletionTime != nil {
					t.Errorf("TasksAverageCompletionTime = %v, want nil", *got.TasksAverageCompletionTime)
				}
			} else {
				if got.TasksAverageCompletionTime == nil {
					t.Fatal("TasksAverageCompletionTime is nil, want non-nil")
				}
				if *got.TasksAverageCompletionTime != tt.wantAvgCompletionApprx {
					t.Errorf("TasksAverageCompletionTime = %v, want %v",
						*got.TasksAverageCompletionTime, tt.wantAvgCompletionApprx)
				}
			}
		})
	}
}

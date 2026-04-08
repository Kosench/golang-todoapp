package statistics_service

import (
	"context"
	"fmt"
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
)

func (s *StatisticsService) GetStatistics(ctx context.Context, userID *int, from *time.Time, to *time.Time) (domain.Statistics, error) {
	if from != nil && to != nil {
		if to.Before(*from) || to.Equal(*from) {
			return domain.Statistics{}, fmt.Errorf("`to` must be after `from`: %w", core_errors.ErrInvalidArgument)
		}
	}

	tasks, err := s.statisticsRepository.GetTasks(ctx, userID, from, to)
	if err != nil {
		return domain.Statistics{}, fmt.Errorf("failed to get tasks from repository: %w", err)
	}

	statistic := calcStatistics(tasks)

	return statistic, nil
}

func calcStatistics(tasks []domain.Task) domain.Statistics {
	if len(tasks) == 0 {
		return domain.Statistics{
			TasksCreated:               0,
			TasksCompleted:             0,
			TasksCompletedRate:         nil,
			TasksAverageCompletionTime: nil,
		}
	}

	taskCreated := len(tasks)
	taskCompleted := 0
	var totalCompletionDuration time.Duration

	for _, task := range tasks {
		if task.Completed {
			taskCompleted++
		}

		completionDuration := task.CompletionDuration()
		if completionDuration != nil {
			totalCompletionDuration += *completionDuration
		}
	}

	tasksCompletedRate := float64(taskCompleted) / float64(taskCreated) * 100

	var tasksAverageCompletionTime *time.Duration
	if taskCompleted > 0 && totalCompletionDuration != 0 {
		avg := totalCompletionDuration / time.Duration(taskCompleted)
		tasksAverageCompletionTime = &avg
	}

	return domain.Statistics{
		TasksCreated:               taskCreated,
		TasksCompleted:             taskCompleted,
		TasksCompletedRate:         &tasksCompletedRate,
		TasksAverageCompletionTime: tasksAverageCompletionTime,
	}
}

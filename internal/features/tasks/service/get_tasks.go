package tasks_service

import (
	"context"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func (s *TaskService) GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error) {
	tasks, err := s.taskRepository.GetTasks(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get tasks from repository: %w", err)
	}

	return tasks, nil
}

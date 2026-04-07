package tasks_service

import (
	"context"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func (s *TaskService) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {
	if err := task.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf("validate task domain: %w", err)
	}

	createdTask, err := s.taskRepository.CreateTask(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to create task: %w", err)
	}

	return createdTask, nil
}

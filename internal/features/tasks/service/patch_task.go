package tasks_service

import (
	"context"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

func (s *TaskService) PatchTask(ctx context.Context, id int, patch domain.TaskPatch) (domain.Task, error) {
	task, err := s.taskRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task: %w", err)
	}

	if err = task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf("apply patch task: %w", err)
	}

	patchedTask, err := s.taskRepository.PatchTask(ctx, id, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("patch task: %w", err)
	}

	return patchedTask, nil
}

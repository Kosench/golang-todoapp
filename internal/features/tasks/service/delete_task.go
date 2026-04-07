package tasks_service

import (
	"context"
	"fmt"
)

func (s *TaskService) DeleteTask(ctx context.Context, id int) error {
	if err := s.taskRepository.DeleteTask(ctx, id); err != nil {
		return fmt.Errorf(
			"delete task from repository: %w",
			err,
		)
	}

	return nil
}

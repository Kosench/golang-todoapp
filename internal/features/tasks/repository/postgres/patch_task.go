package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool"
)

func (r *TaskRepository) PatchTask(ctx context.Context, id int, task domain.Task) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE todoapp.tasks
	SET
	    title=$1,
	    description=$2,
	    completed=$3,
	    completed_at=$4,
		version=version+1
	WHERE id=$5 AND version=$6
	RETURNING 
	    id, 
	    version, 
	    title, 
	    description, 
	    completed, 
	    created_at, 
	    completed_at, 
	    author_user_id;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		task.Title,
		task.Description,
		task.Completed,
		task.CompletedAt,
		id,
		task.Version,
	)

	var taskModel TaskModel

	err := row.Scan(
		&taskModel.ID,
		&taskModel.Version,
		&taskModel.Title,
		&taskModel.Description,
		&taskModel.Completed,
		&taskModel.CreatedAt,
		&taskModel.CompletedAt,
		&taskModel.AuthorUserID,
	)

	if err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			exists, existsErr := r.taskExists(ctx, id)
			if existsErr != nil {
				return domain.Task{}, fmt.Errorf("check task existence: %w", existsErr)
			}
			if !exists {
				return domain.Task{}, fmt.Errorf("task with id='%d': %w", id, core_errors.ErrNotFound)
			}

			return domain.Task{}, fmt.Errorf("task with id='%d' concurrently accessed: %w", id, core_errors.ErrConflict)
		}
		if errors.Is(err, core_postgres_pool.ErrViolatesCheckConstraint) {
			return domain.Task{}, fmt.Errorf("task violates database constraint: %w", core_errors.ErrInvalidArgument)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := taskDomainFromModel(taskModel)

	return taskDomain, nil
}

func (r *TaskRepository) taskExists(ctx context.Context, id int) (bool, error) {
	query := `
	SELECT EXISTS(
		SELECT 1
		FROM todoapp.tasks
		WHERE id=$1
	);
	`

	row := r.pool.QueryRow(ctx, query, id)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("scan task exists: %w", err)
	}

	return exists, nil
}

package tasks_service

import (
	"context"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

type TaskService struct {
	taskRepository TasksRepository
}

type TasksRepository interface {
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	GetTasks(ctx context.Context, userID *int, limit *int, offset *int) ([]domain.Task, error)
}

func NewTaskService(repo TasksRepository) *TaskService {
	return &TaskService{taskRepository: repo}
}

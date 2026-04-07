package tasks_postgres_repository

import (
	"time"

	"github.com/Kosench/golang-todoapp/internal/core/domain"
)

type TaskModel struct {
	ID      int
	Version int

	Title        string
	Description  *string
	Completed    bool
	CreatedAt    time.Time
	CompletedAt  *time.Time
	AuthorUserID int
}

func taskDomainsFromModel(tasks []TaskModel) []domain.Task {
	taskDomains := make([]domain.Task, len(tasks))
	for i, task := range tasks {
		taskDomains[i] = domain.NewTask(
			task.ID,
			task.Version,
			task.Title,
			task.Description,
			task.Completed,
			task.CreatedAt,
			task.CompletedAt,
			task.AuthorUserID,
		)
	}

	return taskDomains
}

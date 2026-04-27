package core_postgres_model

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

func TaskDomainFromModel(taskModel TaskModel) domain.Task {
	return domain.NewTask(
		taskModel.ID,
		taskModel.Version,
		taskModel.Title,
		taskModel.Description,
		taskModel.Completed,
		taskModel.CreatedAt,
		taskModel.CompletedAt,
		taskModel.AuthorUserID,
	)
}

func TaskDomainsFromModel(tasks []TaskModel) []domain.Task {
	taskDomains := make([]domain.Task, len(tasks))
	for i, task := range tasks {
		taskDomains[i] = TaskDomainFromModel(task)
	}

	return taskDomains
}

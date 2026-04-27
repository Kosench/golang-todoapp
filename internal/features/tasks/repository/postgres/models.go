package tasks_postgres_repository

import core_postgres_model "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/model"

type TaskModel = core_postgres_model.TaskModel

var (
	taskDomainFromModel  = core_postgres_model.TaskDomainFromModel
	taskDomainsFromModel = core_postgres_model.TaskDomainsFromModel
)

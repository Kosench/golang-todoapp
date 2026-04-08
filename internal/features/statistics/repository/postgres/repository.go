package statistics_postgres_repository

import core_postgres_poll "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool"

type StatisticsRepository struct {
	core_postgres_poll.Pool
}

func NewStatisticsRepository(pool core_postgres_poll.Pool) *StatisticsRepository {
	return &StatisticsRepository{Pool: pool}
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	"github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Kosench/golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Kosench/golang-todoapp/internal/core/transport/http/server"
	statistics_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/statistics/repository/postgres"
	statistics_service "github.com/Kosench/golang-todoapp/internal/features/statistics/service"
	statistics_transport_http "github.com/Kosench/golang-todoapp/internal/features/statistics/transport"
	tasks_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/Kosench/golang-todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/Kosench/golang-todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
	users_service "github.com/Kosench/golang-todoapp/internal/features/users/service"
	users_transport_http "github.com/Kosench/golang-todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"

	_ "github.com/Kosench/golang-todoapp/docs"
)

var (
	timeZone = time.UTC
)

// @title Golang Todo API
// @version 1.0
// @description Todo Application REST-API scheme
// @host 127.0.0.1:5050
// @BasePath /api/v1
func main() {
	time.Local = timeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Printf("failed to init application logger: %w", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("initializing postgres connection pool")
	//pool, err := core_postgres_poll.NewConnectionPool(ctx, core_pgx_pool.NewMustConfig())
	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewMustConfig(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool %w: ", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := users_service.NewUserService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := tasks_postgres_repository.NewTasksRepository(pool)
	tasksService := tasks_service.NewTaskService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing feature", zap.String("feature", "statistics"))
	statsRepository := statistics_postgres_repository.NewStatisticsRepository(pool)
	statsService := statistics_service.NewStatisticsService(statsRepository)
	statsTransportHTTP := statistics_transport_http.NewStatisticsHTTPHandler(statsService)

	logger.Debug("initializing HTTP server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouterV1 := core_http_server.NewVersionAPI(core_http_server.APIVersion1)
	apiVersionRouterV1.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.RegisterRoutes(statsTransportHTTP.Routes()...)

	//apiVersionRouterV2 := core_http_server.NewVersionAPI(
	//	core_http_server.APIVersion2,
	//	core_http_middleware.Dummy("api v2 middleware"),
	//)
	//apiVersionRouterV2.RegisterRoutes(usersTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(
		apiVersionRouterV1,
		//apiVersionRouterV2,
	)

	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Server HTTP run error", zap.Error(err))
	}
}

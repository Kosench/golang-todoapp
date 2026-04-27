package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_pgx_pool "github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool/pgx"
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
	web_fs_repository "github.com/Kosench/golang-todoapp/internal/features/web/repository/file_system"
	web_service "github.com/Kosench/golang-todoapp/internal/features/web/service"
	web_transport_http "github.com/Kosench/golang-todoapp/internal/features/web/transport/http"
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
		fmt.Fprintf(os.Stderr, "failed to init application logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("initializing postgres connection pool")
	pool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewMustConfig(),
	)
	if err != nil {
		logger.Error("failed to init postgres connection pool", zap.Error(err))
		os.Exit(1)
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

	logger.Debug("initializing feature", zap.String("feature", "web"))
	webRepository := web_fs_repository.NewWebRepository()
	webService := web_service.NewWebService(webRepository, os.Getenv("PROJECT_ROOT"))
	webTransportHTTP := web_transport_http.NewWebHTTPHandler(webService)

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

	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterV1.AddRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(statsTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(
		apiVersionRouterV1,
	)

	httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Server HTTP run error", zap.Error(err))
	}
}

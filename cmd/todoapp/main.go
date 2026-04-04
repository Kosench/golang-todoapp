package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	"github.com/Kosench/golang-todoapp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/Kosench/golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Kosench/golang-todoapp/internal/core/transport/http/server"
	users_postgres_repository "github.com/Kosench/golang-todoapp/internal/features/users/repository/postgres"
	users_service "github.com/Kosench/golang-todoapp/internal/features/users/service"
	users_transport_http "github.com/Kosench/golang-todoapp/internal/features/users/transport/http"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger: %w", err)
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

	logger.Debug("initializing HTTP server")
	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	apiVersionRouterV1 := core_http_server.NewVersionAPI(core_http_server.APIVersion1)
	apiVersionRouterV1.RegisterRoutes(usersTransportHTTP.Routes()...)

	//apiVersionRouterV2 := core_http_server.NewVersionAPI(
	//	core_http_server.APIVersion2,
	//	core_http_middleware.Dummy("api v2 middleware"),
	//)
	//apiVersionRouterV2.RegisterRoutes(usersTransportHTTP.Routes()...)

	httpServer.RegisterAPIRouters(
		apiVersionRouterV1,
		//apiVersionRouterV2,
	)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Server HTTP run error", zap.Error(err))
	}
}

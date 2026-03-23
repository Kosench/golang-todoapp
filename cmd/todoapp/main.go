package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Kosench/golang-todoapp/internal/core/logger"
	core_http_middleware "github.com/Kosench/golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/Kosench/golang-todoapp/internal/core/transport/http/server"
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

	logger.Debug("Starting TO DO APP")

	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(nil)
	usersRoutes := usersTransportHTTP.Routes()

	apiVersionRouter := core_http_server.NewVersionAPI(core_http_server.APIVersion1)
	apiVersionRouter.RegisterRoutes(usersRoutes...)

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Panic(),
		core_http_middleware.Trace(),
	)
	httpServer.RegisterAPIRouters(apiVersionRouter)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Server HTTP run error", zap.Error(err))
	}
}

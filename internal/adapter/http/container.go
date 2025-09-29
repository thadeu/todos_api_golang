package http

import (
	"log/slog"

	database "todos/internal/adapter/database/sqlite"
	repository "todos/internal/adapter/database/sqlite/repository"

	"todos/internal/adapter/http/handler"
	"todos/internal/core/port"
	"todos/internal/core/service"
	"todos/internal/core/telemetry"
	"todos/pkg/config"
)

type Container struct {
	UserRepo port.UserRepository
	TodoRepo port.TodoRepository

	UserUseCase port.UserService
	TodoUseCase port.TodoService
	AuthUseCase port.AuthService

	UserHandler *handler.UserHandler
	TodoHandler *handler.TodoHandler
	AuthHandler *handler.AuthHandler
}

func NewContainer(db *database.DB, logger *config.LokiLogger) *Container {
	// Create telemetry probe - centralized point for all telemetry
	probe := telemetry.NewOTELProbe(slog.Default())
	// For testing or when telemetry is disabled, use:
	// probe := telemetry.NewNoOpProbe()

	// Inject probe into repositories
	userRepo := repository.NewUserRepository(db, probe)
	todoRepo := repository.NewTodoRepository(db, probe)

	// Services get probe for business-level telemetry
	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	todoSvc := service.NewTodoService(todoRepo, probe)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	todoHandler := handler.NewTodoHandler(todoSvc, logger)

	return &Container{
		AuthHandler: authHandler,

		TodoRepo:    todoRepo,
		TodoHandler: todoHandler,

		UserRepo:    userRepo,
		UserHandler: userHandler,
	}
}

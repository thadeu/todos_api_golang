package http

import (
	database "todoapp/internal/adapter/database/sqlite"
	repository "todoapp/internal/adapter/database/sqlite/repository"

	"todoapp/internal/adapter/http/handler"
	"todoapp/internal/core/port"
	"todoapp/internal/core/service"
	"todoapp/pkg/config"
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
	userRepo := repository.NewUserRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	authSvc := service.NewAuthService(userRepo)
	userSvc := service.NewUserService(userRepo)
	todoSvc := service.NewTodoService(todoRepo)

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

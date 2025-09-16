package infrastructure

import (
	"database/sql"

	"todoapp/internal/delivery/http/handler"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/impl"
	"todoapp/internal/usecase/interfaces"
	"todoapp/pkg/config"
)

// Container holds all dependencies
type Container struct {
	// Repositories
	UserRepo repositories.UserRepository
	TodoRepo repositories.TodoRepository

	// Use Cases
	UserUseCase interfaces.UserUseCase
	TodoUseCase interfaces.TodoUseCase
	AuthUseCase interfaces.AuthUseCase

	// Handlers
	UserHandler *handler.UserHandler
	TodoHandler *handler.TodoHandler
	AuthHandler *handler.AuthHandler
}

// NewContainer creates a new dependency container
func NewContainer(db *sql.DB, logger *config.LokiLogger) *Container {
	// Initialize repositories
	userRepo := persistence.NewUserRepository(db)
	todoRepo := persistence.NewTodoRepository(db)

	// Initialize use cases
	userUseCase := impl.NewUserUseCase(userRepo)
	todoUseCase := impl.NewTodoUseCase(todoRepo)
	authUseCase := impl.NewAuthUseCase(userRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUseCase)
	todoHandler := handler.NewTodoHandler(todoUseCase, logger)
	authHandler := handler.NewAuthHandler(authUseCase)

	return &Container{
		UserRepo:    userRepo,
		TodoRepo:    todoRepo,
		UserUseCase: userUseCase,
		TodoUseCase: todoUseCase,
		AuthUseCase: authUseCase,
		UserHandler: userHandler,
		TodoHandler: todoHandler,
		AuthHandler: authHandler,
	}
}

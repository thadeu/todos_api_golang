package interfaces

import (
	"context"

	"todoapp/internal/domain/entities"
	"todoapp/pkg/db/cursor"

	"github.com/gin-gonic/gin"
)

// TodoUseCase defines the interface for todo business operations
type TodoUseCase interface {
	// GetTodosWithPagination retrieves todos with pagination
	GetTodosWithPagination(ctx context.Context, userId int, limit int, cursor string) (*cursor.CursorResponse, error)

	// GetAllTodos retrieves all todos for a user
	GetAllTodos(userId int) ([]interface{}, error)

	// CreateTodo creates a new todo
	CreateTodo(ctx context.Context, c *gin.Context, userId int) (entities.Todo, error)

	// UpdateTodoByUUID updates a todo by UUID
	UpdateTodoByUUID(ctx context.Context, c *gin.Context, userId int) (entities.Todo, error)

	// DeleteTodo deletes a todo
	DeleteTodo(c *gin.Context, userId int)

	// DeleteByUUID marks a todo as deleted by UUID
	DeleteByUUID(ctx context.Context, c *gin.Context, userId int) (map[string]any, error)
}

package repositories

import (
	"context"

	"todoapp/internal/domain/entities"
)

// TodoRepository defines the interface for todo data operations
type TodoRepository interface {
	// Create creates a new todo
	Create(ctx context.Context, todo entities.Todo) (entities.Todo, error)

	// GetAll retrieves all todos for a user
	GetAll(userId int) ([]entities.Todo, error)

	// GetAllWithCursor retrieves todos with pagination
	GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]entities.Todo, bool, error)

	// GetByUUID finds a todo by UUID
	GetByUUID(ctx context.Context, id string, userId int) (entities.Todo, error)

	// GetById finds a todo by ID
	GetById(ctx context.Context, id string) (entities.Todo, error)

	// UpdateByUUID updates a todo by UUID
	UpdateByUUID(ctx context.Context, id string, userId int, params interface{}) (entities.Todo, error)

	// DeleteById deletes a todo by ID
	DeleteById(id string) error

	// DeleteByUUID marks a todo as deleted by UUID
	DeleteByUUID(ctx context.Context, uuid string) error
}

package services

import (
	"context"

	"todoapp/internal/domain/entities"
)

// TodoService defines domain-level todo operations
type TodoService interface {
	// ValidateTodoOwnership validates if a todo belongs to a user
	ValidateTodoOwnership(ctx context.Context, todo *entities.Todo, userID int) error

	// CalculateTodoProgress calculates the progress of a todo based on its status
	CalculateTodoProgress(ctx context.Context, todo *entities.Todo) float64
}

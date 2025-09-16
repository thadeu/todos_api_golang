package interfaces

import (
	"context"
	"net/http"

	"todoapp/internal/domain/entities"
)

// UserUseCase defines the interface for user business operations
type UserUseCase interface {
	// GetAllUsers retrieves all users
	GetAllUsers() ([]interface{}, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, r *http.Request) (entities.User, error)

	// DeleteUser deletes a user
	DeleteUser(w http.ResponseWriter, r *http.Request)

	// DeleteByUUID marks a user as deleted by UUID
	DeleteByUUID(ctx context.Context, r *http.Request) (map[string]any, error)
}

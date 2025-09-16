package repositories

import (
	"context"

	"todoapp/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user entities.User) (entities.User, error)

	// GetAllUsers retrieves all users
	GetAllUsers() ([]entities.User, error)

	// GetUserByUUID finds a user by UUID
	GetUserByUUID(uuid string) (entities.User, error)

	// GetUserById finds a user by ID
	GetUserById(id string) (entities.User, error)

	// GetUserByEmail finds a user by email
	GetUserByEmail(ctx context.Context, email string) (entities.User, error)

	// DeleteUser deletes a user by ID
	DeleteUser(id string) error

	// DeleteByUUID marks a user as deleted by UUID
	DeleteByUUID(ctx context.Context, uuid string) error
}

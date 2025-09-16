package interfaces

import (
	"context"

	"todoapp/internal/domain/entities"
)

// AuthRequest represents the request structure for authentication
type AuthRequest struct {
	Email    string `json:"email,omitempty" validate:"required,email,max=255"`
	Password string `json:"password,omitempty" validate:"required,min=6,max=100"`
}

// AuthUseCase defines the interface for authentication business operations
type AuthUseCase interface {
	// Registration creates a new user account
	Registration(ctx context.Context, email string, password string) (*entities.User, error)

	// Authenticate verifies user credentials
	Authenticate(ctx context.Context, email string, password string) (*entities.User, error)

	// GenerateRefreshToken generates a JWT token for the user
	GenerateRefreshToken(user *entities.User) (string, error)
}

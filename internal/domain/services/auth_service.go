package services

import (
	"context"

	"todoapp/internal/domain/entities"
)

// AuthService defines domain-level authentication operations
type AuthService interface {
	// ValidateCredentials validates user credentials
	ValidateCredentials(ctx context.Context, email, password string) (*entities.User, error)

	// CreateUserAccount creates a new user account with proper validation
	CreateUserAccount(ctx context.Context, email, password string) (*entities.User, error)
}

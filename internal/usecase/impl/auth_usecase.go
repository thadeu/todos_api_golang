package impl

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	. "todoapp/pkg/auth"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/usecase/interfaces"
)

// authUseCase implements the AuthUseCase interface
type authUseCase struct {
	userRepo repositories.UserRepository
}

// NewAuthUseCase creates a new auth use case
func NewAuthUseCase(userRepo repositories.UserRepository) interfaces.AuthUseCase {
	return &authUseCase{
		userRepo: userRepo,
	}
}

func (a *authUseCase) Registration(ctx context.Context, email string, password string) (*entities.User, error) {
	// Check if user already exists
	oldUser, err := a.userRepo.GetUserByEmail(ctx, email)
	if err == nil && oldUser.Email != "" {
		slog.Error("User already exists", "error", err)
		return nil, fmt.Errorf("user already exists")
	}

	// Encrypt password
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("Error creating encrypted password", "error", err)
		return nil, fmt.Errorf("error creating encrypted password")
	}

	// Create user entity
	user := entities.User{
		UUID:              uuid.New(),
		Email:             email,
		EncryptedPassword: string(encrypted),
	}

	// Save user
	savedUser, err := a.userRepo.CreateUser(ctx, user)
	if err != nil {
		slog.Error("Unexpected error creating user", "error", err)
		return nil, fmt.Errorf("unexpected error creating user")
	}

	return &savedUser, nil
}

func (a *authUseCase) Authenticate(ctx context.Context, email string, password string) (*entities.User, error) {
	// Get user by email
	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	// Verify password
	userPassword := []byte(user.EncryptedPassword)
	formPassword := []byte(password)

	if err := bcrypt.CompareHashAndPassword(userPassword, formPassword); err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	return &user, nil
}

func (a *authUseCase) GenerateRefreshToken(user *entities.User) (string, error) {
	token, err := CreateJwtTokenForUser(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

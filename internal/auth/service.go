package auth

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	u "todoapp/internal/user"
	. "todoapp/pkg/auth"
)

type AuthService struct {
	repo *u.UserRepository
}

type AuthRequest struct {
	Email    string `json:"email,omitempty" validate:"required,email,max=255"`
	Password string `json:"password,omitempty" validate:"required,min=6,max=100"`
}

func NewAuthService(repo *u.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (t *AuthService) Registration(ctx context.Context, email string, password string) (u.User, error) {
	oldUser, err := t.repo.GetUserByEmail(ctx, email)

	if oldUser.Email != "" {
		slog.Error("User already exists", "error", err)
		return u.User{}, fmt.Errorf("%v", "User already exists")
	}

	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		slog.Error("Error to create encrypted password", "error", err)
		return u.User{}, fmt.Errorf("%v", "Error to create encrypted password")
	}

	user, err := t.repo.CreateUser(ctx, u.User{
		UUID:              uuid.New(),
		Email:             email,
		EncryptedPassword: string(encrypted),
	})

	if err != nil {
		slog.Error("Unexpected error to create a user", "error", err)
		return u.User{}, fmt.Errorf("%v", "Unexpected error to create a user")
	}

	return user, nil
}

func (t *AuthService) Authenticate(ctx context.Context, email string, password string) (u.User, error) {
	user, err := t.repo.GetUserByEmail(ctx, email)

	userPassword := []byte(user.EncryptedPassword)
	formPassword := []byte(password)

	if err != nil {
		return u.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword(userPassword, formPassword); err != nil {
		return u.User{}, fmt.Errorf("%v", "Invalid password")
	}

	return user, nil
}

func (t *AuthService) GenerateRefreshToken(user *u.User) (string, error) {
	token, err := CreateJwtTokenForUser(user.ID)

	if err != nil {
		return "", err
	}

	return token, nil
}

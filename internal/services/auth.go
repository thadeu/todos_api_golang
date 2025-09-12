package services

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	m "todoapp/internal/models"
	ru "todoapp/internal/repositories"
	s "todoapp/internal/shared"
)

type AuthService struct {
	repo *ru.UserRepository
}

type AuthRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func NewAuthService(repo *ru.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

func (t *AuthService) Registration(email string, password string) (m.User, error) {
	oldUser, err := t.repo.GetUserByEmail(email)

	if oldUser.Email != "" {
		slog.Error("User already exists", "error", err)
		return m.User{}, fmt.Errorf("%v", "User already exists")
	}

	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		slog.Error("Error to create encrypted password", "error", err)
		return m.User{}, fmt.Errorf("%v", "Error to create encrypted password")
	}

	user, err := t.repo.CreateUser(m.User{
		UUID:              uuid.New(),
		Email:             email,
		EncryptedPassword: string(encrypted),
	})

	if err != nil {
		slog.Error("Unexpected error to create a user", "error", err)
		return m.User{}, fmt.Errorf("%v", "Unexpected error to create a user")
	}

	return user, nil
}

func (t *AuthService) Authenticate(email string, password string) (m.User, error) {
	user, err := t.repo.GetUserByEmail(email)

	userPassword := []byte(user.EncryptedPassword)
	formPassword := []byte(password)

	if err != nil {
		return m.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword(userPassword, formPassword); err != nil {
		return m.User{}, fmt.Errorf("%v", "Invalid password")
	}

	return user, nil
}

func (t *AuthService) GenerateRefreshToken(user *m.User) (string, error) {
	token, err := s.CreateJwtTokenForUser(user.ID)

	if err != nil {
		return "", err
	}

	return token, nil
}

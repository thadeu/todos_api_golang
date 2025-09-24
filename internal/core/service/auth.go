package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"todoapp/internal/core/domain"
	"todoapp/internal/core/model/request"
	"todoapp/internal/core/port"
	"todoapp/internal/core/util"
)

type AuthService struct {
	repo port.UserRepository
}

func NewAuthService(repo port.UserRepository) *AuthService {
	return &AuthService{repo}
}

func (us *AuthService) Registration(ctx context.Context, req *request.SignUpRequest) (*domain.User, error) {
	oldUser, err := us.repo.GetByEmail(ctx, req.Email)

	if err == nil && oldUser.Email != "" {
		return nil, fmt.Errorf("user already exists")
	}

	encrypted, err := util.GenerateEncrypt(req.Password)

	if err != nil {
		return nil, fmt.Errorf("error creating encrypted password")
	}

	user := domain.User{
		UUID:              uuid.New(),
		Name:              "",
		Email:             req.Email,
		EncryptedPassword: string(encrypted),
		Role:              domain.Profile,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	savedUser, err := us.repo.Create(ctx, user)

	if err != nil {
		return nil, err
	}

	return &savedUser, nil
}

func (us *AuthService) Authenticate(ctx context.Context, req *request.LoginRequest) (*domain.User, error) {
	user, err := us.repo.GetByEmail(ctx, req.Email)

	if err != nil {
		slog.Error("Auth#Authenticate", "get_by_email", err)
		return nil, fmt.Errorf("authentication failed")
	}

	if err := util.ComparePassword(req.Password, user.EncryptedPassword); err != nil {
		slog.Error("Auth#Authenticate", "compare_password", err)
		return nil, fmt.Errorf("compare password failed")
	}

	slog.Info("Auth#Authenticate", "user", user)

	return &user, nil
}

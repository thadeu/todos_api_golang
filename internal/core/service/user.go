package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"todos/internal/core/domain"
	"todos/internal/core/port"
)

type UserService struct {
	repo port.UserRepository
}

func NewUserService(repo port.UserRepository) *UserService {
	return &UserService{repo}
}

func (ts *UserService) Create(ctx context.Context, user domain.User) (domain.User, error) {
	now := time.Now()

	if user.Role == "" {
		user.Role = domain.Profile
	}

	newData := domain.User{
		UUID:              uuid.New(),
		Name:              user.Name,
		Email:             user.Email,
		EncryptedPassword: user.EncryptedPassword,
		Role:              user.Role,
		CreatedAt:         now,
		UpdatedAt:         now,
		DeletedAt:         nil,
	}

	user, err := ts.repo.Create(ctx, newData)

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *UserService) GetUserByUUID(ctx context.Context, uuid string) (domain.User, error) {
	user, err := u.repo.GetByUUID(ctx, uuid)

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *UserService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.repo.GetByEmail(ctx, email)

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *UserService) DeleteByUUID(ctx context.Context, uid string) error {
	err := u.repo.DeleteByUUID(ctx, uid)

	if err != nil {
		return err
	}

	return nil
}

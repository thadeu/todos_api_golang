package port

import (
	"context"
	"todos/internal/core/domain"
)

type UserRepository interface {
	GetByUUID(ctx context.Context, uuid string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	DeleteByUUID(ctx context.Context, uuid string) error
}

type UserService interface {
	GetUserByUUID(ctx context.Context, uuid string) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, user domain.User) (domain.User, error)
	DeleteByUUID(ctx context.Context, uuid string) error
}

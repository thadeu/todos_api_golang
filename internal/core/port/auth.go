package port

import (
	"context"

	"todos/internal/core/domain"
	"todos/internal/core/model/request"
)

type AuthService interface {
	Registration(ctx context.Context, req *request.SignUpRequest) (*domain.User, error)
	Authenticate(ctx context.Context, req *request.LoginRequest) (*domain.User, error)
}

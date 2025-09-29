package port

import (
	"context"

	"todos/internal/core/domain"
	"todos/internal/core/model/response"
)

type TodoRepository interface {
	GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]domain.Todo, bool, error)
	GetByUUID(ctx context.Context, id string) (domain.Todo, error)
	Create(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	DeleteByUUID(ctx context.Context, uuid string) error
}

type TodoService interface {
	GetTodosWithPagination(ctx context.Context, userId int, limit int, cursor string) (*response.CursorResponse, error)
	Create(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error)
	DeleteByUUID(ctx context.Context, uid string) error
}

package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"todoapp/internal/core/domain"
	"todoapp/internal/core/model/response"
	"todoapp/internal/core/port"
	"todoapp/internal/core/util"
)

type TodoService struct {
	repo port.TodoRepository
}

func NewTodoService(repo port.TodoRepository) *TodoService {
	return &TodoService{repo}
}

func (ts *TodoService) GetTodosWithPagination(ctx context.Context, userId int, limit int, cursor string) (*response.CursorResponse, error) {
	rows, hasNext, err := ts.repo.GetAllWithCursor(ctx, userId, limit, cursor)

	data := make([]response.TodoResponse, 0)

	if err != nil {
		dataBytes, _ := util.Serialize(data)

		resp := response.CursorResponse{
			Size: 0,
			Data: dataBytes,
			Pagination: struct {
				HasNext    bool   `json:"has_next"`
				NextCursor string `json:"next_cursor"`
			}{
				HasNext:    false,
				NextCursor: "",
			},
		}

		return &resp, err
	}

	for _, todo := range rows {
		item := response.TodoResponse{
			UUID:        todo.UUID,
			Title:       todo.Title,
			Description: todo.Description,
			Status:      todo.StatusOrFallback(),
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt,
			UpdatedAt:   todo.UpdatedAt,
		}

		data = append(data, item)
	}

	var nextCursor string

	if hasNext && len(rows) > 0 {
		lastTodo := rows[len(rows)-1]
		nextCursor = util.EncodeCursor(lastTodo.CreatedAt.Format(time.RFC3339), lastTodo.ID)
	}

	dataBytes, _ := util.Serialize(data)

	responsable := response.CursorResponse{
		Size: len(data),
		Data: dataBytes,
		Pagination: struct {
			HasNext    bool   `json:"has_next"`
			NextCursor string `json:"next_cursor"`
		}{
			HasNext:    hasNext,
			NextCursor: nextCursor,
		},
	}

	return &responsable, nil
}

func (ts *TodoService) Create(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	now := time.Now()

	newTodo := domain.Todo{
		UUID:        uuid.New(),
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.Status,
		Completed:   todo.Completed,
		UserId:      todo.UserId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	todo, err := ts.repo.Create(ctx, newTodo)

	if err != nil {
		slog.Error(" Repository create failed", "error", err, "title", newTodo.Title)
		return domain.Todo{}, err
	}

	return todo, nil
}

func (u *TodoService) UpdateByUUID(ctx context.Context, todo domain.Todo) (domain.Todo, error) {
	todo, err := u.repo.UpdateByUUID(ctx, todo)

	if err != nil {
		return domain.Todo{}, err
	}

	return todo, nil
}

func (u *TodoService) DeleteByUUID(ctx context.Context, uid string) error {
	err := u.repo.DeleteByUUID(ctx, uid)

	if err != nil {
		return err
	}

	return nil
}

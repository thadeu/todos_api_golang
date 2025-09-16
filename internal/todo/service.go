package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	c "todoapp/pkg/db/cursor"
	. "todoapp/pkg/http"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

type TodoService struct {
	repo *TodoRepository
}

func NewTodoService(repo *TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) StatusOrFallback(todo Todo, fallback ...string) string {
	status := func() string {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		return TodoStatus(todo.Status).String()
	}()

	if status == "" {
		if len(fallback) > 0 && fallback[0] != "" {
			status = fallback[0]
		} else {
			status = "unknown"
		}
	}

	return status
}

func (s *TodoService) GetTodosWithPagination(ctx context.Context, userId int, limit int, cursor string) (*c.CursorResponse, error) {
	ctx, span := CreateChildSpan(ctx, "service.todo.GetTodosWithPagination", []attribute.KeyValue{
		attribute.Int("user.id", userId),
		attribute.Int("todo.limit", limit),
		attribute.String("todo.cursor", cursor),
	})
	defer span.End()

	rows, hasNext, err := s.repo.GetAllWithCursor(ctx, userId, limit, cursor)

	data := make([]TodoResponse, 0)

	if err != nil {
		AddSpanError(span, err)
		dataBytes, _ := json.Marshal(data)
		response := c.CursorResponse{
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

		return &response, err
	}

	for _, todo := range rows {
		item := TodoResponse{
			UUID:        todo.UUID,
			Title:       todo.Title,
			Description: todo.Description,
			Status:      s.StatusOrFallback(todo),
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt,
			UpdatedAt:   todo.UpdatedAt,
		}

		data = append(data, item)
	}

	var nextCursor string

	if hasNext && len(rows) > 0 {
		lastTodo := rows[len(rows)-1]
		nextCursor = c.EncodeCursor(lastTodo.CreatedAt.Format(time.RFC3339), lastTodo.ID)
	}

	dataBytes, _ := json.Marshal(data)

	response := c.CursorResponse{
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

	span.SetAttributes(
		attribute.Int("todo.count", len(data)),
		attribute.Bool("todo.has_next", hasNext),
	)

	return &response, nil
}

func (s *TodoService) GetAllTodos(userId int) ([]TodoResponse, error) {
	rows, err := s.repo.GetAll(userId)

	data := make([]TodoResponse, 0)

	if err != nil {
		return data, err
	}

	for _, todo := range rows {
		item := TodoResponse{
			UUID:        todo.UUID,
			Title:       todo.Title,
			Description: todo.Description,
			Status:      s.StatusOrFallback(todo),
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt,
			UpdatedAt:   todo.UpdatedAt,
		}

		data = append(data, item)
	}

	return data, nil
}

func (s *TodoService) CreateTodo(ctx context.Context, c *gin.Context, userId int) (Todo, error) {
	var params TodoRequest

	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return Todo{}, err
	}

	if err := Validator.Struct(params); err != nil {
		slog.Error("Falha na validação dos parâmetros do Todo", "error", err)
		return Todo{}, err
	}

	statusInt := 0

	if params.Status != "" {
		statusInt, err = StatusToEnum(params.Status)
		if err != nil {
			return Todo{}, err
		}
	}

	now := time.Now()

	newTodo := Todo{
		UUID:        uuid.New(),
		Title:       params.Title,
		Description: params.Description,
		Status:      statusInt,
		Completed:   params.Completed,
		UserId:      userId,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	if err := Validator.Struct(newTodo); err != nil {
		errors := FormatValidationErrors(err)
		slog.Error("Falha na validação do Todo", "errors", errors)

		return Todo{}, fmt.Errorf("%v", errors[len(errors)-1].Message)
	}

	todo, err := s.repo.Create(ctx, newTodo)

	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (s *TodoService) UpdateTodoByUUID(ctx context.Context, c *gin.Context, userId int) (Todo, error) {
	id := c.Param("uuid")

	var params TodoRequest
	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return Todo{}, err
	}

	if err := Validator.Struct(params); err != nil {
		slog.Error("Falha na validação do Todo", "error", err)
		return Todo{}, err
	}

	todo, err := s.repo.UpdateByUUID(ctx, id, userId, params)

	if err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func (s *TodoService) DeleteTodo(c *gin.Context, userId int) {
	id := c.Param("uuid")

	err := s.repo.DeleteById(id)

	if err != nil {
		slog.Error("Erro ao deletar todo", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Erro ao deletar todo",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deletado com sucesso"})
}

func (s *TodoService) DeleteByUUID(ctx context.Context, c *gin.Context, userId int) (map[string]any, error) {
	id := c.Param("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID é obrigatório")
	}

	_, err := s.repo.GetByUUID(ctx, id, userId)

	if err != nil {
		return nil, fmt.Errorf("desculpe, mas seu todo não foi encontrado")
	}

	if err := s.repo.DeleteByUUID(ctx, id); err != nil {
		return nil, err
	}

	return nil, nil
}

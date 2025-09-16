package impl

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

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/usecase/interfaces"
)

// TodoRequest represents the request structure for todo operations
type TodoRequest struct {
	Title       string     `json:"title" validate:"required,min=3,max=255"`
	Description string     `json:"description,omitempty" validate:"max=1000"`
	Status      string     `json:"status,omitempty"`
	Completed   bool       `json:"completed,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Implement TodoRequestInterface for persistence layer
func (tr TodoRequest) GetTitle() string       { return tr.Title }
func (tr TodoRequest) GetDescription() string { return tr.Description }
func (tr TodoRequest) GetStatus() string      { return tr.Status }
func (tr TodoRequest) GetCompleted() bool     { return tr.Completed }

// TodoResponse represents the response structure for todo data
type TodoResponse struct {
	UUID        uuid.UUID `json:"uuid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// todoUseCase implements the TodoUseCase interface
type todoUseCase struct {
	todoRepo repositories.TodoRepository
}

// NewTodoUseCase creates a new todo use case
func NewTodoUseCase(todoRepo repositories.TodoRepository) interfaces.TodoUseCase {
	return &todoUseCase{
		todoRepo: todoRepo,
	}
}

func (t *todoUseCase) StatusOrFallback(todo entities.Todo, fallback ...string) string {
	status := func() string {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		return entities.TodoStatus(todo.Status).String()
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

func (t *todoUseCase) GetTodosWithPagination(ctx context.Context, userId int, limit int, cursor string) (*c.CursorResponse, error) {
	ctx, span := CreateChildSpan(ctx, "usecase.todo.GetTodosWithPagination", []attribute.KeyValue{
		attribute.Int("user.id", userId),
		attribute.Int("todo.limit", limit),
		attribute.String("todo.cursor", cursor),
	})
	defer span.End()

	rows, hasNext, err := t.todoRepo.GetAllWithCursor(ctx, userId, limit, cursor)

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
			Status:      t.StatusOrFallback(todo),
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

func (t *todoUseCase) GetAllTodos(userId int) ([]interface{}, error) {
	rows, err := t.todoRepo.GetAll(userId)

	data := make([]TodoResponse, 0)

	if err != nil {
		return []interface{}{}, err
	}

	for _, todo := range rows {
		item := TodoResponse{
			UUID:        todo.UUID,
			Title:       todo.Title,
			Description: todo.Description,
			Status:      t.StatusOrFallback(todo),
			Completed:   todo.Completed,
			CreatedAt:   todo.CreatedAt,
			UpdatedAt:   todo.UpdatedAt,
		}

		data = append(data, item)
	}

	// Convert to []interface{} for compatibility
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = v
	}

	return result, nil
}

func (t *todoUseCase) CreateTodo(ctx context.Context, c *gin.Context, userId int) (entities.Todo, error) {
	var params TodoRequest

	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return entities.Todo{}, err
	}

	if err := Validator.Struct(params); err != nil {
		slog.Error("Validation failed for Todo parameters", "error", err)
		return entities.Todo{}, err
	}

	statusInt := 0

	if params.Status != "" {
		statusInt, err = StatusToEnum(params.Status)
		if err != nil {
			return entities.Todo{}, err
		}
	}

	now := time.Now()

	newTodo := entities.Todo{
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
		slog.Error("Validation failed for Todo", "errors", errors)

		return entities.Todo{}, fmt.Errorf("%v", errors[len(errors)-1].Message)
	}

	todo, err := t.todoRepo.Create(ctx, newTodo)

	if err != nil {
		return entities.Todo{}, err
	}

	return todo, nil
}

func (t *todoUseCase) UpdateTodoByUUID(ctx context.Context, c *gin.Context, userId int) (entities.Todo, error) {
	id := c.Param("uuid")

	var params TodoRequest
	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return entities.Todo{}, err
	}

	if err := Validator.Struct(params); err != nil {
		slog.Error("Validation failed for Todo", "error", err)
		return entities.Todo{}, err
	}

	todo, err := t.todoRepo.UpdateByUUID(ctx, id, userId, params)

	if err != nil {
		return entities.Todo{}, err
	}

	return todo, nil
}

func (t *todoUseCase) DeleteTodo(c *gin.Context, userId int) {
	id := c.Param("uuid")

	err := t.todoRepo.DeleteById(id)

	if err != nil {
		slog.Error("Error deleting todo", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting todo",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (t *todoUseCase) DeleteByUUID(ctx context.Context, c *gin.Context, userId int) (map[string]any, error) {
	id := c.Param("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := t.todoRepo.GetByUUID(ctx, id, userId)

	if err != nil {
		return nil, fmt.Errorf("sorry, but your todo was not found")
	}

	if err := t.todoRepo.DeleteByUUID(ctx, id); err != nil {
		return nil, err
	}

	return nil, nil
}

// StatusToEnum converts string status to enum
func StatusToEnum(status string) (int, error) {
	switch status {
	case "pending":
		return int(entities.TodoStatusPending), nil
	case "in_progress":
		return int(entities.TodoStatusInProgress), nil
	case "in_review":
		return int(entities.TodoStatusInReview), nil
	case "completed":
		return int(entities.TodoStatusCompleted), nil
	default:
		return 0, fmt.Errorf("invalid status: %s", status)
	}
}

package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/pkg/config"
	. "todoapp/pkg/http"
	. "todoapp/pkg/response"
	. "todoapp/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"todoapp/internal/usecase/interfaces"
)

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

// TodoHandler handles HTTP requests for todo operations
type TodoHandler struct {
	todoUseCase interfaces.TodoUseCase
	Logger      *LokiLogger
}

// NewTodoHandler creates a new todo handler
func NewTodoHandler(todoUseCase interfaces.TodoUseCase, logger *LokiLogger) *TodoHandler {
	return &TodoHandler{
		todoUseCase: todoUseCase,
		Logger:      logger,
	}
}

func (t *TodoHandler) GetAllTodos(c *gin.Context) {
	ctx, span := CreateChildSpan(c.Request.Context(), "handler.todo.GetAllTodos", []attribute.KeyValue{
		attribute.String("handler.operation", "GetAllTodos"),
		attribute.String("handler.method", c.Request.Method),
		attribute.String("handler.path", c.FullPath()),
	})

	defer span.End()

	userId := c.GetInt("x-user-id")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit"))

	if limit <= 0 {
		limit = 10
	}

	span.SetAttributes(
		attribute.Int("user.id", userId),
		attribute.String("todo.cursor", cursor),
		attribute.Int("todo.limit", limit),
	)

	data, err := t.todoUseCase.GetTodosWithPagination(ctx, userId, limit, cursor)

	if err != nil {
		AddSpanError(span, err)

		t.Logger.Logger.Ctx(ctx).Error("Failed to get todos",
			zap.Error(err),
			zap.Int("user_id", userId),
		)

		SendInternalError(c, "Error getting todos")
		return
	}

	span.SetAttributes(
		attribute.Int("http.status_code", http.StatusOK),
		attribute.String("response.type", "success"),
	)

	c.JSON(http.StatusOK, data)
}

func (t *TodoHandler) CreateTodo(c *gin.Context) {
	startTime := time.Now()

	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	todo, err := t.todoUseCase.CreateTodo(ctx, c, userId)

	if err != nil {
		slog.Error("Error creating todo", "error", err)

		if validationErrors := FormatValidationErrors(err); len(validationErrors) > 0 {
			SendValidationError(c, err)
			return
		}

		SendBadRequestError(c, "creation", err.Error())
		return
	}

	response := TodoResponse{
		UUID:        todo.UUID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      getStatusString(todo.Status),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})

	endTime := time.Now()
	slog.Info("Todo created", "time", endTime.Sub(startTime))
}

func (t *TodoHandler) UpdateTodo(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	todo, err := t.todoUseCase.UpdateTodoByUUID(ctx, c, userId)

	if err != nil {
		if validationErrors := FormatValidationErrors(err); len(validationErrors) > 0 {
			SendValidationError(c, err)
			return
		}

		SendBadRequestError(c, "update", err.Error())
		return
	}

	response := TodoResponse{
		UUID:        todo.UUID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      getStatusString(todo.Status),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (t *TodoHandler) DeleteByUUID(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	_, err := t.todoUseCase.DeleteByUUID(ctx, c, userId)

	if err != nil {
		SendNotFoundError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deleted successfully",
	})
}

// getStatusString converts status int to string
func getStatusString(status int) string {
	switch status {
	case 0:
		return "pending"
	case 1:
		return "in_progress"
	case 2:
		return "in_review"
	case 3:
		return "completed"
	default:
		return "unknown"
	}
}

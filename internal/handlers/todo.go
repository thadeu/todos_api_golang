package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/shared"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type TodoHandler struct {
	Service *TodoService
	Logger  *LokiLogger
}

func NewTodoHandler(service *TodoService, logger *LokiLogger) *TodoHandler {
	return &TodoHandler{Service: service, Logger: logger}
}

func (t *TodoHandler) GetAllTodos(c *gin.Context) {
	// Criar span manual para teste
	tracer := otel.Tracer("todoapp")
	ctx, span := tracer.Start(c.Request.Context(), "GetAllTodos-manual")
	defer span.End()

	userId := c.GetInt("x-user-id")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit"))

	if limit <= 0 {
		limit = 10
	}

	// Adicionar atributos ao span
	span.SetAttributes(
		attribute.Int("user.id", userId),
		attribute.String("todo.cursor", cursor),
		attribute.Int("todo.limit", limit),
	)

	data, err := t.Service.GetTodosWithPagination(ctx, userId, limit, cursor)

	if err != nil {
		t.Logger.Logger.Ctx(ctx).Error("Failed to get todos",
			zap.Error(err),
			zap.Int("user_id", userId),
		)
		SendInternalError(c, "Erro ao buscar todos")
		return
	}

	c.JSON(http.StatusOK, data)
}

func (t *TodoHandler) CreateTodo(c *gin.Context) {
	startTime := time.Now()

	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	todo, err := t.Service.CreateTodo(ctx, c, userId)

	if err != nil {
		slog.Error("Erro ao criar todo", "error", err)

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
		Status:      t.Service.StatusOrFallback(todo),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})

	endTime := time.Now()
	slog.Info("Todo criado", "time", endTime.Sub(startTime))
}

func (t *TodoHandler) UpdateTodo(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	todo, err := t.Service.UpdateTodoByUUID(ctx, c, userId)

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
		Status:      t.Service.StatusOrFallback(todo),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (t *TodoHandler) DeleteByUUID(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()
	_, err := t.Service.DeleteByUUID(ctx, c, userId)

	if err != nil {
		SendNotFoundError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deletado com sucesso",
	})
}

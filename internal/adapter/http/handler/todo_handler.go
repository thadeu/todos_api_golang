package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	. "todos/internal/adapter/http/helper"
	. "todos/internal/adapter/http/validation"
	"todos/internal/core/domain"
	"todos/internal/core/model/request"
	"todos/internal/core/model/response"
	"todos/internal/core/port"
	"todos/internal/core/util"
	"todos/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TodoHandler struct {
	svc    port.TodoService
	Logger *config.LokiLogger
}

func NewTodoHandler(todoUseCase port.TodoService, logger *config.LokiLogger) *TodoHandler {
	return &TodoHandler{
		svc:    todoUseCase,
		Logger: logger,
	}
}

func (t *TodoHandler) GetAllTodos(c *gin.Context) {
	ctx := c.Request.Context()
	userId, _ := c.Get("x-user-id")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit"))

	if limit <= 0 {
		limit = 10
	}

	data, err := t.svc.GetTodosWithPagination(ctx, userId.(int), limit, cursor)

	if err != nil {
		t.Logger.Logger.Ctx(ctx).Error("Failed to get todos",
			zap.Error(err),
			zap.Int("user_id", userId.(int)),
		)

		SendInternalError(c, "Error getting todos")
		return
	}

	c.JSON(http.StatusOK, data)
}

func (t *TodoHandler) CreateTodo(c *gin.Context) {
	ctx := c.Request.Context()

	userId, _ := c.Get("x-user-id")

	params, err := util.ParamsToMap[request.TodoRequest](c)

	if err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	todo := domain.Todo{
		Title:       params.Title,
		Description: params.Description,
		Completed:   params.Completed,
		UserId:      userId.(int),
	}

	status, err := todo.StatusToEnum(params.Status)

	if err != nil {
		SendBadRequestError(c, "status", err.Error())
		return
	}

	todo.Status = status

	if err := Validator.Struct(todo); err != nil {
		SendValidationError(c, err)
		return
	}

	slog.Info("Todo#create", "todo", todo)

	todo, err = t.svc.Create(ctx, todo)

	if err != nil {
		slog.Error("Error creating todo", "error", err)

		if validationErrors := FormatValidationErrors(err); len(validationErrors) > 0 {
			SendValidationError(c, err)
			return
		}

		SendBadRequestError(c, "creation", err.Error())
		return
	}

	response := response.TodoResponse{
		UUID:        todo.UUID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.StatusOrFallback(),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	SendSuccess(c, http.StatusCreated, response)
}

func (t *TodoHandler) UpdateTodo(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	ctx := c.Request.Context()

	params, err := util.ParamsToMap[request.TodoRequest](c)

	if err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	todo := domain.Todo{
		UUID:        uuid.MustParse(c.Param("uuid")),
		Title:       params.Title,
		Description: params.Description,
		Completed:   params.Completed,
		UserId:      userId,
	}

	status, err := todo.StatusToEnum(params.Status)

	if err != nil {
		SendBadRequestError(c, "status", err.Error())
		return
	}

	todo.Status = status

	todo, err = t.svc.UpdateByUUID(ctx, todo)

	if err != nil {
		if validationErrors := FormatValidationErrors(err); len(validationErrors) > 0 {
			SendValidationError(c, err)
			return
		}

		SendBadRequestError(c, "update", err.Error())
		return
	}

	response := response.TodoResponse{
		UUID:        todo.UUID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      todo.StatusOrFallback(),
		Completed:   todo.Completed,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (t *TodoHandler) DeleteByUUID(c *gin.Context) {
	ctx := c.Request.Context()

	err := t.svc.DeleteByUUID(ctx, c.Param("uuid"))

	if err != nil {
		SendNotFoundError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deleted successfully",
	})
}

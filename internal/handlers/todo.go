package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	"todoapp/internal/shared"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	Service *TodoService
}

func NewTodoHandler(service *TodoService) *TodoHandler {
	return &TodoHandler{Service: service}
}

func (t *TodoHandler) GetAllTodos(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit"))

	if limit <= 0 {
		limit = 10
	}

	data, err := t.Service.GetTodosWithPagination(userId, limit, cursor)

	if err != nil {
		slog.Error("Erro ao buscar todos", "error", err)
		shared.SendInternalError(c, "Erro ao buscar todos")
		return
	}

	c.JSON(http.StatusOK, data)
}

func (t *TodoHandler) CreateTodo(c *gin.Context) {
	startTime := time.Now()

	userId := c.GetInt("x-user-id")
	todo, err := t.Service.CreateTodo(c, userId)

	if err != nil {
		slog.Error("Erro ao criar todo", "error", err)

		if validationErrors := shared.FormatValidationErrors(err); len(validationErrors) > 0 {
			shared.SendValidationError(c, err)
			return
		}

		shared.SendBadRequestError(c, "creation", err.Error())
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
	todo, err := t.Service.UpdateTodoByUUID(c, userId)

	if err != nil {
		if validationErrors := shared.FormatValidationErrors(err); len(validationErrors) > 0 {
			shared.SendValidationError(c, err)
			return
		}

		shared.SendBadRequestError(c, "update", err.Error())
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
	_, err := t.Service.DeleteByUUID(c, userId)

	if err != nil {
		shared.SendNotFoundError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Todo deletado com sucesso",
	})
}

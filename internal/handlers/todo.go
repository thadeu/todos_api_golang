package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"

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
		slog.Error("Error fetching todos", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error fetching todos",
		})

		return
	}

	c.JSON(http.StatusOK, data)
}

func (t *TodoHandler) CreateTodo(c *gin.Context) {
	startTime := time.Now()

	userId := c.GetInt("x-user-id")
	todo, err := t.Service.CreateTodo(c, userId)

	if err != nil {
		slog.Error("Error creating todo", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"messages": []string{err.Error()},
		})

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

	c.JSON(http.StatusAccepted, response)

	endTime := time.Now()
	slog.Info("Todo created", "time", endTime.Sub(startTime))
}

func (t *TodoHandler) UpdateTodo(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	todo, err := t.Service.UpdateTodoByUUID(c, userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{err.Error()},
		})

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

	c.JSON(http.StatusOK, response)
}

func (t *TodoHandler) DeleteByUUID(c *gin.Context) {
	userId := c.GetInt("x-user-id")
	_, err := t.Service.DeleteByUUID(c, userId)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"messages": []string{err.Error()},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

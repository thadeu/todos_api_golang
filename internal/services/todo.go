package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	m "todoapp/internal/models"
	ru "todoapp/internal/repositories"
	c "todoapp/pkg/cursor"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TodoService struct {
	repo *ru.TodoRepository
}

func NewTodoService(repo *ru.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) StatusOrFallback(todo m.Todo, fallback ...string) string {
	status := func() string {
		defer func() {
			if r := recover(); r != nil {
				// slog.Warn("Invalid todo status, using fallback", "status", todo.Status, "uuid", todo.UUID)
			}
		}()

		return ru.TodoStatus(todo.Status).String()
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

func (s *TodoService) GetTodosWithPagination(userId int, limit int, cursor string) (*c.CursorResponse, error) {
	rows, hasNext, err := s.repo.GetAllWithCursor(userId, limit, cursor)

	data := make([]ru.TodoResponse, 0)

	if err != nil {
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
		item := ru.TodoResponse{
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

	return &response, nil
}

func (s *TodoService) GetAllTodos(userId int) ([]ru.TodoResponse, error) {
	rows, err := s.repo.GetAll(userId)

	data := make([]ru.TodoResponse, 0)

	if err != nil {
		return data, err
	}

	for _, todo := range rows {
		item := ru.TodoResponse{
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

func (s *TodoService) CreateTodo(c *gin.Context, userId int) (m.Todo, error) {
	var params ru.TodoRequest

	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return m.Todo{}, err
	}

	// Converter status string para int
	statusInt := 0 // default para pending
	if params.Status != "" {
		statusInt, err = ru.StatusToEnum(params.Status)
		if err != nil {
			return m.Todo{}, err
		}
	}

	now := time.Now()

	newTodo := m.Todo{
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

	todo, err := s.repo.Create(newTodo)

	if err != nil {
		return m.Todo{}, err
	}

	return todo, nil
}

func (s *TodoService) UpdateTodoByUUID(c *gin.Context, userId int) (m.Todo, error) {
	id := c.Param("uuid")

	var params ru.TodoRequest
	err := json.NewDecoder(c.Request.Body).Decode(&params)

	if err != nil {
		return m.Todo{}, err
	}

	// A conversão de status string para int é feita no repository

	todo, err := s.repo.UpdateByUUID(id, userId, params)

	if err != nil {
		return m.Todo{}, err
	}

	return todo, nil
}

func (s *TodoService) DeleteTodo(c *gin.Context, userId int) {
	id := c.Param("uuid")

	err := s.repo.DeleteById(id)

	if err != nil {
		slog.Error("Error deleting todo", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting todo",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func (s *TodoService) DeleteByUUID(c *gin.Context, userId int) (map[string]any, error) {
	id := c.Param("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := s.repo.GetByUUID(id, userId)

	if err != nil {
		return nil, fmt.Errorf("%s", "Sorry, but your todo is not found")
	}

	if err := s.repo.DeleteByUUID(id); err != nil {
		return nil, err
	}

	return nil, nil
}

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
		nextCursor = c.EncodeCursor("", rows[len(rows)-1].ID)
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

func (s *TodoService) CreateTodo(r *http.Request, userId int) (m.Todo, error) {
	var params ru.TodoRequest

	err := json.NewDecoder(r.Body).Decode(&params)

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

func (s *TodoService) UpdateTodoByUUID(r *http.Request, userId int) (m.Todo, error) {
	id := r.PathValue("uuid")

	var params ru.TodoRequest
	err := json.NewDecoder(r.Body).Decode(&params)

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

func (s *TodoService) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("uuid")

	err := s.repo.DeleteById(id)

	if err != nil {
		slog.Error("Error deleting todo", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	resp := map[string]any{
		"message": "User deleted successfully",
	}

	json.NewEncoder(w).Encode(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *TodoService) DeleteByUUID(r *http.Request, userId int) (map[string]any, error) {
	id := r.PathValue("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := s.repo.GetByUUID(id, userId)

	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteByUUID(id); err != nil {
		return nil, err
	}

	return nil, nil
}

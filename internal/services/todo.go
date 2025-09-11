package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	m "todoapp/internal/models"
	ru "todoapp/internal/repositories"

	"github.com/google/uuid"
)

type TodoService struct {
	repo *ru.TodoRepository
}

func NewTodoService(repo *ru.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
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

	now := time.Now()

	newUser := m.Todo{
		UUID:        uuid.New(),
		Title:       params.Title,
		Description: params.Description,
		Completed:   params.Completed,
		UserId:      userId,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	todo, err := s.repo.Create(newUser)

	if err != nil {
		return m.Todo{}, err
	}

	return todo, nil
}

func (s *TodoService) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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

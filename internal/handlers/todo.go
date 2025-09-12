package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/shared"
)

type TodoHandler struct {
	Service *TodoService
}

func NewTodoHandler(service *TodoService) *TodoHandler {
	return &TodoHandler{Service: service}
}

func (t *TodoHandler) Register() {
	http.HandleFunc("GET /todos", JwtAuthMiddleware(t.GetAllTodos))
	http.HandleFunc("POST /todos", JwtAuthMiddleware(t.CreateTodo))
	http.HandleFunc("PUT /todo/{uuid}", JwtAuthMiddleware(t.UpdateTodo))
	http.HandleFunc("DELETE /todos/{uuid}", JwtAuthMiddleware(t.DeleteByUUID))
}

func (t *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userId := r.Context().Value("x-user-id").(int)
	cursor := r.URL.Query().Get("cursor")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Default limit if not specified
	if limit <= 0 {
		limit = 10
	}

	data, err := t.Service.GetTodosWithPagination(userId, limit, cursor)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"message": "Error fetching todos"})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(data)
}

func (t *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	userId := r.Context().Value("x-user-id").(int)
	todo, err := t.Service.CreateTodo(r, userId)

	if err != nil {
		slog.Error("Error creating todo", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(response)

	endTime := time.Now()
	slog.Info("Todo created", "time", endTime.Sub(startTime))
}

func (t *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("x-user-id").(int)
	todo, err := t.Service.UpdateTodoByUUID(r, userId)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

func (t *TodoHandler) DeleteByUUID(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("x-user-id").(int)
	_, err := t.Service.DeleteByUUID(r, userId)

	if err != nil {
		slog.Error("Error deleting todo", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{"message": "Todo deleted successfully"})
}

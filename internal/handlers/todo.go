package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
)

type TodoHandler struct {
	Service *TodoService
}

func NewTodoHandler(service *TodoService) *TodoHandler {
	return &TodoHandler{Service: service}
}

func (t *TodoHandler) Register() {
	http.HandleFunc("GET /todos", t.GetAllTodos)
	http.HandleFunc("POST /todos", t.CreateTodo)
	http.HandleFunc("DELETE /todos/{uuid}", t.DeleteByUUID)
}

func (t *TodoHandler) GetAllTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userId, err := strconv.Atoi(r.Header.Get("X-User-ID"))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"message": "Unauthorized request"})

		return
	}

	users, err := t.Service.GetAllTodos(userId)

	if err != nil {
		slog.Error("Error fetching todos", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"message": "Error fetching todos"})

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(GetAllTodosResponse{Data: users, Size: len(users)})
}

func (t *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	userId, err := strconv.Atoi(r.Header.Get("X-User-ID"))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"message": "Unauthorized request"})

		return
	}

	todo, err := t.Service.CreateTodo(r, userId)

	if err != nil {
		slog.Error("Error creating user", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	response := TodoResponse{
		UUID:        todo.UUID,
		Title:       todo.Title,
		Description: todo.Description,
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

func (t *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {}

func (t *TodoHandler) DeleteByUUID(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.Header.Get("X-User-ID"))

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"message": "Unauthorized request"})

		return
	}

	_, err = t.Service.DeleteByUUID(r, userId)

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

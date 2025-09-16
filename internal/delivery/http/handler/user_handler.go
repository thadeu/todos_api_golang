package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"todoapp/internal/usecase/interfaces"
)

// UserResponse represents the response structure for user data
type UserResponse struct {
	UUID      string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Email     string     `json:"email,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase interfaces.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase interfaces.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userUseCase.GetAllUsers()
	if err != nil {
		slog.Error("Error getting users", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, err := h.userUseCase.CreateUser(ctx, r)
	if err != nil {
		slog.Error("Error creating user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		UUID:      user.UUID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	h.userUseCase.DeleteUser(w, r)
}

func (h *UserHandler) DeleteByUUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := h.userUseCase.DeleteByUUID(ctx, r)
	if err != nil {
		slog.Error("Error deleting user", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "User deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
)

type UserHandler struct {
	Service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{Service: service}
}

func (u *UserHandler) Register() {
	http.HandleFunc("GET /users", u.GetAllUsers)
	http.HandleFunc("POST /users", u.CreateUser)
	http.HandleFunc("DELETE /users/{uuid}", u.DeleteByUUID)
}

func (u *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := u.Service.GetAllUsers()

	if err != nil {
		slog.Error("Error fetching users", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(GetAllUsersResponse{Data: users, Size: len(users)})
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	slog.Info("Creating new user", "method", r.Method, "path", r.URL.Path)

	startTime := time.Now()

	user, err := u.Service.CreateUser(r)

	if err != nil {
		slog.Error("Error creating user", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	response := UserResponse{
		UUID:      user.UUID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(response)

	endTime := time.Now()
	slog.Info("User created", "time", endTime.Sub(startTime))

}

func (u *UserHandler) DeleteByUUID(w http.ResponseWriter, r *http.Request) {
	_, err := u.Service.DeleteByUUID(r)

	if err != nil {
		slog.Error("Error deleting user", "error", err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{"message": "User deleted successfully"})
}

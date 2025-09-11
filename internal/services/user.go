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

type UserService struct {
	repo *ru.UserRepository
}

func NewUserService(repo *ru.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAllUsers() ([]ru.UserResponse, error) {
	rows, err := s.repo.GetAllUsers()

	data := make([]ru.UserResponse, 0)

	if err != nil {
		return data, err
	}

	for _, user := range rows {
		item := ru.UserResponse{
			UUID:      user.UUID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
		}

		data = append(data, item)
	}

	return data, nil
}

func (s *UserService) CreateUser(r *http.Request) (m.User, error) {
	var params ru.UserRequest

	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return m.User{}, err
	}

	now := time.Now()

	newUser := m.User{
		UUID:      uuid.New(),
		Name:      params.Name,
		Email:     params.Email,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	user, err := s.repo.CreateUser(newUser)

	if err != nil {
		return m.User{}, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := s.repo.DeleteUser(id)

	if err != nil {
		slog.Error("Error deleting user", "error", err)

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

func (s *UserService) DeleteByUUID(r *http.Request) (map[string]any, error) {
	id := r.PathValue("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := s.repo.GetUserByUUID(id)

	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteByUUID(id); err != nil {
		return nil, err
	}

	return nil, nil
}

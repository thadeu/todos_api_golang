package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAllUsers() ([]UserResponse, error) {
	rows, err := s.repo.GetAllUsers()

	data := make([]UserResponse, 0)

	if err != nil {
		return data, err
	}

	for _, user := range rows {
		item := UserResponse{
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

func (s *UserService) CreateUser(ctx context.Context, r *http.Request) (User, error) {
	var params UserRequest

	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return User{}, err
	}

	now := time.Now()

	newUser := User{
		UUID:      uuid.New(),
		Name:      params.Name,
		Email:     params.Email,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	user, err := s.repo.CreateUser(ctx, newUser)

	if err != nil {
		return User{}, err
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

func (s *UserService) DeleteByUUID(ctx context.Context, r *http.Request) (map[string]any, error) {
	id := r.PathValue("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := s.repo.GetUserByUUID(id)

	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteByUUID(ctx, id); err != nil {
		return nil, err
	}

	return nil, nil
}

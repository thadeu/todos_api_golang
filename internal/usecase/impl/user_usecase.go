package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/usecase/interfaces"
)

// UserRequest represents the request structure for user creation
type UserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// UserResponse represents the response structure for user data
type UserResponse struct {
	UUID      string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Email     string     `json:"email,omitempty"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	userRepo repositories.UserRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repositories.UserRepository) interfaces.UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (u *userUseCase) GetAllUsers() ([]interface{}, error) {
	rows, err := u.userRepo.GetAllUsers()

	data := make([]UserResponse, 0)

	if err != nil {
		return []interface{}{}, err
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

	// Convert to []interface{} for compatibility
	result := make([]interface{}, len(data))
	for i, v := range data {
		result[i] = v
	}

	return result, nil
}

func (u *userUseCase) CreateUser(ctx context.Context, r *http.Request) (entities.User, error) {
	var params UserRequest

	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		return entities.User{}, err
	}

	now := time.Now()

	newUser := entities.User{
		UUID:      uuid.New(),
		Name:      params.Name,
		Email:     params.Email,
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	user, err := u.userRepo.CreateUser(ctx, newUser)

	if err != nil {
		return entities.User{}, err
	}

	return user, nil
}

func (u *userUseCase) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := u.userRepo.DeleteUser(id)

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

func (u *userUseCase) DeleteByUUID(ctx context.Context, r *http.Request) (map[string]any, error) {
	id := r.PathValue("uuid")

	if id == "" {
		return nil, fmt.Errorf("ID is required")
	}

	_, err := u.userRepo.GetUserByUUID(id)

	if err != nil {
		return nil, err
	}

	if err := u.userRepo.DeleteByUUID(ctx, id); err != nil {
		return nil, err
	}

	return nil, nil
}

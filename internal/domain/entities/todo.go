package entities

import (
	"time"

	"github.com/google/uuid"
)

// TodoStatus represents the possible states of a todo
type TodoStatus int

const (
	TodoStatusPending TodoStatus = iota
	TodoStatusInProgress
	TodoStatusInReview
	TodoStatusCompleted
)

// String returns the string representation of the todo status
func (t TodoStatus) String() string {
	return []string{"pending", "in_progress", "in_review", "completed"}[t]
}

// Todo represents a todo entity in the domain
type Todo struct {
	ID          int
	UUID        uuid.UUID
	Title       string `validate:"required,min=2,max=255"`
	Description string `validate:"max=255"`
	Status      int
	Completed   bool `validate:"boolean"`
	UserId      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// IsDeleted checks if the todo is marked as deleted
func (t *Todo) IsDeleted() bool {
	return t.DeletedAt != nil
}

// BelongsToUser checks if the todo belongs to the specified user
func (t *Todo) BelongsToUser(userID int) bool {
	return t.UserId == userID
}

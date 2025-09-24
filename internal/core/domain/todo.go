package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TodoStatus int

const (
	TodoStatusPending TodoStatus = iota
	TodoStatusInProgress
	TodoStatusInReview
	TodoStatusCompleted
)

type Todo struct {
	ID          int
	UUID        uuid.UUID
	Title       string `validate:"min=3,max=255"`
	Description string `validate:"max=255"`
	Status      int    `validate:"oneof=0 1 2 3"`
	Completed   bool   `validate:"boolean"`
	UserId      int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func (t *Todo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID,
		"uuid":        t.UUID,
		"title":       t.Title,
		"description": t.Description,
		"status":      t.Status,
		"completed":   t.Completed,
		"user_id":     t.UserId,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
	}
}

func (t *Todo) IsDeleted() bool {
	return t.DeletedAt != nil
}

func (t *Todo) BelongsToUser(userID int) bool {
	return t.UserId == userID
}

func (t *Todo) StatusOrFallback(fallback ...string) string {
	status := func() string {
		defer func() {
			if r := recover(); r != nil {
			}
		}()

		return TodoStatus(t.Status).String()
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

func (t TodoStatus) String() string {
	return []string{"pending", "in_progress", "in_review", "completed"}[t]
}

func (t *Todo) StatusToEnum(status string) (int, error) {
	switch status {
	case "pending", "":
		return int(TodoStatusPending), nil
	case "in_progress":
		return int(TodoStatusInProgress), nil
	case "in_review":
		return int(TodoStatusInReview), nil
	case "completed":
		return int(TodoStatusCompleted), nil
	default:
		return -1, fmt.Errorf("invalid status: %s", status)
	}
}

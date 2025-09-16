package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTodo_IsDeleted(t *testing.T) {
	t.Run("should return false when DeletedAt is nil", func(t *testing.T) {
		todo := Todo{
			DeletedAt: nil,
		}

		assert.False(t, todo.IsDeleted())
	})

	t.Run("should return true when DeletedAt is not nil", func(t *testing.T) {
		now := time.Now()
		todo := Todo{
			DeletedAt: &now,
		}

		assert.True(t, todo.IsDeleted())
	})
}

func TestTodo_BelongsToUser(t *testing.T) {
	t.Run("should return true when todo belongs to user", func(t *testing.T) {
		todo := Todo{
			UserId: 123,
		}

		assert.True(t, todo.BelongsToUser(123))
	})

	t.Run("should return false when todo does not belong to user", func(t *testing.T) {
		todo := Todo{
			UserId: 123,
		}

		assert.False(t, todo.BelongsToUser(456))
	})
}

func TestTodoStatus_String(t *testing.T) {
	tests := []struct {
		status   TodoStatus
		expected string
	}{
		{TodoStatusPending, "pending"},
		{TodoStatusInProgress, "in_progress"},
		{TodoStatusInReview, "in_review"},
		{TodoStatusCompleted, "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

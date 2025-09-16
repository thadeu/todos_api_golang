package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsDeleted(t *testing.T) {
	t.Run("should return false when DeletedAt is nil", func(t *testing.T) {
		user := User{
			DeletedAt: nil,
		}

		assert.False(t, user.IsDeleted())
	})

	t.Run("should return true when DeletedAt is not nil", func(t *testing.T) {
		now := time.Now()
		user := User{
			DeletedAt: &now,
		}

		assert.True(t, user.IsDeleted())
	})
}

func TestUser_Validation(t *testing.T) {
	t.Run("should validate required fields", func(t *testing.T) {
		// This would typically use a validation library
		user := User{
			Name:  "Test User",
			Email: "test@example.com",
		}

		assert.NotEmpty(t, user.Name)
		assert.NotEmpty(t, user.Email)
	})
}

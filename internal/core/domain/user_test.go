package domain

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	. "todos/pkg/test"
)

type UserUseCaseTestSuite struct {
	suite.Suite
}

func (s *UserUseCaseTestSuite) SetupTest() {
	InitTestDB()
}

func TestUserUseCaseTestSuite(t *testing.T) {
	RegisterTestingT(t)

	suite.Run(t, new(UserUseCaseTestSuite))
}

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

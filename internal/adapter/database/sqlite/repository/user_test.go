package repository_test

import (
	"context"
	"testing"

	. "todos/pkg/test"

	"todos/internal/adapter/database/sqlite/repository"
	"todos/internal/core/domain"
	"todos/internal/core/port"
	"todos/internal/core/telemetry"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo port.UserRepository
}

func (s *UserRepositoryTestSuite) SetupTest() {
	db := InitTestDB()
	probe := telemetry.NewNoOpProbe() // Use NoOpProbe for tests

	s.repo = repository.NewUserRepository(db, probe)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestRepository_CreateUser_Success() {
	user, err := s.repo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), user.ID)
	assert.NotEmpty(s.T(), user.UUID)
	assert.Equal(s.T(), "Test User", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *UserRepositoryTestSuite) TestRepository_DeleteUser_Success() {
	ctx := context.Background()

	user, _ := s.repo.Create(ctx, domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	err := s.repo.DeleteByUUID(ctx, user.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.repo.GetByUUID(ctx, user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	ctx := context.Background()

	user, _ := s.repo.Create(ctx, domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	err := s.repo.DeleteByUUID(ctx, user.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.repo.GetByUUID(ctx, user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

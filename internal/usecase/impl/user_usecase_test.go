package impl

import (
	"context"
	"testing"

	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/interfaces"
)

type UserUseCaseTestSuite struct {
	suite.Suite
	UseCase interfaces.UserUseCase
	repo    repositories.UserRepository
}

func (s *UserUseCaseTestSuite) SetupTest() {
	db := InitTestDB()
	repo := persistence.NewUserRepository(db)
	s.UseCase = NewUserUseCase(repo)
	s.repo = repo
}

func (s *UserUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestUserUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUseCaseTestSuite))
}

func (s *UserUseCaseTestSuite) TestUseCase_GetAllUsers_Empty() {
	users, err := s.UseCase.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *UserUseCaseTestSuite) TestUseCase_GetAllUsers_WithData() {
	user1 := entities.User{
		UUID:  uuid.New(),
		Name:  "Test User 1",
		Email: "test1@example.com",
	}
	s.repo.CreateUser(context.Background(), user1)

	user2 := entities.User{
		UUID:  uuid.New(),
		Name:  "Test User 2",
		Email: "test2@example.com",
	}
	s.repo.CreateUser(context.Background(), user2)

	users, err := s.UseCase.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)
}

func (s *UserUseCaseTestSuite) TestUseCase_CreateUser_Success() {
	user, _ := s.repo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	assert.NotEmpty(s.T(), user.ID)
	assert.NotEmpty(s.T(), user.UUID)
	assert.Equal(s.T(), "Test User", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *UserUseCaseTestSuite) TestUseCase_DeleteByUUID_Success() {
	user, _ := s.repo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User 2",
		Email: "test@example.com",
	})

	assert.NotEmpty(s.T(), user.ID)
	assert.NotEmpty(s.T(), user.UUID)
	assert.Equal(s.T(), "Test User 2", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *UserUseCaseTestSuite) TestUseCase_DeleteByUUID_NotFound() {
	err := s.repo.DeleteByUUID(context.Background(), "non-existent-uuid")
	assert.Error(s.T(), err)
}

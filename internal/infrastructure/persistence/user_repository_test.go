package persistence

import (
	"context"
	"strconv"
	"testing"

	"todoapp/internal/domain/entities"
	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	repo *userRepository
}

func (s *UserRepositoryTestSuite) SetupTest() {
	db := InitTestDB()
	s.repo = NewUserRepository(db).(*userRepository)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
	users, err := s.repo.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *UserRepositoryTestSuite) TestRepository_GetAllUsers_WithData() {
	user1 := entities.User{
		UUID:  uuid.New(),
		Name:  "Test User 1",
		Email: "test1@example.com",
	}
	user2 := entities.User{
		UUID:  uuid.New(),
		Name:  "Test User 2",
		Email: "test2@example.com",
	}

	s.repo.CreateUser(context.Background(), user1)
	s.repo.CreateUser(context.Background(), user2)

	users, err := s.repo.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)
}

func (s *UserRepositoryTestSuite) TestRepository_CreateUser_Success() {
	user, err := s.repo.CreateUser(context.Background(), entities.User{
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
	user, _ := s.repo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	err := s.repo.DeleteUser(strconv.Itoa(user.ID))
	assert.NoError(s.T(), err)

	_, err = s.repo.GetUserByUUID(user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	user, _ := s.repo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	err := s.repo.DeleteByUUID(context.Background(), user.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.repo.GetUserByUUID(user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

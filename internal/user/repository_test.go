package user

import (
	"context"
	"strconv"
	"testing"

	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	setup *TestSetup[UserRepository]
}

func (s *UserRepositoryTestSuite) SetupTest() {
	db := InitTestDB()
	repo := NewUserRepository(db)
	s.setup = SetupTest(s.T(), repo)
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
	users, err := s.setup.Repo.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *UserRepositoryTestSuite) TestRepository_GetAllUsers_WithData() {
	s.setup.Repo.CreateUser(context.Background(), NewUser[User]())
	s.setup.Repo.CreateUser(context.Background(), NewUser[User]())

	users, err := s.setup.Repo.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)
}

func (s *UserRepositoryTestSuite) TestRepository_CreateUser_Success() {
	user, err := s.setup.Repo.CreateUser(context.Background(), User{
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
	user, _ := s.setup.Repo.CreateUser(context.Background(), NewUser[User]())

	err := s.setup.Repo.DeleteUser(strconv.Itoa(user.ID))
	assert.NoError(s.T(), err)

	_, err = s.setup.Repo.GetUserByUUID(user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

func (s *UserRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	user, _ := s.setup.Repo.CreateUser(context.Background(), NewUser[User]())

	err := s.setup.Repo.DeleteByUUID(context.Background(), user.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.setup.Repo.GetUserByUUID(user.UUID.String())

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

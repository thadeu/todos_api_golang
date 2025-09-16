package user

import (
	"context"
	"testing"

	. "todoapp/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite
	Service *UserService
	setup   *TestSetup[UserRepository]
}

func (s *ServiceTestSuite) SetupTest() {
	db := InitTestDB()
	repo := NewUserRepository(db)
	s.Service = NewUserService(repo)
	s.setup = SetupTest(s.T(), repo)
}

func (s *ServiceTestSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
}

func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) TestService_GetAllUsers_Empty() {
	users, err := s.Service.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *ServiceTestSuite) TestService_GetAllUsers_WithData() {
	user1 := NewUser[User]()
	s.setup.Repo.CreateUser(context.Background(), user1)

	user2 := NewUser[User]()
	s.setup.Repo.CreateUser(context.Background(), user2)

	users, err := s.Service.GetAllUsers()

	assert.NoError(s.T(), err)
	assert.Len(s.T(), users, 2)
}

func (s *ServiceTestSuite) TestService_CreateUser_Success() {
	user, _ := s.setup.Repo.CreateUser(context.Background(), NewUser[User](map[string]any{
		"Name":  "Test User",
		"Email": "test@example.com",
	}))

	assert.NotEmpty(s.T(), user.ID)
	assert.NotEmpty(s.T(), user.UUID)
	assert.Equal(s.T(), "Test User", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *ServiceTestSuite) TestService_DeleteByUUID_Success() {
	user, _ := s.setup.Repo.CreateUser(context.Background(), NewUser[User](map[string]any{
		"Name":  "Test User 2",
		"Email": "test@example.com",
	}))

	assert.NotEmpty(s.T(), user.ID)
	assert.NotEmpty(s.T(), user.UUID)
	assert.Equal(s.T(), "Test User 2", user.Name)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *ServiceTestSuite) TestService_DeleteByUUID_NotFound() {
	err := s.setup.Repo.DeleteByUUID(context.Background(), "non-existent-uuid")
	assert.Error(s.T(), err)
}

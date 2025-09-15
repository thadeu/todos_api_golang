package services_test

import (
	"context"
	"testing"
	"todoapp/internal/factories"

	. "todoapp/internal/models"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TodoServiceTestSuite struct {
	suite.Suite
	Service  *TodoService
	UserRepo *UserRepository
	setup    *TestSetup[TodoRepository]
}

func (s *TodoServiceTestSuite) SetupTest() {
	db := InitTestDB()
	repo := NewTodoRepository(db)
	s.UserRepo = NewUserRepository(db)
	s.Service = NewTodoService(repo)
	s.setup = SetupTest(s.T(), repo)
}

func (s *TodoServiceTestSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
}

func TestTodoServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *TodoServiceTestSuite) TestService_GetAllUsers_Empty() {
	todos, err := s.Service.GetAllTodos(0)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), todos)
}

func (s *TodoServiceTestSuite) TestService_GetAllUsers_WithData() {
	user, _ := s.UserRepo.CreateUser(context.Background(), factories.NewUser[User]())

	item1 := factories.NewTodo[Todo](map[string]any{
		"UserId": user.ID,
	})
	s.setup.Repo.Create(context.Background(), item1)

	item2 := factories.NewTodo[Todo]()
	s.setup.Repo.Create(context.Background(), item2)

	todos, err := s.Service.GetAllTodos(user.ID)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), todos, 2)
}

func (s *TodoServiceTestSuite) TestService_CreateUser_Success() {
	data, _ := s.setup.Repo.Create(context.Background(), factories.NewTodo[Todo](map[string]any{
		"Title": "Test Todo",
	}))

	assert.NotEmpty(s.T(), data.ID)
	assert.NotEmpty(s.T(), data.UUID)
	assert.Equal(s.T(), "Test Todo", data.Title)

}

func (s *TodoServiceTestSuite) TestService_DeleteByUUID_Success() {
	data, _ := s.setup.Repo.Create(context.Background(), factories.NewTodo[Todo](map[string]any{
		"Title": "Test Todo 2",
	}))

	assert.NotEmpty(s.T(), data.ID)
	assert.NotEmpty(s.T(), data.UUID)
	assert.Equal(s.T(), "Test Todo 2", data.Title)
}

func (s *TodoServiceTestSuite) TestService_DeleteByUUID_NotFound() {
	err := s.setup.Repo.DeleteByUUID(context.Background(), "non-existent-uuid")
	assert.Error(s.T(), err)
}

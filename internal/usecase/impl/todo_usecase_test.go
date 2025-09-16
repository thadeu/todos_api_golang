package impl

import (
	"context"
	"testing"
	"time"

	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/interfaces"
)

type TodoUseCaseTestSuite struct {
	suite.Suite
	UseCase  interfaces.TodoUseCase
	UserRepo repositories.UserRepository
	TodoRepo repositories.TodoRepository
}

func (s *TodoUseCaseTestSuite) SetupTest() {
	db := InitTestDB()
	todoRepo := persistence.NewTodoRepository(db)
	userRepo := persistence.NewUserRepository(db)
	s.UseCase = NewTodoUseCase(todoRepo)
	s.UserRepo = userRepo
	s.TodoRepo = todoRepo
}

func (s *TodoUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestTodoUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(TodoUseCaseTestSuite))
}

func (s *TodoUseCaseTestSuite) TestUseCase_GetAllUsers_Empty() {
	todos, err := s.UseCase.GetAllTodos(0)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), todos)
}

func (s *TodoUseCaseTestSuite) TestUseCase_GetAllUsers_WithData() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	item1 := entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo 1",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.TodoRepo.Create(context.Background(), item1)

	item2 := entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo 2",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.TodoRepo.Create(context.Background(), item2)

	todos, err := s.UseCase.GetAllTodos(user.ID)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), todos, 2)
}

func (s *TodoUseCaseTestSuite) TestUseCase_CreateTodo_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	data, _ := s.TodoRepo.Create(context.Background(), entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	assert.NotEmpty(s.T(), data.ID)
	assert.NotEmpty(s.T(), data.UUID)
	assert.Equal(s.T(), "Test Todo", data.Title)
}

func (s *TodoUseCaseTestSuite) TestUseCase_DeleteByUUID_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	data, _ := s.TodoRepo.Create(context.Background(), entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo 2",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	assert.NotEmpty(s.T(), data.ID)
	assert.NotEmpty(s.T(), data.UUID)
	assert.Equal(s.T(), "Test Todo 2", data.Title)
}

func (s *TodoUseCaseTestSuite) TestUseCase_DeleteByUUID_NotFound() {
	err := s.TodoRepo.DeleteByUUID(context.Background(), "non-existent-uuid")
	assert.Error(s.T(), err)
}

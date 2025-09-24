package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	. "todoapp/pkg/test"

	"github.com/stretchr/testify/assert"

	"todoapp/internal/adapter/database/sqlite/repository"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	"todoapp/internal/core/service"
)

type TodoUseCaseTestSuite struct {
	suite.Suite
	UseCase  service.TodoService
	UserRepo port.UserRepository
	TodoRepo port.TodoRepository
}

func (s *TodoUseCaseTestSuite) SetupTest() {
	db := InitTestDB()

	todoRepo := repository.NewTodoRepository(db)
	userRepo := repository.NewUserRepository(db)

	s.UseCase = *service.NewTodoService(todoRepo)
	s.UserRepo = userRepo

	s.TodoRepo = todoRepo
}

func (s *TodoUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestTodoUseCaseTestSuite(t *testing.T) {
	RegisterTestingT(t)

	suite.Run(t, new(TodoUseCaseTestSuite))
}

func (s *TodoUseCaseTestSuite) TestUseCase_GetAllUsers_Empty() {
	todos, err := s.UseCase.GetTodosWithPagination(context.Background(), 0, 1, "")

	// assert.NoError(s.T(), err)
	// assert.Empty(s.T(), todos)
	Expect(err).To(BeNil())

	// <[]uint8 | len:2, cap:8>: []
	Expect(todos.Data.MarshalJSON()).To(Equal([]byte("[]")))
}

func (s *TodoUseCaseTestSuite) TestUseCase_GetAllUsers_WithData() {
	user, _ := s.UserRepo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	item1 := domain.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo 1",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.TodoRepo.Create(context.Background(), item1)

	item2 := domain.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo 2",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.TodoRepo.Create(context.Background(), item2)

	todos, err := s.UseCase.GetTodosWithPagination(context.Background(), user.ID, 100, "")

	Expect(err).To(BeNil())
	Expect(todos.Size).To(Equal(2))
}

func (s *TodoUseCaseTestSuite) TestUseCase_CreateTodo_Success() {
	user, _ := s.UserRepo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	data, _ := s.TodoRepo.Create(context.Background(), domain.Todo{
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
	user, _ := s.UserRepo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	data, _ := s.TodoRepo.Create(context.Background(), domain.Todo{
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

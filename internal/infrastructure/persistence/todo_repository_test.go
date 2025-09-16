package persistence

import (
	"context"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"todoapp/internal/domain/entities"
	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TodoRepositoryTestSuite struct {
	suite.Suite
	TodoRepo *todoRepository
	UserRepo *userRepository
}

func (s *TodoRepositoryTestSuite) SetupTest() {
	db := InitTestDB()
	s.TodoRepo = NewTodoRepository(db).(*todoRepository)
	s.UserRepo = NewUserRepository(db).(*userRepository)
}

func (s *TodoRepositoryTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestTodoRepositoryTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(TodoRepositoryTestSuite))
}

func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
	users, err := s.TodoRepo.GetAll(0)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_WithData() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	todo := entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.TodoRepo.Create(context.Background(), todo)

	data, err := s.TodoRepo.GetAll(user.ID)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), data, 1)
}

func (s *TodoRepositoryTestSuite) TestRepository_CreateTodo_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	status := int(entities.TodoStatusPending)

	todo, err := s.TodoRepo.Create(context.Background(), entities.Todo{
		UUID:        uuid.New(),
		Title:       "My User",
		Description: "Some description",
		Status:      status,
		Completed:   false,
		UserId:      user.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DeletedAt:   nil,
	})

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), todo.ID)
	assert.NotEmpty(s.T(), todo.UUID)
	assert.Equal(s.T(), "My User", todo.Title)
	assert.Equal(s.T(), user.ID, todo.UserId)
}

func (s *TodoRepositoryTestSuite) TestRepository_DeleteUser_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	todo := entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	savedTodo, _ := s.TodoRepo.Create(context.Background(), todo)

	err := s.TodoRepo.DeleteById(strconv.Itoa(savedTodo.ID))
	assert.NoError(s.T(), err)

	_, err = s.TodoRepo.GetByUUID(context.Background(), savedTodo.UUID.String(), user.ID)

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

func (s *TodoRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), entities.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	todo := entities.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	savedTodo, _ := s.TodoRepo.Create(context.Background(), todo)

	err := s.TodoRepo.DeleteByUUID(context.Background(), savedTodo.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.TodoRepo.GetByUUID(context.Background(), savedTodo.UUID.String(), user.ID)

	Expect(err.Error()).To(ContainSubstring("no rows"))
}

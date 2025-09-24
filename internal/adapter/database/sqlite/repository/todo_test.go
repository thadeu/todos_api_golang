package repository_test

import (
	"context"
	"testing"
	"time"

	. "todoapp/pkg/test"

	"todoapp/internal/adapter/database/sqlite/repository"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TodoRepositoryTestSuite struct {
	suite.Suite
	TodoRepo port.TodoRepository
	UserRepo port.UserRepository
}

func (s *TodoRepositoryTestSuite) SetupTest() {
	db := InitTestDB()

	s.TodoRepo = repository.NewTodoRepository(db)
	s.UserRepo = repository.NewUserRepository(db)
}

func TestTodoRepositoryTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(TodoRepositoryTestSuite))
}

func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
	users, _, err := s.TodoRepo.GetAllWithCursor(context.Background(), 0, 10, "")

	Expect(err).To(BeNil())
	Expect(users).To(BeEmpty())
}

func (s *TodoRepositoryTestSuite) TestRepository_CreateTodo_Success() {
	user, _ := s.UserRepo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	status := int(domain.TodoStatusPending)

	todo, err := s.TodoRepo.Create(context.Background(), domain.Todo{
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

	Expect(err).To(BeNil())

	Expect(todo.Title).To(Equal("My User"))
	Expect(todo.UserId).To(Equal(user.ID))
}

func (s *TodoRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	user, _ := s.UserRepo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	todo := domain.Todo{
		UUID:      uuid.New(),
		Title:     "Test Todo",
		UserId:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	savedTodo, _ := s.TodoRepo.Create(context.Background(), todo)

	err := s.TodoRepo.DeleteByUUID(context.Background(), savedTodo.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.TodoRepo.GetByUUID(context.Background(), savedTodo.UUID.String())

	Expect(err.Error()).To(ContainSubstring("no rows"))
}

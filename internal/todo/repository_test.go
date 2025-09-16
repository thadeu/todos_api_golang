package todo_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	. "todoapp/internal/todo"
	. "todoapp/internal/user"
	. "todoapp/pkg/test"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TodoRepositoryTestSuite struct {
	suite.Suite
	UserRepo *UserRepository
	setup    *TestSetup[TodoRepository]
}

func (s *TodoRepositoryTestSuite) SetupTest() {
	db := InitTestDB()
	repo := NewTodoRepository(db)
	s.UserRepo = NewUserRepository(db)
	s.setup = SetupTest(s.T(), repo)
}

func (s *TodoRepositoryTestSuite) TearDownTest() {
	TeardownTest[TodoRepository](s.T(), s.setup)
}

func TestTodoRepositoryTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(TodoRepositoryTestSuite))
}

func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
	users, err := s.setup.Repo.GetAll(0)

	assert.NoError(s.T(), err)
	assert.Empty(s.T(), users)
}

func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_WithData() {
	user, _ := s.UserRepo.CreateUser(context.Background(), NewUser[User]())

	s.setup.Repo.Create(context.Background(), NewTodo[Todo](map[string]any{
		"UserId": user.ID,
	}))

	data, err := s.setup.Repo.GetAll(user.ID)

	assert.NoError(s.T(), err)
	assert.Len(s.T(), data, 1)
}

func (s *TodoRepositoryTestSuite) TestRepository_CreateTodo_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), NewUser[User]())

	status := int(TodoStatusPending)

	todo, err := s.setup.Repo.Create(context.Background(), Todo{
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
	user, _ := s.UserRepo.CreateUser(context.Background(), NewUser[User]())

	todo, _ := s.setup.Repo.Create(context.Background(), NewTodo[Todo](map[string]any{
		"UserId": user.ID,
	}))

	err := s.setup.Repo.DeleteById(strconv.Itoa(todo.ID))
	assert.NoError(s.T(), err)

	_, err = s.setup.Repo.GetByUUID(context.Background(), todo.UUID.String(), user.ID)

	assert.Contains(s.T(), err.Error(), "no rows")
	assert.Error(s.T(), err)
}

func (s *TodoRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
	user, _ := s.UserRepo.CreateUser(context.Background(), NewUser[User]())

	todo, _ := s.setup.Repo.Create(context.Background(), NewTodo[Todo](map[string]any{
		"UserId": user.ID,
	}))

	err := s.setup.Repo.DeleteByUUID(context.Background(), todo.UUID.String())
	assert.NoError(s.T(), err)

	_, err = s.setup.Repo.GetByUUID(context.Background(), todo.UUID.String(), user.ID)

	Expect(err.Error()).To(ContainSubstring("no rows"))
}

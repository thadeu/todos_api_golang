package repository_test

// import (
// 	"context"
// 	"fmt"
// 	"log/slog"
// 	"os"
// 	"testing"
// 	"time"

// 	// . "todoapp/pkg/test"

// 	"todoapp/internal/adapter/database/postgres"
// 	repository "todoapp/internal/adapter/database/postgres/repository"
// 	"todoapp/internal/core/domain"
// 	"todoapp/internal/core/port"

// 	"github.com/google/uuid"
// 	. "github.com/onsi/gomega"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/wait"
// )

// type TodoRepositoryTestSuite struct {
// 	suite.Suite
// 	TodoRepo    port.TodoRepository
// 	UserRepo    port.UserRepository
// 	pgContainer testcontainers.Container
// 	DB          *postgres.DB
// }

// func (s *TodoRepositoryTestSuite) SetupSuite() {
// 	ctx := context.Background()

// 	req := testcontainers.GenericContainerRequest{
// 		Started: true,
// 		ContainerRequest: testcontainers.ContainerRequest{
// 			Image:        "postgres:15-alpine",
// 			ExposedPorts: []string{"5432/tcp"},
// 			Env: map[string]string{
// 				"POSTGRES_DB":       "testdb",
// 				"POSTGRES_USER":     "test",
// 				"POSTGRES_PASSWORD": "test",
// 			},
// 			WaitingFor: wait.ForLog("Ready to accept connections"),
// 		},
// 	}

// 	pgContainer, _ := testcontainers.GenericContainer(ctx, req)

// 	host, _ := pgContainer.Host(ctx)
// 	port, _ := pgContainer.MappedPort(ctx, "5432")

// 	url := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
// 	slog.Info("DATABASE_URL", "url", url)

// 	os.Setenv("DATABASE_URL", url)
// 	newDb, err := postgres.NewDB()

// 	if err != nil {
// 		slog.Error("Error creating database", "error", err)
// 		panic(err)
// 	}

// 	s.DB = newDb
// 	s.TodoRepo = repository.NewTodoRepository(s.DB)
// 	s.UserRepo = repository.NewUserRepository(s.DB)
// }

// func TestTodoRepositoryTestSuite(t *testing.T) {
// 	RegisterTestingT(t)
// 	suite.Run(t, new(TodoRepositoryTestSuite))
// }

// func (s *TodoRepositoryTestSuite) TestRepository_GetAllUsers_Empty() {
// 	users, _, err := s.TodoRepo.GetAllWithCursor(context.Background(), 0, 10, "")

// 	Expect(err).To(BeNil())
// 	Expect(users).To(BeEmpty())
// }

// func (s *TodoRepositoryTestSuite) TestRepository_CreateTodo_Success() {
// 	user, _ := s.UserRepo.Create(context.Background(), domain.User{
// 		UUID:  uuid.New(),
// 		Name:  "Test User",
// 		Email: "test@example.com",
// 	})

// 	status := int(domain.TodoStatusPending)

// 	todo, err := s.TodoRepo.Create(context.Background(), domain.Todo{
// 		UUID:        uuid.New(),
// 		Title:       "My User",
// 		Description: "Some description",
// 		Status:      status,
// 		Completed:   false,
// 		UserId:      user.ID,
// 		CreatedAt:   time.Now(),
// 		UpdatedAt:   time.Now(),
// 		DeletedAt:   nil,
// 	})

// 	Expect(err).To(BeNil())

// 	Expect(todo.Title).To(Equal("My User"))
// 	Expect(todo.UserId).To(Equal(user.ID))
// }

// func (s *TodoRepositoryTestSuite) TestRepository_DeleteByUUID_Success() {
// 	user, _ := s.UserRepo.Create(context.Background(), domain.User{
// 		UUID:  uuid.New(),
// 		Name:  "Test User",
// 		Email: "test@example.com",
// 	})

// 	todo := domain.Todo{
// 		UUID:      uuid.New(),
// 		Title:     "Test Todo",
// 		UserId:    user.ID,
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}

// 	savedTodo, _ := s.TodoRepo.Create(context.Background(), todo)

// 	err := s.TodoRepo.DeleteByUUID(context.Background(), savedTodo.UUID.String())
// 	assert.NoError(s.T(), err)

// 	_, err = s.TodoRepo.GetByUUID(context.Background(), savedTodo.UUID.String())

// 	Expect(err.Error()).To(ContainSubstring("no rows"))
// }

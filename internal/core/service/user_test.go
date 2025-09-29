package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	. "todoapp/pkg/test"

	"todoapp/internal/adapter/database/sqlite/repository"
	"todoapp/internal/core/domain"
	"todoapp/internal/core/port"
	"todoapp/internal/core/service"
	"todoapp/internal/core/telemetry"
)

type UserUseCaseTestSuite struct {
	suite.Suite
	UseCase *service.UserService
	repo    port.UserRepository
}

func (s *UserUseCaseTestSuite) SetupTest() {
	db := InitTestDB()
	probe := telemetry.NewNoOpProbe() // Use NoOpProbe for tests

	repo := repository.NewUserRepository(db, probe)

	s.UseCase = service.NewUserService(repo)
	s.repo = repo
}

func (s *UserUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestUserUseCaseTestSuite(t *testing.T) {
	RegisterTestingT(t)

	suite.Run(t, new(UserUseCaseTestSuite))
}

func (s *UserUseCaseTestSuite) TestUseCase_CreateUser_Success() {
	user, _ := s.repo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	})

	Expect(user.ID).To(BeNumerically(">", 0))
	Expect(user.UUID).NotTo(BeEmpty())
	Expect(user.Name).To(Equal("Test User"))
	Expect(user.Email).To(Equal("test@example.com"))
}

func (s *UserUseCaseTestSuite) TestUseCase_DeleteByUUID_Success() {
	user, _ := s.repo.Create(context.Background(), domain.User{
		UUID:  uuid.New(),
		Name:  "Test User 2",
		Email: "test2@example.com",
	})

	err := s.repo.DeleteByUUID(context.Background(), user.UUID.String())
	Expect(err).To(BeNil())

	_, err = s.repo.GetByUUID(context.Background(), user.UUID.String())
	Expect(err).To(HaveOccurred())
}

func (s *UserUseCaseTestSuite) TestUseCase_DeleteByUUID_NotFound() {
	err := s.repo.DeleteByUUID(context.Background(), "non-existent-uuid")

	Expect(err).To(HaveOccurred())
}

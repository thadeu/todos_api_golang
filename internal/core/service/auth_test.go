package service_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	. "todos/pkg/test"

	"github.com/stretchr/testify/assert"

	"todos/internal/adapter/database/sqlite/repository"
	"todos/internal/core/model/request"
	"todos/internal/core/port"
	"todos/internal/core/service"
	"todos/internal/core/telemetry"
)

type AuthUseCaseTestSuite struct {
	suite.Suite
	UseCase port.AuthService
	repo    port.UserRepository
}

func (s *AuthUseCaseTestSuite) SetupTest() {
	db := InitTestDB()
	probe := telemetry.NewNoOpProbe() // Use NoOpProbe for tests

	repo := repository.NewUserRepository(db, probe)

	s.UseCase = service.NewAuthService(repo)
	s.repo = repo
}

func (s *AuthUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestAuthUseCaseTestSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(AuthUseCaseTestSuite))
}

func (s *AuthUseCaseTestSuite) TestUseCase_Registration_Success() {
	req := &request.SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := s.UseCase.Registration(context.Background(), req)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), "test@example.com", user.Email)
}

func (s *AuthUseCaseTestSuite) TestUseCase_Registration_UserAlreadyExists() {
	req := &request.SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := s.UseCase.Registration(context.Background(), req)
	assert.NoError(s.T(), err)

	_, err = s.UseCase.Registration(context.Background(), req)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "user already exists")
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_Success() {
	signUpReq := &request.SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginReq := &request.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	createdUser, err := s.UseCase.Registration(context.Background(), signUpReq)
	assert.NoError(s.T(), err)

	authenticatedUser, err := s.UseCase.Authenticate(context.Background(), loginReq)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), authenticatedUser)
	assert.Equal(s.T(), createdUser.Email, authenticatedUser.Email)
	assert.Equal(s.T(), createdUser.UUID, authenticatedUser.UUID)
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_InvalidPassword() {
	signUpReq := &request.SignUpRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginFailedReq := &request.LoginRequest{
		Email:    "test@example.com",
		Password: "unknow-password",
	}

	_, err := s.UseCase.Registration(context.Background(), signUpReq)
	assert.NoError(s.T(), err)

	_, err = s.UseCase.Authenticate(context.Background(), loginFailedReq)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "compare password failed")
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_UserNotFound() {
	loginReq := &request.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := s.UseCase.Authenticate(context.Background(), loginReq)

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "authentication failed")
}

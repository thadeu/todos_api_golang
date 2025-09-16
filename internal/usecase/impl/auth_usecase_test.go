package impl

import (
	"context"
	"testing"

	. "todoapp/pkg/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/interfaces"
)

type AuthUseCaseTestSuite struct {
	suite.Suite
	UseCase interfaces.AuthUseCase
	repo    repositories.UserRepository
}

func (s *AuthUseCaseTestSuite) SetupTest() {
	db := InitTestDB()
	repo := persistence.NewUserRepository(db)
	s.UseCase = NewAuthUseCase(repo)
	s.repo = repo
}

func (s *AuthUseCaseTestSuite) TearDownTest() {
	// Cleanup if needed
}

func TestAuthUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(AuthUseCaseTestSuite))
}

func (s *AuthUseCaseTestSuite) TestUseCase_Registration_Success() {
	user, err := s.UseCase.Registration(context.Background(), "test@example.com", "password123")

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), "test@example.com", user.Email)
	assert.NotEmpty(s.T(), user.EncryptedPassword)
	assert.NotEmpty(s.T(), user.UUID)
}

func (s *AuthUseCaseTestSuite) TestUseCase_Registration_UserAlreadyExists() {
	// Create first user
	_, err := s.UseCase.Registration(context.Background(), "test@example.com", "password123")
	assert.NoError(s.T(), err)

	// Try to create user with same email
	_, err = s.UseCase.Registration(context.Background(), "test@example.com", "password123")
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "user already exists")
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_Success() {
	// Create user first
	createdUser, err := s.UseCase.Registration(context.Background(), "test@example.com", "password123")
	assert.NoError(s.T(), err)

	// Authenticate user
	authenticatedUser, err := s.UseCase.Authenticate(context.Background(), "test@example.com", "password123")

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), authenticatedUser)
	assert.Equal(s.T(), createdUser.Email, authenticatedUser.Email)
	assert.Equal(s.T(), createdUser.UUID, authenticatedUser.UUID)
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_InvalidPassword() {
	// Create user first
	_, err := s.UseCase.Registration(context.Background(), "test@example.com", "password123")
	assert.NoError(s.T(), err)

	// Try to authenticate with wrong password
	_, err = s.UseCase.Authenticate(context.Background(), "test@example.com", "wrongpassword")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "authentication failed")
}

func (s *AuthUseCaseTestSuite) TestUseCase_Authenticate_UserNotFound() {
	// Try to authenticate non-existent user
	_, err := s.UseCase.Authenticate(context.Background(), "nonexistent@example.com", "password123")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "authentication failed")
}

func (s *AuthUseCaseTestSuite) TestUseCase_GenerateRefreshToken_Success() {
	// Create user first
	user, err := s.UseCase.Registration(context.Background(), "test@example.com", "password123")
	assert.NoError(s.T(), err)

	// Generate token
	token, err := s.UseCase.GenerateRefreshToken(user)

	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), token)
}

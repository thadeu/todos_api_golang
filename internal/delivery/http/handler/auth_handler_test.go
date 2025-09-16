package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/impl"
	. "todoapp/pkg/response"
	. "todoapp/pkg/test"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerSuite struct {
	suite.Suite
	UserRepo repositories.UserRepository
	Router   *gin.Engine
	DB       *sql.DB
}

var globalAuthHandler *AuthHandler

func (s *AuthHandlerSuite) SetupSuite() {
	globalAuthHandler = &AuthHandler{}
}

func (s *AuthHandlerSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	s.DB = InitTestDB()
	s.UserRepo = persistence.NewUserRepository(s.DB)

	// Create use case and handler
	authUseCase := impl.NewAuthUseCase(s.UserRepo)
	globalAuthHandler = NewAuthHandler(authUseCase)

	// Setup router directly to avoid import cycle
	s.Router = setupTestRouter(globalAuthHandler)
}

func (s *AuthHandlerSuite) TearDownTest() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func TestAuthHandlerSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(AuthHandlerSuite))
}

func setupTestRouter(authHandler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Public routes
	public := router.Group("/")
	{
		public.POST("/signup", authHandler.RegisterByEmailAndPassword)
		public.POST("/auth", authHandler.AuthByEmailAndPassword)
	}

	return router
}

func (a *AuthHandlerSuite) TestSignUpUserSuccess() {
	reqBody := strings.NewReader(`{"email": "eu@test.com", "password": "12345678"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)

	rr := httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusCreated))

	body, _ := io.ReadAll(rr.Body)
	data := gin.H{}
	json.Unmarshal(body, &data)

	expectedMessage := fmt.Sprintf("User %s created successfully", "eu@test.com")

	Expect(data["message"]).To(Equal(expectedMessage))
}

func (a *AuthHandlerSuite) TestSignUpUserValidationError() {
	reqBody := strings.NewReader(`{"email": "invalid-email", "password": "123"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)

	rr := httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusBadRequest))

	body, _ := io.ReadAll(rr.Body)
	data := ErrorResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Error.Code).To(Equal("VALIDATION_ERROR"))
	Expect(len(data.Error.Errors)).To(BeNumerically(">", 0))
}

func (a *AuthHandlerSuite) TestAuthUserSuccess() {
	// First create a user
	reqBody := strings.NewReader(`{"email": "test@example.com", "password": "12345678"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	// Now authenticate
	reqBody = strings.NewReader(`{"email": "test@example.com", "password": "12345678"}`)
	req, _ = http.NewRequest("POST", "/auth", reqBody)
	rr = httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))

	body, _ := io.ReadAll(rr.Body)
	data := gin.H{}
	json.Unmarshal(body, &data)

	Expect(data["refresh_token"]).ToNot(BeEmpty())
}

func (a *AuthHandlerSuite) TestAuthUserInvalidCredentials() {
	reqBody := strings.NewReader(`{"email": "test@example.com", "password": "wrongpassword"}`)
	req, _ := http.NewRequest("POST", "/auth", reqBody)
	rr := httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusUnauthorized))

	body, _ := io.ReadAll(rr.Body)
	data := ErrorResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Error.Code).To(Equal("UNAUTHORIZED"))
	Expect(len(data.Error.Errors)).To(BeNumerically(">", 0))
	Expect(data.Error.Errors[0].Message).To(Equal("Invalid email or password"))
}

package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "todoapp/pkg/test"

	"todoapp/internal/adapter/database/sqlite/repository"
	"todoapp/internal/core/model/response"
	"todoapp/internal/core/port"
	"todoapp/internal/core/service"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerSuite struct {
	suite.Suite
	UserRepo port.UserRepository
	Router   *gin.Engine
	DB       *sql.DB
}

var globalAuthHandler *AuthHandler

func (s *AuthHandlerSuite) SetupSuite() {
	globalAuthHandler = &AuthHandler{}
}

func (s *AuthHandlerSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	db := InitTestDB()
	s.UserRepo = repository.NewUserRepository(db)

	authUseCase := service.NewAuthService(s.UserRepo)
	globalAuthHandler = NewAuthHandler(authUseCase)

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

	body, _ := io.ReadAll(rr.Body)

	data := gin.H{}
	json.Unmarshal(body, &data)

	newData := data["data"].(map[string]any)

	Expect(rr.Code).To(Equal(http.StatusCreated))
	Expect(newData["email"]).To(Equal("eu@test.com"))
}

func (a *AuthHandlerSuite) TestSignUpUserValidationError() {
	reqBody := strings.NewReader(`{"email": "invalid-email", "password": "123"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)

	rr := httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusBadRequest))

	body, _ := io.ReadAll(rr.Body)
	data := response.ErrorResponse{}
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
	data := response.ErrorResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Error.Code).To(Equal("UNAUTHORIZED"))
	Expect(len(data.Error.Errors)).To(BeNumerically(">", 0))
	Expect(data.Error.Errors[0].Message).To(Equal("Invalid email or password"))
}

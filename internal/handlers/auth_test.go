package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "todoapp/internal/api"
	. "todoapp/internal/handlers"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/test"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerSuite struct {
	suite.Suite
	UserRepo *UserRepository
	Router   *gin.Engine
	Setup    *TestSetup[UserRepository]
}

var globalAuthHandler *AuthHandler

func (s *AuthHandlerSuite) SetupSuite() {
	globalAuthHandler = &AuthHandler{}
}

func (s *AuthHandlerSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	db := InitTestDB()

	repo := NewUserRepository(db)
	s.Setup = SetupTest(s.T(), repo)
	s.UserRepo = NewUserRepository(db)

	globalAuthHandler.Service = NewAuthService(s.Setup.Repo)

	s.Router = api.SetupRouter(api.HandlersConfig{
		AuthHandler: globalAuthHandler,
	})
}

func (s *AuthHandlerSuite) TearDownTest() {
	TeardownTest(s.T(), s.Setup)
}

func TestAuthHandlerSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(AuthHandlerSuite))
}

type BodyResponse struct {
	Message string `json:"message"`
}

func (a *AuthHandlerSuite) TestSignUpUserSuccess() {
	reqBody := strings.NewReader(`{"email": "eu@test.com", "password": "12345678"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)

	rr := httptest.NewRecorder()

	// http.DefaultServeMux.ServeHTTP(rr, req)
	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))

	body, _ := io.ReadAll(rr.Body)
	data := BodyResponse{}
	json.Unmarshal(body, &data)

	expectedMessage := fmt.Sprintf("User %s was created successfully", "eu@test.com")

	Expect(data.Message).To(Equal(expectedMessage))
}

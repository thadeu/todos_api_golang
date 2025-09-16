package auth_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "todoapp/internal/auth"
	. "todoapp/internal/user"
	api "todoapp/pkg/api"
	. "todoapp/pkg/response"
	. "todoapp/pkg/test"

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

	s.Router = api.SetupRouterForTests(api.HandlersConfig{
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

func (a *AuthHandlerSuite) TestSignUpUserSuccess() {
	reqBody := strings.NewReader(`{"email": "eu@test.com", "password": "12345678"}`)
	req, _ := http.NewRequest("POST", "/signup", reqBody)

	rr := httptest.NewRecorder()

	a.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusCreated))

	body, _ := io.ReadAll(rr.Body)
	data := gin.H{}
	json.Unmarshal(body, &data)

	expectedMessage := fmt.Sprintf("Usuário %s foi criado com sucesso", "eu@test.com")

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

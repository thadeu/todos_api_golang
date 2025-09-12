package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "todoapp/internal/handlers"
	// . "todoapp/internal/models"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"

	// . "todoapp/internal/shared"
	. "todoapp/internal/test"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerSuite struct {
	suite.Suite
	UserRepo *UserRepository
	setup    *TestSetup[UserRepository]
}

var globalAuthHandler *AuthHandler

func (s *AuthHandlerSuite) SetupSuite() {
	globalAuthHandler = &AuthHandler{}
	globalAuthHandler.Register()
}

func (s *AuthHandlerSuite) SetupTest() {
	db := InitTestDB()

	repo := NewUserRepository(db)
	s.setup = SetupTest(s.T(), repo)
	s.UserRepo = NewUserRepository(db)

	globalAuthHandler.Service = NewAuthService(s.setup.Repo)
}

func (s *AuthHandlerSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
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

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))

	body, _ := io.ReadAll(rr.Body)
	data := BodyResponse{}
	json.Unmarshal(body, &data)

	expectedMessage := fmt.Sprintf("User %s was created successfully", "eu@test.com")

	Expect(data.Message).To(Equal(expectedMessage))
}

package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todoapp/internal/factories"

	. "todoapp/internal/handlers"
	. "todoapp/internal/models"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
	. "todoapp/internal/test"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"
)

type UserHandlerSuite struct {
	suite.Suite
	setup *TestSetup[UserRepository]
}

var globalHandler *UserHandler

func (s *UserHandlerSuite) SetupSuite() {
	globalHandler = &UserHandler{}
	globalHandler.Register()
}

func (s *UserHandlerSuite) SetupTest() {
	db := InitTestDB()
	repo := NewUserRepository(db)
	s.setup = SetupTest(s.T(), repo)
	globalHandler.Service = NewUserService(s.setup.Repo)
}

func (s *UserHandlerSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
}

func TestHandlerSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(UserHandlerSuite))
}

func (s *UserHandlerSuite) TestGetAllUsersWithData() {
	s.setup.Repo.CreateUser(factories.NewUser[User](map[string]any{
		"Name": "User1",
	}))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/users", nil)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := GetAllUsersResponse{}
	json.Unmarshal(body, &data)

	Expect(len(data.Data)).To(Equal(1))
	Expect(data.Size).To(Equal(1))

	first := data.Data[0]
	Expect(first.Name).To(Equal("User1"))
}

func (s *UserHandlerSuite) TestCreateUser() {
	reqBody := strings.NewReader(`{"name": "User2", "email": "user2@example.com"}`)

	req, _ := http.NewRequest("POST", "/users", reqBody)
	rr := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusAccepted))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := UserResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Name).To(Equal("User2"))
	Expect(data.Email).To(Equal("user2@example.com"))
	Expect(data.UUID).To(Not(BeEmpty()))
}

func (s *UserHandlerSuite) TestDeleteByUUIDWhenIdExists() {
	user, _ := s.setup.Repo.CreateUser(factories.NewUser[User](map[string]any{
		"Name": "User",
	}))

	path := fmt.Sprintf("/users/%s", user.UUID.String())
	req, _ := http.NewRequest("DELETE", path, nil)
	rr := httptest.NewRecorder()

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := map[string]any{}
	json.Unmarshal(body, &data)

	Expect(data["message"]).To(Equal("User deleted successfully"))
}

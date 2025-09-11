package handlers_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

type TodoHandlerSuite struct {
	suite.Suite
	UserRepo *UserRepository
	setup    *TestSetup[TodoRepository]
}

var globalTodoHandler *TodoHandler

func (s *TodoHandlerSuite) SetupSuite() {
	globalTodoHandler = &TodoHandler{}
	globalTodoHandler.Register()
}

func (s *TodoHandlerSuite) SetupTest() {
	db := InitTestDB()
	repo := NewTodoRepository(db)
	s.setup = SetupTest(s.T(), repo)
	s.UserRepo = NewUserRepository(db)
	globalTodoHandler.Service = NewTodoService(s.setup.Repo)
}

func (s *TodoHandlerSuite) TearDownTest() {
	TeardownTest(s.T(), s.setup)
}

func TestTodoHandlerSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(TodoHandlerSuite))
}

func CreateUser(s *TodoHandlerSuite) User {
	user, _ := s.UserRepo.CreateUser(factories.NewUser[User](map[string]any{
		"Name": "User1",
	}))

	return user
}

func (s *TodoHandlerSuite) TestGetAllUsersWithData() {
	user := CreateUser(s)

	s.setup.Repo.Create(factories.NewTodo[Todo](map[string]any{
		"Title":  "User1",
		"UserId": user.ID,
	}))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todos", nil)

	req.Header.Set("X-User-ID", strconv.Itoa(user.ID))

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := GetAllTodosResponse{}
	json.Unmarshal(body, &data)

	Expect(len(data.Data)).To(Equal(1))
	Expect(data.Size).To(Equal(1))

	first := data.Data[0]
	Expect(first.Title).To(Equal("User1"))
}

func (s *TodoHandlerSuite) TestCreateTodo() {
	user := CreateUser(s)

	reqBody := strings.NewReader(`{"Title": "User2", "Description": "user2@example.com", "UserId": ` + strconv.Itoa(user.ID) + `}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	// Set temporary header for testing
	req.Header.Set("X-User-ID", strconv.Itoa(user.ID))

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusAccepted))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := TodoResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Title).To(Equal("User2"))
	Expect(data.UUID).To(Not(BeEmpty()))
}

func (s *TodoHandlerSuite) TestDeleteByUUIDWhenIdExists() {
	user := CreateUser(s)

	todo, _ := s.setup.Repo.Create(factories.NewTodo[Todo](map[string]any{
		"Title":  "User",
		"UserId": user.ID,
	}))

	path := fmt.Sprintf("/todos/%s", todo.UUID.String())
	req, _ := http.NewRequest("DELETE", path, nil)
	rr := httptest.NewRecorder()

	req.Header.Set("X-User-ID", strconv.Itoa(user.ID))
	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := map[string]any{}
	json.Unmarshal(body, &data)

	Expect(data["message"]).To(Equal("Todo deleted successfully"))
}

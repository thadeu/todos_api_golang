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
	. "todoapp/internal/shared"
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
		"Name":              "User99",
		"Email":             "user99@example.com",
		"EncryptedPassword": "12345678",
	}))

	return user
}

func CreateTodo(s *TodoHandlerSuite, userId int) Todo {
	data, _ := s.setup.Repo.Create(factories.NewTodo[Todo](map[string]any{
		"Title":  "Task Created",
		"UserId": userId,
	}))

	return data
}

func (s *TodoHandlerSuite) TestGetAllTodosWithData() {
	user := CreateUser(s)

	s.setup.Repo.Create(factories.NewTodo[Todo](map[string]any{
		"Title":  "99",
		"Status": int(TodoStatusPending),
		"UserId": user.ID,
	}))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todos", nil)

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := GetAllTodosResponse{}
	json.Unmarshal(body, &data)

	Expect(len(data.Data)).To(Equal(1))
	Expect(data.Size).To(Equal(1))

	first := data.Data[0]
	Expect(first.Title).To(Equal("99"))
}

func (s *TodoHandlerSuite) TestCreateTodo() {
	user := CreateUser(s)

	reqBody := strings.NewReader(`{"title": "User2", "description": "user2@example.com", "status": "pending", "completed": false}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	// Set temporary header for testing
	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusAccepted))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := TodoResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Title).To(Equal("User2"))
	Expect(data.UUID).To(Not(BeEmpty()))
}

func (s *TodoHandlerSuite) TestUpdateTodoToCompleted() {
	user := CreateUser(s)
	todo := CreateTodo(s, user.ID)

	reqBody := strings.NewReader(`{
		"title": "Task Updated",
		"status": "completed",
		"completed": true
	}`)

	path := fmt.Sprintf("/todo/%s", todo.UUID.String())
	req, _ := http.NewRequest("PUT", path, reqBody)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := TodoResponse{}
	json.Unmarshal(body, &data)

	Expect(data.UUID).To(Not(BeEmpty()))
	Expect(data.Title).To(Equal("Task Updated"))
	Expect(data.Completed).To(BeTrue())
	Expect(data.Status).To(Equal("completed"))
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

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := map[string]any{}
	json.Unmarshal(body, &data)

	Expect(data["message"]).To(Equal("Todo deleted successfully"))
}

func (s *TodoHandlerSuite) TestCreateTodoWithDifferentStatuses() {
	user := CreateUser(s)

	// Test with in_review status
	reqBody := strings.NewReader(`{"title": "Review Task", "description": "Task in review", "status": "in_review", "completed": false}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusAccepted))
	Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := TodoResponse{}
	json.Unmarshal(body, &data)

	Expect(data.Title).To(Equal("Review Task"))
	Expect(data.Status).To(Equal("in_review"))
	Expect(data.UUID).To(Not(BeEmpty()))
}

func (s *TodoHandlerSuite) TestCreateTodoWithInvalidStatus() {
	user := CreateUser(s)

	// Test with invalid status
	reqBody := strings.NewReader(`{"title": "Invalid Task", "description": "Task with invalid status", "status": "invalid_status", "completed": false}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	http.DefaultServeMux.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusInternalServerError))
}

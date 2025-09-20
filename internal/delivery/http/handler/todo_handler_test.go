package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"todoapp/internal/delivery/http/middleware"
	"todoapp/internal/domain/entities"
	"todoapp/internal/domain/repositories"
	"todoapp/internal/infrastructure/persistence"
	"todoapp/internal/usecase/impl"
	. "todoapp/pkg/auth"
	. "todoapp/pkg/test"
	factory "todoapp/pkg/test/factory"

	c "todoapp/pkg/db/cursor"
)

type TodoHandlerSuite struct {
	suite.Suite
	UserRepo repositories.UserRepository
	TodoRepo repositories.TodoRepository
	Router   *gin.Engine
	DB       *sql.DB
}

var globalTodoHandler *TodoHandler
var ctx = context.Background()

func (s *TodoHandlerSuite) SetupSuite() {
	globalTodoHandler = &TodoHandler{}
}

func (s *TodoHandlerSuite) SetupTest() {
	s.DB = InitTestDB()
	s.TodoRepo = persistence.NewTodoRepository(s.DB)
	s.UserRepo = persistence.NewUserRepository(s.DB)

	// Create use case and handler
	todoUseCase := impl.NewTodoUseCase(s.TodoRepo)
	globalTodoHandler = NewTodoHandler(todoUseCase, nil)

	// Setup router directly to avoid import cycle
	s.Router = setupTodoTestRouter(globalTodoHandler)
}

func (s *TodoHandlerSuite) TearDownTest() {
	if s.DB != nil {
		s.DB.Close()
	}
}

func TestTodoHandlerSuite(t *testing.T) {
	RegisterTestingT(t)
	suite.Run(t, new(TodoHandlerSuite))
}

func setupTodoTestRouter(todoHandler *TodoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Protected routes
	protected := router.Group("/")
	protected.Use(middleware.CurrentMiddleware())
	protected.Use(middleware.GinJwtMiddleware())
	{
		protected.GET("/todos", todoHandler.GetAllTodos)
		protected.POST("/todos", todoHandler.CreateTodo)
		protected.PUT("/todo/:uuid", todoHandler.UpdateTodo)
		protected.DELETE("/todos/:uuid", todoHandler.DeleteByUUID)
	}

	return router
}

func CreateUser(s *TodoHandlerSuite) entities.User {
	user, _ := s.UserRepo.CreateUser(ctx, factory.NewUser[entities.User](map[string]any{
		"Name":              "User99",
		"Email":             "user99@example.com",
		"EncryptedPassword": "12345678",
	}))

	return user
}

func CreateTodo(s *TodoHandlerSuite, userId int) entities.Todo {
	data, _ := s.TodoRepo.Create(ctx, factory.NewTodo[entities.Todo](map[string]any{
		"Title":  "Task Created",
		"UserId": userId,
	}))

	return data
}

func (s *TodoHandlerSuite) TestGetAllTodosWithData() {
	user := CreateUser(s)

	s.TodoRepo.Create(ctx, factory.NewTodo[entities.Todo](map[string]any{
		"Title":  "99",
		"Status": int(entities.TodoStatusPending),
		"UserId": user.ID,
	}))

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todos", nil)

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := c.CursorResponse{}
	json.Unmarshal(body, &data)

	var todos []TodoResponse
	json.Unmarshal(data.Data, &todos)

	Expect(len(todos)).To(Equal(1))
	Expect(data.Size).To(Equal(1))

	first := todos[0]
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

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusCreated))
	Expect(rr.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

	body, _ := io.ReadAll(rr.Body)

	response := struct {
		Data TodoResponse `json:"data"`
	}{}
	json.Unmarshal(body, &response)

	Expect(response.Data.Title).To(Equal("User2"))
	Expect(response.Data.UUID).To(Not(BeEmpty()))
}

func (s *TodoHandlerSuite) TestCreateTodoValidationError() {
	user := CreateUser(s)

	reqBody := strings.NewReader(`{"title": "ab", "description": "test description"}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusBadRequest))

	body, _ := io.ReadAll(rr.Body)

	errorResponse := struct {
		Error struct {
			Code   string `json:"code"`
			Errors []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"errors"`
		} `json:"error"`
	}{}
	json.Unmarshal(body, &errorResponse)

	Expect(errorResponse.Error.Code).To(Equal("VALIDATION_ERROR"))
	Expect(len(errorResponse.Error.Errors)).To(BeNumerically(">", 0))
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

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

	body, _ := io.ReadAll(rr.Body)

	response := struct {
		Data TodoResponse `json:"data"`
	}{}
	json.Unmarshal(body, &response)

	Expect(response.Data.UUID).To(Not(BeEmpty()))
	Expect(response.Data.Title).To(Equal("Task Updated"))
	Expect(response.Data.Completed).To(BeTrue())
	Expect(response.Data.Status).To(Equal("completed"))
}

func (s *TodoHandlerSuite) TestDeleteByUUIDWhenIdExists() {
	user := CreateUser(s)

	todo, _ := s.TodoRepo.Create(ctx, factory.NewTodo[entities.Todo](map[string]any{
		"Title":  "User",
		"UserId": user.ID,
	}))

	path := fmt.Sprintf("/todos/%s", todo.UUID.String())
	req, _ := http.NewRequest("DELETE", path, nil)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
	Expect(rr.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

	body, _ := io.ReadAll(rr.Body)

	data := gin.H{}
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

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusCreated))
	Expect(rr.Header().Get("Content-Type")).To(ContainSubstring("application/json"))

	body, _ := io.ReadAll(rr.Body)

	response := struct {
		Data TodoResponse `json:"data"`
	}{}
	json.Unmarshal(body, &response)

	Expect(response.Data.Title).To(Equal("Review Task"))
	Expect(response.Data.Status).To(Equal("in_review"))
	Expect(response.Data.UUID).To(Not(BeEmpty()))
}

func (s *TodoHandlerSuite) TestCreateTodoWithInvalidStatus() {
	user := CreateUser(s)

	// Test with invalid status
	reqBody := strings.NewReader(`{"title": "Invalid Task", "description": "Task with invalid status", "status": "invalid_status", "completed": false}`)

	req, _ := http.NewRequest("POST", "/todos", reqBody)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusBadRequest))
}

func (s *TodoHandlerSuite) TestDeleteTodoWithSuccess() {
	user := CreateUser(s)
	todo := CreateTodo(s, user.ID)

	path := fmt.Sprintf("/todos/%s", todo.UUID.String())
	req, _ := http.NewRequest("DELETE", path, nil)
	rr := httptest.NewRecorder()

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))
}

func (s *TodoHandlerSuite) TestPaginationWithCursor() {
	user := CreateUser(s)

	baseTime := time.Now()

	for i := 1; i <= 5; i++ {
		s.TodoRepo.Create(ctx, factory.NewTodo[entities.Todo](map[string]any{
			"Title":     fmt.Sprintf("Task %d", i),
			"Status":    int(entities.TodoStatusPending),
			"UserId":    user.ID,
			"CreatedAt": baseTime.Add(time.Duration(i) * time.Minute), // Task 5 is newest
		}))
	}

	// Test first page (limit=2)
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/todos?limit=2", nil)

	jwtToken, _ := CreateJwtTokenForUser(user.ID)
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr, req)

	Expect(rr.Code).To(Equal(http.StatusOK))

	body, _ := io.ReadAll(rr.Body)

	// Dados retornados diretamente
	data := c.CursorResponse{}
	json.Unmarshal(body, &data)

	var todos []TodoResponse
	json.Unmarshal(data.Data, &todos)

	// First page should have 2 items and hasNext=true
	Expect(len(todos)).To(Equal(2))
	Expect(data.Size).To(Equal(2))
	Expect(data.Pagination.HasNext).To(BeTrue())
	Expect(data.Pagination.NextCursor).ToNot(BeEmpty())

	// Test second page using the cursor
	rr2 := httptest.NewRecorder()
	encodedCursor := url.QueryEscape(data.Pagination.NextCursor)
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/todos?limit=2&cursor=%s", encodedCursor), nil)
	req2.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr2, req2)

	Expect(rr2.Code).To(Equal(http.StatusOK))

	body2, _ := io.ReadAll(rr2.Body)
	data2 := c.CursorResponse{}
	json.Unmarshal(body2, &data2)

	var todos2 []TodoResponse
	json.Unmarshal(data2.Data, &todos2)

	// Second page should have 2 items and hasNext=true
	Expect(len(todos2)).To(Equal(2))
	Expect(data2.Size).To(Equal(2))
	Expect(data2.Pagination.HasNext).To(BeTrue())
	Expect(data2.Pagination.NextCursor).ToNot(BeEmpty())

	// Verify the cursors are different
	Expect(data2.Pagination.NextCursor).ToNot(Equal(data.Pagination.NextCursor))

	// Test third page
	rr3 := httptest.NewRecorder()
	encodedCursor2 := url.QueryEscape(data2.Pagination.NextCursor)
	req3, _ := http.NewRequest("GET", fmt.Sprintf("/todos?limit=2&cursor=%s", encodedCursor2), nil)
	req3.Header.Set("Authorization", "Bearer "+jwtToken)

	s.Router.ServeHTTP(rr3, req3)

	Expect(rr3.Code).To(Equal(http.StatusOK))

	body3, _ := io.ReadAll(rr3.Body)
	data3 := c.CursorResponse{}
	json.Unmarshal(body3, &data3)

	var todos3 []TodoResponse
	json.Unmarshal(data3.Data, &todos3)

	// Third page should have 1 item and hasNext=false
	Expect(len(todos3)).To(Equal(1))
	Expect(data3.Size).To(Equal(1))
	Expect(data3.Pagination.HasNext).To(BeFalse())
	Expect(data3.Pagination.NextCursor).To(BeEmpty())

	// Verify all todos are different
	allTitles := []string{}
	for _, todo := range todos {
		allTitles = append(allTitles, todo.Title)
	}
	for _, todo := range todos2 {
		allTitles = append(allTitles, todo.Title)
	}
	for _, todo := range todos3 {
		allTitles = append(allTitles, todo.Title)
	}

	// Should have 5 unique titles
	Expect(len(allTitles)).To(Equal(5))
}

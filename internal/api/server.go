package api

import (
	"log/slog"
	"net/http"

	. "todoapp/internal"
	. "todoapp/internal/handlers"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
)

func StartServer() {
	db := InitDB()

	user := NewUserRepository(db)
	// userService := NewUserService(user)
	// userHandler := NewUserHandler(userService)
	// userHandler.Register()

	todo := NewTodoRepository(db)
	toService := NewTodoService(todo)
	toHandler := NewTodoHandler(toService)
	toHandler.Register()

	authService := NewAuthService(user)
	authHandler := NewAuthHandler(authService)
	authHandler.Register()

	port := GetServerPort()
	slog.Info("Server starting", "port", port)

	http.ListenAndServe(":"+port, nil)
}

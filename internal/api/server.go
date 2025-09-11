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

	userRepo := NewUserRepository(db)
	userService := NewUserService(userRepo)
	userHandler := NewUserHandler(userService)
	userHandler.Register()

	todoRepo := NewTodoRepository(db)
	toService := NewTodoService(todoRepo)
	toHandler := NewTodoHandler(toService)
	toHandler.Register()

	port := GetServerPort()
	slog.Info("Server starting", "port", port)

	http.ListenAndServe(":"+port, nil)
}

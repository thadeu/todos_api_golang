package api

import (
	"log/slog"
	"net/http"
	"time"

	. "todoapp/internal"
	. "todoapp/internal/handlers"
	. "todoapp/internal/repositories"
	. "todoapp/internal/services"
)

func StartServer() {
	db := InitDB()

	user := NewUserRepository(db)
	todo := NewTodoRepository(db)

	todoService := NewTodoService(todo)
	authService := NewAuthService(user)

	todoHandler := NewTodoHandler(todoService)
	authHandler := NewAuthHandler(authService)

	router := SetupRouter(HandlersConfig{
		AuthHandler: authHandler,
		TodoHandler: todoHandler,
	})

	port := GetServerPort()
	slog.Info("Server starting", "port", port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server failed to start", "error", err)
	}

}

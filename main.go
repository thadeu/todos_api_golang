package main

import (
	"log/slog"
	"net/http"
)

func main() {
	db := initDB()
	repo := NewRepository(db)
	service := NewService(repo)

	handler := NewHandler(service)
	handler.registerUser()

	slog.Info("Server starting", "port", "8080")
	http.ListenAndServe(":8080", nil)
}

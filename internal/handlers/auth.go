package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "todoapp/internal/services"
)

type AuthHandler struct {
	Service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

func (t *AuthHandler) Register() {
	http.HandleFunc("POST /signup", t.RegisterByEmailAndPassword)
	http.HandleFunc("POST /auth", t.AuthByEmailAndPassword)
}

func (t *AuthHandler) RegisterByEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	var params AuthRequest
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Invalid params", err.Error()}})
		return
	}

	user, err := t.Service.Registration(params.Email, params.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{err.Error()}})
		return
	}

	w.WriteHeader(http.StatusOK)
	message := fmt.Sprintf("User %s was created successfully", user.Email)
	json.NewEncoder(w).Encode(map[string]any{"message": message})
}

func (t *AuthHandler) AuthByEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	var params AuthRequest
	err := json.NewDecoder(r.Body).Decode(&params)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Invalid params", err.Error()}})
		return
	}

	user, err := t.Service.Authenticate(params.Email, params.Password)

	// Paranoid response
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Email or password invalid", err.Error()}})
		return
	}

	refreshToken, err := t.Service.GenerateRefreshToken(&user)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{"errors": []string{"Generate refresh token failed", err.Error()}})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"refresh_token": refreshToken})
}

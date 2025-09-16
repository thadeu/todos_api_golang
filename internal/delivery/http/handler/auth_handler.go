package handler

import (
	"fmt"
	"net/http"

	. "todoapp/pkg/http"
	. "todoapp/pkg/response"

	"github.com/gin-gonic/gin"

	"todoapp/internal/usecase/interfaces"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	authUseCase interfaces.AuthUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase interfaces.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (a *AuthHandler) RegisterByEmailAndPassword(c *gin.Context) {
	var params interfaces.AuthRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	user, err := a.authUseCase.Registration(c.Request.Context(), params.Email, params.Password)

	if err != nil {
		SendBadRequestError(c, "registration", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("User %s created successfully", user.Email),
	})
}

func (a *AuthHandler) AuthByEmailAndPassword(c *gin.Context) {
	var params interfaces.AuthRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	user, err := a.authUseCase.Authenticate(c.Request.Context(), params.Email, params.Password)

	if err != nil {
		SendUnauthorizedError(c, "Invalid email or password")
		return
	}

	refreshToken, err := a.authUseCase.GenerateRefreshToken(user)

	if err != nil {
		SendUnauthorizedError(c, "Failed to generate access token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": refreshToken,
	})
}

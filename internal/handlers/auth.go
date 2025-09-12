package handlers

import (
	"fmt"
	"net/http"

	. "todoapp/internal/services"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	Service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{Service: service}
}

func (t *AuthHandler) RegisterByEmailAndPassword(c *gin.Context) {
	var params AuthRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{"Invalid params", err.Error()},
		})

		return
	}

	user, err := t.Service.Registration(params.Email, params.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{"Invalid params", err.Error()},
		})

		return
	}

	message := fmt.Sprintf("User %s was created successfully", user.Email)

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (t *AuthHandler) AuthByEmailAndPassword(c *gin.Context) {
	var params AuthRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": []string{"Invalid params", err.Error()},
		})

		return
	}

	user, err := t.Service.Authenticate(params.Email, params.Password)

	// Paranoid response
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []string{"Email or password invalid", err.Error()},
		})

		return
	}

	refreshToken, err := t.Service.GenerateRefreshToken(&user)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errors": []string{"Generate refresh token failed", err.Error()},
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"refresh_token": refreshToken})
}

package handlers

import (
	"fmt"
	"net/http"

	. "todoapp/internal/services"
	"todoapp/internal/shared"

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
		shared.SendBadRequestError(c, "request", "Parâmetros inválidos na requisição")
		return
	}

	// Validar os parâmetros usando o validator
	if err := shared.Validator.Struct(params); err != nil {
		shared.SendValidationError(c, err)
		return
	}

	user, err := t.Service.Registration(c.Request.Context(), params.Email, params.Password)

	if err != nil {
		shared.SendBadRequestError(c, "registration", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Usuário %s foi criado com sucesso", user.Email),
	})
}

func (t *AuthHandler) AuthByEmailAndPassword(c *gin.Context) {
	var params AuthRequest

	if err := c.ShouldBindJSON(&params); err != nil {
		shared.SendBadRequestError(c, "request", "Parâmetros inválidos na requisição")
		return
	}

	// Validar os parâmetros usando o validator
	if err := shared.Validator.Struct(params); err != nil {
		shared.SendValidationError(c, err)
		return
	}

	user, err := t.Service.Authenticate(c.Request.Context(), params.Email, params.Password)

	// Paranoid response
	if err != nil {
		shared.SendUnauthorizedError(c, "Email ou senha inválidos")
		return
	}

	refreshToken, err := t.Service.GenerateRefreshToken(&user)

	if err != nil {
		shared.SendUnauthorizedError(c, "Falha ao gerar token de acesso")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": refreshToken,
	})
}

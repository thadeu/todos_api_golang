package handler

import (
	"log/slog"
	"net/http"

	"todos/internal/adapter/http/helper"
	. "todos/internal/adapter/http/helper"
	. "todos/internal/adapter/http/validation"
	"todos/internal/core/model/request"
	"todos/internal/core/model/response"
	"todos/internal/core/port"
	"todos/internal/core/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc port.AuthService
}

func NewAuthHandler(svc port.AuthService) *AuthHandler {
	return &AuthHandler{
		svc: svc,
	}
}

func (a *AuthHandler) RegisterByEmailAndPassword(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := util.ParamsToMap[request.SignUpRequest](c)

	if err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	user, err := a.svc.Registration(ctx, &params)

	if err != nil {
		slog.Error("error", "error", err)
		SendBadRequestError(c, "registration", err.Error())
		return
	}

	userResponse := response.UserResponse{
		UUID:      user.UUID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	SendSuccess(c, http.StatusCreated, userResponse)
}

func (a *AuthHandler) AuthByEmailAndPassword(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := util.ParamsToMap[request.LoginRequest](c)

	if err != nil {
		SendBadRequestError(c, "request", "Invalid request parameters")
		return
	}

	if err := Validator.Struct(params); err != nil {
		SendValidationError(c, err)
		return
	}

	user, err := a.svc.Authenticate(ctx, &params)

	if err != nil {
		slog.Error("AuthByEmailAndPassword", "after_authenticate", err)
		SendUnauthorizedError(c, "Invalid email or password")
		return
	}

	refreshToken, err := helper.CreateJwtTokenForUser(user.ID)

	if err != nil {
		SendUnauthorizedError(c, "Invalid email or password")
		return
	}

	if err != nil {
		SendUnauthorizedError(c, "Failed to generate access token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": refreshToken,
	})
}
